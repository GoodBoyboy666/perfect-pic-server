package utils

import (
	"bytes"
	"io"
	"testing"
)

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantOK   bool
	}{
		{name: "too_short", username: "abc", wantOK: false},
		{name: "too_long", username: "aaaaaaaaaaaaaaaaaaaaa", wantOK: false}, // 21
		{name: "invalid_charset", username: "ab-cd", wantOK: false},
		{name: "reserved_admin", username: "admin", wantOK: false},
		{name: "reserved_case_insensitive", username: "RoOt", wantOK: false},
		{name: "pure_number", username: "123456", wantOK: false},
		{name: "valid", username: "user_123", wantOK: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, _ := ValidateUsername(tt.username)
			if ok != tt.wantOK {
				t.Fatalf("ValidateUsername(%q) ok=%v want=%v", tt.username, ok, tt.wantOK)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantOK   bool
	}{
		{name: "too_short", password: "a1b2c3", wantOK: false},
		{name: "no_number", password: "abcdefgh", wantOK: false},
		{name: "no_letter", password: "12345678", wantOK: false},
		{name: "non_ascii", password: "abc12345你好", wantOK: false},
		{name: "valid_simple", password: "abc12345", wantOK: true},
		{name: "valid_with_punct", password: "Abc12345!@", wantOK: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, _ := ValidatePassword(tt.password)
			if ok != tt.wantOK {
				t.Fatalf("ValidatePassword(%q) ok=%v want=%v", tt.password, ok, tt.wantOK)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name   string
		email  string
		wantOK bool
	}{
		{name: "empty", email: "", wantOK: false},
		{name: "missing_at", email: "a.example.com", wantOK: false},
		{name: "missing_tld", email: "a@b", wantOK: false},
		{name: "valid", email: "a.b+tag@example.com", wantOK: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, _ := ValidateEmail(tt.email)
			if ok != tt.wantOK {
				t.Fatalf("ValidateEmail(%q) ok=%v want=%v", tt.email, ok, tt.wantOK)
			}
		})
	}
}

func TestValidateImageContent(t *testing.T) {
	pngBytes := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // signature
		0x00, 0x00, 0x00, 0x0D, // IHDR length
		0x49, 0x48, 0x44, 0x52, // IHDR
		0x00, 0x00, 0x00, 0x01, // width=1
		0x00, 0x00, 0x00, 0x01, // height=1
		0x08, 0x02, 0x00, 0x00, 0x00, // bit depth/color type/etc
	}

	tests := []struct {
		name   string
		data   []byte
		ext    string
		wantOK bool
	}{
		{name: "png_ok", data: pngBytes, ext: ".png", wantOK: true},
		{name: "png_mismatch_ext", data: pngBytes, ext: ".jpg", wantOK: false},
		{name: "unsupported", data: []byte("not an image"), ext: ".png", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.data)
			ok, _ := ValidateImageContent(r, tt.ext)
			if ok != tt.wantOK {
				t.Fatalf("ValidateImageContent ok=%v want=%v", ok, tt.wantOK)
			}

			// Ensure ValidateImageContent resets the reader position on success or failure.
			if _, err := r.Read(make([]byte, 1)); err != nil && err != io.EOF {
				t.Fatalf("reader should still be readable after ValidateImageContent: %v", err)
			}
		})
	}
}
