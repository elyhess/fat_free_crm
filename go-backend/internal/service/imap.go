package service

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/mail"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

// IMAPConfig holds IMAP connection settings.
type IMAPConfig struct {
	Server              string
	Port                string
	User                string
	Password            string
	SSL                 bool
	Address             string // the dropbox email address
	ScanFolder          string // e.g. "INBOX"
	MoveToFolder        string // optional: archive processed emails
	MoveInvalidToFolder string // optional: move invalid emails
	AttachToAccount     bool   // also attach to contact's account
}

// DropboxProcessor processes incoming emails from an IMAP mailbox
// and attaches them to CRM entities.
type DropboxProcessor struct {
	db       *gorm.DB
	cfg      IMAPConfig
	emailSvc *EmailService
	baseURL  string
}

func NewDropboxProcessor(db *gorm.DB, cfg IMAPConfig, emailSvc *EmailService, baseURL string) *DropboxProcessor {
	return &DropboxProcessor{db: db, cfg: cfg, emailSvc: emailSvc, baseURL: baseURL}
}

// Process connects to IMAP, fetches unread messages, and processes them.
func (d *DropboxProcessor) Process() error {
	if d.cfg.Server == "" || d.cfg.User == "" {
		return fmt.Errorf("IMAP not configured")
	}

	addr := fmt.Sprintf("%s:%s", d.cfg.Server, d.cfg.Port)

	var client *imapclient.Client
	var err error

	if d.cfg.SSL {
		client, err = imapclient.DialTLS(addr, &imapclient.Options{
			TLSConfig: &tls.Config{ServerName: d.cfg.Server},
		})
	} else {
		client, err = imapclient.DialInsecure(addr, nil)
	}
	if err != nil {
		return fmt.Errorf("IMAP connect: %w", err)
	}
	defer client.Close()

	if err := client.Login(d.cfg.User, d.cfg.Password).Wait(); err != nil {
		return fmt.Errorf("IMAP login: %w", err)
	}

	folder := d.cfg.ScanFolder
	if folder == "" {
		folder = "INBOX"
	}

	if _, err := client.Select(folder, nil).Wait(); err != nil {
		return fmt.Errorf("IMAP select %s: %w", folder, err)
	}

	// Search for unseen messages
	criteria := &imap.SearchCriteria{NotFlag: []imap.Flag{imap.FlagSeen}}
	searchData, err := client.Search(criteria, nil).Wait()
	if err != nil {
		return fmt.Errorf("IMAP search: %w", err)
	}

	seqNums := searchData.AllSeqNums()
	if len(seqNums) == 0 {
		slog.Info("dropbox: no new messages")
		return nil
	}

	seqSet := imap.SeqSetNum(seqNums...)
	fetchOptions := &imap.FetchOptions{
		Envelope:    true,
		BodySection: []*imap.FetchItemBodySection{{}},
	}

	messages, err := client.Fetch(seqSet, fetchOptions).Collect()
	if err != nil {
		return fmt.Errorf("IMAP fetch: %w", err)
	}

	for _, msg := range messages {
		if err := d.processMessage(msg); err != nil {
			slog.Error("dropbox: process message", "error", err)
		}
	}

	return nil
}

func (d *DropboxProcessor) processMessage(msg *imapclient.FetchMessageBuffer) error {
	env := msg.Envelope
	if env == nil {
		return fmt.Errorf("no envelope")
	}

	// Get the raw body
	var bodyContent string
	for _, buf := range msg.BodySection {
		raw := string(buf.Bytes)
		parsed, err := mail.ReadMessage(strings.NewReader(raw))
		if err != nil {
			bodyContent = raw
		} else {
			_ = parsed
			bodyContent = raw
		}
		break
	}

	imapMsgID := env.MessageID
	if imapMsgID == "" {
		imapMsgID = fmt.Sprintf("gen-%d", time.Now().UnixNano())
	}

	// Check for duplicate
	var count int64
	d.db.Model(&model.Email{}).Where("imap_message_id = ?", imapMsgID).Count(&count)
	if count > 0 {
		slog.Info("dropbox: skipping duplicate", "message_id", imapMsgID)
		return nil
	}

	// Get sender
	sentFrom := ""
	if len(env.From) > 0 {
		sentFrom = env.From[0].Addr()
	}

	sentTo := ""
	if len(env.To) > 0 {
		addrs := make([]string, len(env.To))
		for i, a := range env.To {
			addrs[i] = a.Addr()
		}
		sentTo = strings.Join(addrs, ", ")
	}

	cc := ""
	if len(env.Cc) > 0 {
		addrs := make([]string, len(env.Cc))
		for i, a := range env.Cc {
			addrs[i] = a.Addr()
		}
		cc = strings.Join(addrs, ", ")
	}

	// Find sender user
	var user model.User
	if err := d.db.Where("email = ? AND deleted_at IS NULL AND suspended_at IS NULL", sentFrom).First(&user).Error; err != nil {
		slog.Warn("dropbox: unknown sender", "from", sentFrom)
		return nil // skip emails from unknown senders
	}

	// Try to match entity
	mediatorType, mediatorID := d.matchEntity(bodyContent, sentTo, cc)

	now := time.Now().UTC()
	var sentAt *time.Time
	if !env.Date.IsZero() {
		t := env.Date.UTC()
		sentAt = &t
	}

	email := model.Email{
		ImapMessageID: imapMsgID,
		UserID:        &user.ID,
		MediatorType:  mediatorType,
		MediatorID:    mediatorID,
		SentFrom:      sentFrom,
		SentTo:        sentTo,
		CC:            cc,
		Subject:       env.Subject,
		Body:          bodyContent,
		SentAt:        sentAt,
		ReceivedAt:    &now,
		CreatedAt:     &now,
		UpdatedAt:     &now,
		State:         "Expanded",
	}

	if err := d.db.Create(&email).Error; err != nil {
		return fmt.Errorf("create email record: %w", err)
	}

	slog.Info("dropbox: email processed",
		"from", sentFrom, "subject", env.Subject,
		"mediator", fmt.Sprintf("%s#%d", mediatorType, mediatorID))

	return nil
}

// matchEntity tries to find a CRM entity to attach the email to.
func (d *DropboxProcessor) matchEntity(body, sentTo, cc string) (string, int64) {
	// Strategy 1: Check first line for explicit keyword
	firstLine := strings.TrimSpace(strings.SplitN(body, "\n", 2)[0])
	firstLineLower := strings.ToLower(firstLine)

	keywords := map[string]string{
		"account":     "Account",
		"campaign":    "Campaign",
		"contact":     "Contact",
		"lead":        "Lead",
		"opportunity": "Opportunity",
	}

	for keyword, modelType := range keywords {
		if strings.HasPrefix(firstLineLower, keyword+":") {
			name := strings.TrimSpace(firstLine[len(keyword)+1:])
			if id := d.findEntityByName(modelType, name); id > 0 {
				return modelType, id
			}
		}
	}

	// Strategy 2: Check recipients for matching entity email
	allRecipients := sentTo + ", " + cc
	for _, addr := range strings.Split(allRecipients, ",") {
		addr = strings.TrimSpace(addr)
		if addr == "" || strings.EqualFold(addr, d.cfg.Address) {
			continue
		}
		if mType, mID := d.findEntityByEmail(addr); mID > 0 {
			return mType, mID
		}
	}

	return "", 0
}

func (d *DropboxProcessor) findEntityByName(modelType, name string) int64 {
	table := strings.ToLower(modelType) + "s"
	if modelType == "Contact" || modelType == "Lead" {
		parts := strings.SplitN(name, " ", 2)
		if len(parts) == 2 {
			var id int64
			d.db.Table(table).Where("first_name ILIKE ? AND last_name ILIKE ? AND deleted_at IS NULL",
				parts[0], parts[1]).Select("id").Scan(&id)
			if id > 0 {
				return id
			}
		}
		return 0
	}
	var id int64
	d.db.Table(table).Where("name ILIKE ? AND deleted_at IS NULL", name).Select("id").Scan(&id)
	return id
}

func (d *DropboxProcessor) findEntityByEmail(email string) (string, int64) {
	var id int64
	d.db.Table("contacts").Where("email ILIKE ? AND deleted_at IS NULL", email).Select("id").Scan(&id)
	if id > 0 {
		return "Contact", id
	}
	d.db.Table("leads").Where("email ILIKE ? AND deleted_at IS NULL", email).Select("id").Scan(&id)
	if id > 0 {
		return "Lead", id
	}
	d.db.Table("accounts").Where("email ILIKE ? AND deleted_at IS NULL", email).Select("id").Scan(&id)
	if id > 0 {
		return "Account", id
	}
	return "", 0
}
