package auth

import (
	"testing"
)

func TestDigestPassword_MatchesRails(t *testing.T) {
	// Known values from Rails:
	// password: "Dem0P@ssword!!"
	// salt: "8KXTUTk2yZPqHvsed4NH"
	// stretches: 20
	// expected hash: 01f6c04b6b18f5fb8abac910ef65c94483c80bab0ccdb82a376961a526248e43958cded817fb5604297248c4d0478021a283d170d75f54641d9b59875db0e239
	password := "Dem0P@ssword!!"
	salt := "8KXTUTk2yZPqHvsed4NH"
	expected := "01f6c04b6b18f5fb8abac910ef65c94483c80bab0ccdb82a376961a526248e43958cded817fb5604297248c4d0478021a283d170d75f54641d9b59875db0e239"

	result := DigestPassword(password, salt, 20)
	if result != expected {
		t.Errorf("hash mismatch\nexpected: %s\ngot:      %s", expected, result)
	}
}

func TestVerifyPassword_Valid(t *testing.T) {
	password := "Dem0P@ssword!!"
	salt := "8KXTUTk2yZPqHvsed4NH"
	encrypted := "01f6c04b6b18f5fb8abac910ef65c94483c80bab0ccdb82a376961a526248e43958cded817fb5604297248c4d0478021a283d170d75f54641d9b59875db0e239"

	if !VerifyPassword(password, encrypted, salt, 20) {
		t.Error("expected password to verify successfully")
	}
}

func TestVerifyPassword_Invalid(t *testing.T) {
	salt := "8KXTUTk2yZPqHvsed4NH"
	encrypted := "01f6c04b6b18f5fb8abac910ef65c94483c80bab0ccdb82a376961a526248e43958cded817fb5604297248c4d0478021a283d170d75f54641d9b59875db0e239"

	if VerifyPassword("wrong-password", encrypted, salt, 20) {
		t.Error("expected wrong password to fail verification")
	}
}

func TestDigestPassword_SingleStretch(t *testing.T) {
	// With 1 stretch (test mode), verify it's a single SHA-512
	result := DigestPassword("test", "salt", 1)
	if len(result) != 128 { // SHA-512 hex is 128 chars
		t.Errorf("expected 128 char hex string, got %d chars", len(result))
	}
}

func TestValidatePasswordComplexity_Valid(t *testing.T) {
	err := ValidatePasswordComplexity("Dem0P@ssword!!")
	if err.HasErrors() {
		t.Errorf("expected no errors for valid password, got: %v", err.Messages())
	}
}

func TestValidatePasswordComplexity_TooShort(t *testing.T) {
	err := ValidatePasswordComplexity("Sh0rt!")
	if !err.TooShort {
		t.Error("expected TooShort error")
	}
}

func TestValidatePasswordComplexity_NoDigit(t *testing.T) {
	err := ValidatePasswordComplexity("NoDigitsHere!!!!!")
	if !err.NoDigit {
		t.Error("expected NoDigit error")
	}
}

func TestValidatePasswordComplexity_NoLower(t *testing.T) {
	err := ValidatePasswordComplexity("NOLOWERCASE1!!!")
	if !err.NoLower {
		t.Error("expected NoLower error")
	}
}

func TestValidatePasswordComplexity_NoUpper(t *testing.T) {
	err := ValidatePasswordComplexity("nouppercase1!!!")
	if !err.NoUpper {
		t.Error("expected NoUpper error")
	}
}

func TestValidatePasswordComplexity_NoSymbol(t *testing.T) {
	err := ValidatePasswordComplexity("NoSymbolHere1234")
	if !err.NoSymbol {
		t.Error("expected NoSymbol error")
	}
}

func TestValidatePasswordComplexity_TooLong(t *testing.T) {
	long := make([]byte, 129)
	for i := range long {
		long[i] = 'A'
	}
	long[0] = 'a'
	long[1] = '1'
	long[2] = '!'
	err := ValidatePasswordComplexity(string(long))
	if !err.TooLong {
		t.Error("expected TooLong error")
	}
}

func TestPasswordComplexityError_Messages(t *testing.T) {
	err := ValidatePasswordComplexity("short")
	msgs := err.Messages()
	if len(msgs) == 0 {
		t.Error("expected at least one message")
	}
}
