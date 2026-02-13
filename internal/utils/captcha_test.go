package utils

import "testing"

func TestCaptcha_GenerateAndVerify(t *testing.T) {
	id, b64, answer, err := MakeCaptcha()
	if err != nil {
		t.Fatalf("MakeCaptcha error: %v", err)
	}
	if id == "" || b64 == "" || answer == "" {
		t.Fatalf("expected non-empty captcha fields, id=%q b64=%q answer=%q", id, b64, answer)
	}

	if ok := VerifyCaptcha(id, answer); !ok {
		t.Fatalf("expected captcha verify to succeed for correct answer")
	}

	// Captcha is one-time due to clear=true in VerifyCaptcha wrapper.
	if ok := VerifyCaptcha(id, answer); ok {
		t.Fatalf("expected captcha verify to fail after being cleared")
	}

	id2, _, _, err := MakeCaptcha()
	if err != nil {
		t.Fatalf("MakeCaptcha(2) error: %v", err)
	}
	if ok := VerifyCaptcha(id2, "wrong"); ok {
		t.Fatalf("expected captcha verify to fail for wrong answer")
	}
}
