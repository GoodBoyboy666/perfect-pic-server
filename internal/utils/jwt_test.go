package utils

import (
	"testing"
	"time"
)

func TestLoginToken_RoundTrip(t *testing.T) {
	token, err := GenerateLoginToken(123, "alice", true, time.Hour)
	if err != nil {
		t.Fatalf("GenerateLoginToken error: %v", err)
	}
	claims, err := ParseLoginToken(token)
	if err != nil {
		t.Fatalf("ParseLoginToken error: %v", err)
	}
	if claims.ID != 123 || claims.Username != "alice" || claims.Admin != true || claims.Type != "login" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestParseLoginToken_RejectsWrongType(t *testing.T) {
	emailToken, err := GenerateEmailToken(1, "a@example.com", time.Hour)
	if err != nil {
		t.Fatalf("GenerateEmailToken error: %v", err)
	}
	_, err = ParseLoginToken(emailToken)
	if err == nil {
		t.Fatalf("expected error for wrong token type")
	}
}

func TestParseLoginToken_Expired(t *testing.T) {
	token, err := GenerateLoginToken(1, "alice", false, -1*time.Second)
	if err != nil {
		t.Fatalf("GenerateLoginToken error: %v", err)
	}
	_, err = ParseLoginToken(token)
	if err == nil {
		t.Fatalf("expected expired token error")
	}
}

func TestEmailToken_RoundTrip(t *testing.T) {
	token, err := GenerateEmailToken(1, "a@example.com", time.Hour)
	if err != nil {
		t.Fatalf("GenerateEmailToken: %v", err)
	}
	claims, err := ParseEmailToken(token)
	if err != nil {
		t.Fatalf("ParseEmailToken: %v", err)
	}
	if claims.Email != "a@example.com" || claims.Type != "email_verify" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestEmailChangeToken_RoundTrip(t *testing.T) {
	token, err := GenerateEmailChangeToken(1, "old@example.com", "new@example.com", time.Hour)
	if err != nil {
		t.Fatalf("GenerateEmailChangeToken: %v", err)
	}
	claims, err := ParseEmailChangeToken(token)
	if err != nil {
		t.Fatalf("ParseEmailChangeToken: %v", err)
	}
	if claims.OldEmail != "old@example.com" || claims.NewEmail != "new@example.com" || claims.Type != "email_change" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}
