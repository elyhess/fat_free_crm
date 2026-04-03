package service

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

// entityShortcuts maps short codes to entity types (matching Rails).
var entityShortcuts = map[string]string{
	"ac": "Account",
	"ca": "Campaign",
	"co": "Contact",
	"le": "Lead",
	"op": "Opportunity",
	"ta": "Task",
}

// entityFullNames maps full names to Rails model types.
var entityFullNames = map[string]string{
	"account":     "Account",
	"campaign":    "Campaign",
	"contact":     "Contact",
	"lead":        "Lead",
	"opportunity": "Opportunity",
	"task":        "Task",
}

// subjectPattern matches [entity_type:id] or [shortcut:id] in email subjects.
var subjectPattern = regexp.MustCompile(`\[(\w+):(\d+)\]`)

// CommentReplyProcessor processes reply emails and creates comments.
type CommentReplyProcessor struct {
	db  *gorm.DB
	cfg IMAPConfig
}

func NewCommentReplyProcessor(db *gorm.DB, cfg IMAPConfig) *CommentReplyProcessor {
	return &CommentReplyProcessor{db: db, cfg: cfg}
}

// Process connects to IMAP, fetches unread messages, and creates comments.
func (p *CommentReplyProcessor) Process() error {
	if p.cfg.Server == "" || p.cfg.User == "" {
		return fmt.Errorf("IMAP not configured for comment replies")
	}

	addr := fmt.Sprintf("%s:%s", p.cfg.Server, p.cfg.Port)

	var client *imapclient.Client
	var err error

	if p.cfg.SSL {
		client, err = imapclient.DialTLS(addr, &imapclient.Options{
			TLSConfig: &tls.Config{ServerName: p.cfg.Server},
		})
	} else {
		client, err = imapclient.DialInsecure(addr, nil)
	}
	if err != nil {
		return fmt.Errorf("IMAP connect: %w", err)
	}
	defer client.Close()

	if err := client.Login(p.cfg.User, p.cfg.Password).Wait(); err != nil {
		return fmt.Errorf("IMAP login: %w", err)
	}

	folder := p.cfg.ScanFolder
	if folder == "" {
		folder = "INBOX"
	}

	if _, err := client.Select(folder, nil).Wait(); err != nil {
		return fmt.Errorf("IMAP select %s: %w", folder, err)
	}

	criteria := &imap.SearchCriteria{NotFlag: []imap.Flag{imap.FlagSeen}}
	searchData, err := client.Search(criteria, nil).Wait()
	if err != nil {
		return fmt.Errorf("IMAP search: %w", err)
	}

	seqNums := searchData.AllSeqNums()
	if len(seqNums) == 0 {
		slog.Info("comment_replies: no new messages")
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
		if err := p.processReply(msg); err != nil {
			slog.Error("comment_replies: process message", "error", err)
		}
	}

	return nil
}

func (p *CommentReplyProcessor) processReply(msg *imapclient.FetchMessageBuffer) error {
	env := msg.Envelope
	if env == nil {
		return fmt.Errorf("no envelope")
	}

	// Parse entity reference from subject
	matches := subjectPattern.FindStringSubmatch(env.Subject)
	if matches == nil {
		slog.Info("comment_replies: no entity reference in subject", "subject", env.Subject)
		return nil
	}

	entityKey := strings.ToLower(matches[1])
	entityID, _ := strconv.ParseInt(matches[2], 10, 64)
	if entityID == 0 {
		return nil
	}

	// Resolve entity type
	entityType := entityFullNames[entityKey]
	if entityType == "" {
		entityType = entityShortcuts[entityKey]
	}
	if entityType == "" {
		slog.Warn("comment_replies: unknown entity type", "key", entityKey)
		return nil
	}

	// Verify entity exists
	tableName := strings.ToLower(entityType) + "s"
	var count int64
	p.db.Table(tableName).Where("id = ? AND deleted_at IS NULL", entityID).Count(&count)
	if count == 0 {
		slog.Warn("comment_replies: entity not found", "type", entityType, "id", entityID)
		return nil
	}

	// Get sender
	sentFrom := ""
	if len(env.From) > 0 {
		sentFrom = env.From[0].Addr()
	}

	// Find sender user
	var user model.User
	if err := p.db.Where("email = ? AND deleted_at IS NULL AND suspended_at IS NULL", sentFrom).First(&user).Error; err != nil {
		slog.Warn("comment_replies: unknown sender", "from", sentFrom)
		return nil
	}

	// Extract reply body from raw content
	var bodyContent string
	for _, buf := range msg.BodySection {
		parsed, err := mail.ReadMessage(strings.NewReader(string(buf.Bytes)))
		if err != nil {
			bodyContent = string(buf.Bytes)
		} else {
			b := make([]byte, 4096)
			n, _ := parsed.Body.Read(b)
			bodyContent = string(b[:n])
		}
		break
	}

	// Strip quoted content
	replyBody := extractReply(bodyContent)
	if strings.TrimSpace(replyBody) == "" {
		slog.Info("comment_replies: empty reply body", "from", sentFrom)
		return nil
	}

	// Create comment
	now := time.Now().UTC()
	comment := model.Comment{
		UserID:          user.ID,
		CommentableType: entityType,
		CommentableID:   entityID,
		Comment:         strings.TrimSpace(replyBody),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := p.db.Create(&comment).Error; err != nil {
		return fmt.Errorf("create comment: %w", err)
	}

	slog.Info("comment_replies: comment created",
		"from", sentFrom, "entity", fmt.Sprintf("%s#%d", entityType, entityID))

	return nil
}

// extractReply strips quoted content from an email reply.
func extractReply(body string) string {
	lines := strings.Split(body, "\n")
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, ">") {
			break
		}
		if strings.HasPrefix(trimmed, "On ") && strings.HasSuffix(trimmed, "wrote:") {
			break
		}
		if strings.HasPrefix(trimmed, "------") || strings.HasPrefix(trimmed, "______") {
			break
		}
		if strings.Contains(trimmed, "Original Message") {
			break
		}
		result = append(result, line)
	}

	return strings.Join(result, "\n")
}
