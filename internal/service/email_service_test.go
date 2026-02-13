package service

import (
	"strings"
	"testing"

	"perfect-pic-server/internal/config"
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/db"
	"perfect-pic-server/internal/model"
)

func TestRenderTemplate(t *testing.T) {
	out, err := renderTemplate("hi {{.Name}}", map[string]string{"Name": "alice"})
	if err != nil {
		t.Fatalf("renderTemplate: %v", err)
	}
	if out != "hi alice" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestFormatAddressHeader(t *testing.T) {
	header, addr, err := formatAddressHeader("Alice <alice@example.com>")
	if err != nil {
		t.Fatalf("formatAddressHeader: %v", err)
	}
	if addr != "alice@example.com" {
		t.Fatalf("unexpected addr: %q", addr)
	}
	if !strings.Contains(header, "<alice@example.com>") {
		t.Fatalf("unexpected header: %q", header)
	}
	if strings.ContainsAny(header, "\r\n") {
		t.Fatalf("header contains CRLF: %q", header)
	}

	header2, addr2, err := formatAddressHeader("bob@example.com")
	if err != nil {
		t.Fatalf("formatAddressHeader: %v", err)
	}
	if header2 != "bob@example.com" || addr2 != "bob@example.com" {
		t.Fatalf("unexpected header/addr: %q %q", header2, addr2)
	}

	_, _, err = formatAddressHeader("not-an-email")
	if err == nil {
		t.Fatalf("expected error for invalid address")
	}
}

func TestBuildEmailMessage(t *testing.T) {
	msg, err := buildEmailMessage("from@example.com", "to@example.com", "主题", "<p>hi</p>")
	if err != nil {
		t.Fatalf("buildEmailMessage: %v", err)
	}
	s := string(msg)
	if !strings.Contains(s, "Subject:") || !strings.Contains(s, "MIME-Version: 1.0") {
		t.Fatalf("missing headers in message: %q", s)
	}
	if !strings.Contains(s, "<p>hi</p>") {
		t.Fatalf("missing body in message")
	}
}

func TestSendTestEmail_MissingHost(t *testing.T) {
	// config.InitConfig in TestMain sets defaults; SMTP host default is empty.
	err := SendTestEmail("a@example.com")
	if err == nil {
		t.Fatalf("expected error when SMTP host is missing")
	}
}

func TestSendVerificationEmail_SMTPDisabledNoop(t *testing.T) {
	// Default settings enable_smtp=false, so this should be a no-op.
	if err := SendVerificationEmail("a@example.com", "alice", "http://example/verify"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestSendEmailChangeVerification_SMTPDisabledNoop(t *testing.T) {
	if err := SendEmailChangeVerification("a@example.com", "alice", "old@example.com", "new@example.com", "http://example/verify"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestSendPasswordResetEmail_SMTPDisabledNoop(t *testing.T) {
	if err := SendPasswordResetEmail("a@example.com", "alice", "http://example/reset"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestSendMailWithSSL_DialFailure(t *testing.T) {
	_, _ = buildEmailMessage("from@example.com", "to@example.com", "sub", "body")
	err := sendMailWithSSL("127.0.0.1:1", nil, "from@example.com", []string{"to@example.com"}, []byte("x"))
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestEmailSendFunctions_AttemptSendAndFailFast(t *testing.T) {
	setupTestDB(t)

	// Enable SMTP in settings.
	_ = db.DB.Save(&model.Setting{Key: consts.ConfigEnableSMTP, Value: "true"}).Error
	ClearCache()

	// Re-init config with an unreachable SMTP host/port so SendMail fails quickly.
	cfgDir := t.TempDir()
	t.Setenv("PERFECT_PIC_SERVER_MODE", "debug")
	t.Setenv("PERFECT_PIC_JWT_SECRET", "test_secret")
	t.Setenv("PERFECT_PIC_SMTP_HOST", "127.0.0.1")
	t.Setenv("PERFECT_PIC_SMTP_PORT", "1")
	t.Setenv("PERFECT_PIC_SMTP_FROM", "Perfect Pic <from@example.com>")
	config.InitConfig(cfgDir)

	if err := SendVerificationEmail("to@example.com", "alice", "http://example/verify"); err == nil {
		t.Fatalf("expected send failure")
	}
	if err := SendEmailChangeVerification("to@example.com", "alice", "old@example.com", "new@example.com", "http://example/verify"); err == nil {
		t.Fatalf("expected send failure")
	}
	if err := SendPasswordResetEmail("to@example.com", "alice", "http://example/reset"); err == nil {
		t.Fatalf("expected send failure")
	}
	if err := SendTestEmail("to@example.com"); err == nil {
		t.Fatalf("expected send failure")
	}
}
