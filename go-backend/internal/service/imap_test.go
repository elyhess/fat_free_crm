package service

import (
	"testing"

	"gorm.io/gorm"
)

func seedImapTestUser(db *gorm.DB) {
	db.Exec(`INSERT INTO users (id, username, email, encrypted_password, admin, created_at, updated_at)
		VALUES (1, 'admin', 'admin@test.com', 'x', true, NOW(), NOW())
		ON CONFLICT (id) DO NOTHING`)
}

func TestMatchEntity_ByKeyword(t *testing.T) {
	db := testDB(t)
	seedImapTestUser(db)
	db.Exec(`INSERT INTO accounts (id, user_id, name, created_at, updated_at) VALUES (1, 1, 'Acme Corp', NOW(), NOW())`)

	proc := &DropboxProcessor{db: db, cfg: IMAPConfig{Address: "dropbox@test.com"}}

	mType, mID := proc.matchEntity("Account: Acme Corp\nRest of the body", "someone@test.com", "")
	if mType != "Account" || mID != 1 {
		t.Errorf("expected Account#1, got %s#%d", mType, mID)
	}
}

func TestMatchEntity_ByKeyword_Contact(t *testing.T) {
	db := testDB(t)
	seedImapTestUser(db)
	db.Exec(`INSERT INTO contacts (id, user_id, first_name, last_name, created_at, updated_at)
		VALUES (1, 1, 'John', 'Doe', NOW(), NOW())`)

	proc := &DropboxProcessor{db: db, cfg: IMAPConfig{Address: "dropbox@test.com"}}

	mType, mID := proc.matchEntity("Contact: John Doe\nbody here", "someone@test.com", "")
	if mType != "Contact" || mID != 1 {
		t.Errorf("expected Contact#1, got %s#%d", mType, mID)
	}
}

func TestMatchEntity_ByRecipientEmail(t *testing.T) {
	db := testDB(t)
	seedImapTestUser(db)
	db.Exec(`INSERT INTO contacts (id, user_id, first_name, last_name, email, created_at, updated_at)
		VALUES (1, 1, 'Jane', 'Smith', 'jane@example.com', NOW(), NOW())`)

	proc := &DropboxProcessor{db: db, cfg: IMAPConfig{Address: "dropbox@test.com"}}

	mType, mID := proc.matchEntity("No keyword here", "dropbox@test.com, jane@example.com", "")
	if mType != "Contact" || mID != 1 {
		t.Errorf("expected Contact#1, got %s#%d", mType, mID)
	}
}

func TestMatchEntity_SkipsDropboxAddress(t *testing.T) {
	db := testDB(t)
	proc := &DropboxProcessor{db: db, cfg: IMAPConfig{Address: "dropbox@test.com"}}

	mType, mID := proc.matchEntity("No keyword", "dropbox@test.com", "")
	if mType != "" || mID != 0 {
		t.Errorf("expected no match, got %s#%d", mType, mID)
	}
}

func TestMatchEntity_NoMatch(t *testing.T) {
	db := testDB(t)
	proc := &DropboxProcessor{db: db, cfg: IMAPConfig{Address: "dropbox@test.com"}}

	mType, mID := proc.matchEntity("Random body", "unknown@example.com", "")
	if mType != "" || mID != 0 {
		t.Errorf("expected no match, got %s#%d", mType, mID)
	}
}
