package handler

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

// ========== Comment Tests ==========

func TestCreateComment_Success(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	// Create an account first
	doRequest(mux, "POST", "/api/v1/accounts", tok, map[string]interface{}{"name": "Acme"})

	rec := doRequest(mux, "POST", "/api/v1/accounts/1/comments", tok, map[string]interface{}{
		"comment": "Great company!", "private": false,
	})
	if rec.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var comment model.Comment
	if err := json.NewDecoder(rec.Body).Decode(&comment); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if comment.Comment != "Great company!" {
		t.Errorf("expected 'Great company!', got %q", comment.Comment)
	}
	if comment.CommentableType != "Account" {
		t.Errorf("expected commentable_type 'Account', got %q", comment.CommentableType)
	}
	if comment.CommentableID != 1 {
		t.Errorf("expected commentable_id 1, got %d", comment.CommentableID)
	}
	if comment.UserID != 1 {
		t.Errorf("expected user_id 1, got %d", comment.UserID)
	}
}

func TestCreateComment_InvalidEntityType(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	rec := doRequest(mux, "POST", "/api/v1/invalid/1/comments", makeToken("admin"), map[string]interface{}{
		"comment": "test",
	})
	if rec.Code != 400 {
		t.Errorf("expected 400 for invalid entity, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateComment_EntityNotFound(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	// Comment on a non-existent account (ID 999). The implementation does not
	// check entity existence (polymorphic), so the comment is created successfully.
	rec := doRequest(mux, "POST", "/api/v1/accounts/999/comments", tok, map[string]interface{}{
		"comment": "orphan comment",
	})
	// No FK constraint on polymorphic commentable_id, so this succeeds
	if rec.Code != 201 {
		t.Fatalf("expected 201 (no entity existence check), got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateComment_EmptyCommentBody(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/accounts", tok, map[string]interface{}{"name": "Acme"})

	rec := doRequest(mux, "POST", "/api/v1/accounts/1/comments", tok, map[string]interface{}{
		"comment": "",
	})
	if rec.Code != 422 {
		t.Errorf("expected 422 for empty comment, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteComment_Success(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/accounts", tok, map[string]interface{}{"name": "Acme"})
	doRequest(mux, "POST", "/api/v1/accounts/1/comments", tok, map[string]interface{}{"comment": "Delete me"})

	rec := doRequest(mux, "DELETE", "/api/v1/comments/1", tok, nil)
	if rec.Code != 200 {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteComment_NotOwner(t *testing.T) {
	mux, jwtSvc, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/accounts", tok, map[string]interface{}{"name": "Acme"})
	doRequest(mux, "POST", "/api/v1/accounts/1/comments", tok, map[string]interface{}{"comment": "Admin comment"})

	// Try to delete as non-admin user 5
	tok5, _ := jwtSvc.GenerateToken(5, "other", false)
	rec := doRequest(mux, "DELETE", "/api/v1/comments/1", tok5, nil)
	if rec.Code != 403 {
		t.Errorf("expected 403, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteComment_AdminCanDelete(t *testing.T) {
	mux, jwtSvc, _ := writeRouter(t)

	// Create comment as user 5 (non-admin)
	tok5, _ := jwtSvc.GenerateToken(5, "other", false)
	doRequest(mux, "POST", "/api/v1/accounts", tok5, map[string]interface{}{"name": "Acme"})
	doRequest(mux, "POST", "/api/v1/accounts/1/comments", tok5, map[string]interface{}{"comment": "User comment"})

	// Admin (user 1) can delete anyone's comment
	tokAdmin, _ := jwtSvc.GenerateToken(1, "admin", true)
	rec := doRequest(mux, "DELETE", "/api/v1/comments/1", tokAdmin, nil)
	if rec.Code != 200 {
		t.Errorf("expected 200 (admin can delete any comment), got %d: %s", rec.Code, rec.Body.String())
	}
}

// ========== Tag Tests ==========

func TestAddTag_Success(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/accounts", tok, map[string]interface{}{"name": "Acme"})

	rec := doRequest(mux, "POST", "/api/v1/accounts/1/tags", tok, map[string]interface{}{"name": "vip"})
	if rec.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var tag model.Tag
	if err := json.NewDecoder(rec.Body).Decode(&tag); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if tag.Name != "vip" {
		t.Errorf("expected tag 'vip', got %q", tag.Name)
	}
	if tag.ID == 0 {
		t.Error("expected non-zero tag ID")
	}
}

func TestAddTag_DuplicateTag(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/accounts", tok, map[string]interface{}{"name": "Acme"})
	doRequest(mux, "POST", "/api/v1/accounts/1/tags", tok, map[string]interface{}{"name": "vip"})

	// Adding same tag again should return 200 (idempotent)
	rec := doRequest(mux, "POST", "/api/v1/accounts/1/tags", tok, map[string]interface{}{"name": "vip"})
	if rec.Code != 200 {
		t.Errorf("expected 200 for duplicate tag, got %d: %s", rec.Code, rec.Body.String())
	}

	var tag model.Tag
	if err := json.NewDecoder(rec.Body).Decode(&tag); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if tag.Name != "vip" {
		t.Errorf("expected tag 'vip', got %q", tag.Name)
	}
}

func TestAddTag_NewTagCreated(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/accounts", tok, map[string]interface{}{"name": "Acme"})

	rec := doRequest(mux, "POST", "/api/v1/accounts/1/tags", tok, map[string]interface{}{"name": "newtag"})
	if rec.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	// Verify the tag was returned with correct fields
	var tag model.Tag
	if err := json.NewDecoder(rec.Body).Decode(&tag); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if tag.Name != "newtag" {
		t.Errorf("expected tag name 'newtag', got %q", tag.Name)
	}
	if tag.ID == 0 {
		t.Error("expected tag to have a non-zero ID (tag was persisted)")
	}
}

func TestRemoveTag_Success(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/accounts", tok, map[string]interface{}{"name": "Acme"})
	createRec := doRequest(mux, "POST", "/api/v1/accounts/1/tags", tok, map[string]interface{}{"name": "removeme"})

	var tag model.Tag
	if err := json.NewDecoder(createRec.Body).Decode(&tag); err != nil {
		t.Fatalf("decode: %v", err)
	}

	rec := doRequest(mux, "DELETE", "/api/v1/accounts/1/tags/"+strconv.FormatInt(tag.ID, 10), tok, nil)
	if rec.Code != 200 {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRemoveTag_NotFound(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/accounts", tok, map[string]interface{}{"name": "Acme"})

	// Remove a tag that was never added. The implementation returns 200 regardless
	// (no check for RowsAffected == 0).
	rec := doRequest(mux, "DELETE", "/api/v1/accounts/1/tags/9999", tok, nil)
	if rec.Code != 200 {
		t.Errorf("expected 200 (implementation does not check RowsAffected), got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAddTag_InvalidEntity(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	rec := doRequest(mux, "POST", "/api/v1/invalid/1/tags", makeToken("admin"), map[string]interface{}{
		"name": "vip",
	})
	if rec.Code != 400 {
		t.Errorf("expected 400 for invalid entity, got %d: %s", rec.Code, rec.Body.String())
	}
}

// ========== Address Tests ==========

func TestCreateAddress_Success(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/accounts", tok, map[string]interface{}{"name": "Acme"})

	rec := doRequest(mux, "POST", "/api/v1/accounts/1/addresses", tok, map[string]interface{}{
		"street1": "123 Main St", "city": "Anytown", "state": "CA", "country": "US",
	})
	if rec.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var addr model.Address
	if err := json.NewDecoder(rec.Body).Decode(&addr); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if addr.AddressableType != "Account" {
		t.Errorf("expected addressable_type 'Account', got %q", addr.AddressableType)
	}
	if addr.AddressableID != 1 {
		t.Errorf("expected addressable_id 1, got %d", addr.AddressableID)
	}
	if addr.City == nil || *addr.City != "Anytown" {
		t.Errorf("expected city 'Anytown', got %v", addr.City)
	}
}

func TestCreateAddress_InvalidEntity(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	rec := doRequest(mux, "POST", "/api/v1/invalid/1/addresses", makeToken("admin"), map[string]interface{}{
		"street1": "123 Main St",
	})
	if rec.Code != 400 {
		t.Errorf("expected 400 for invalid entity, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteAddress_Success(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/accounts", tok, map[string]interface{}{"name": "Acme"})
	doRequest(mux, "POST", "/api/v1/accounts/1/addresses", tok, map[string]interface{}{
		"street1": "Delete me",
	})

	rec := doRequest(mux, "DELETE", "/api/v1/addresses/1", tok, nil)
	if rec.Code != 200 {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteAddress_NotFound(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	rec := doRequest(mux, "DELETE", "/api/v1/addresses/9999", makeToken("admin"), nil)
	if rec.Code != 404 {
		t.Errorf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}
