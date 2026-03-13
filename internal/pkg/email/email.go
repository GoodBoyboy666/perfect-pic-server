package email

import (
	"fmt"
	"mime"
	"strings"
	"time"
)

type Email struct {
	From    string
	To      []string
	Subject string
	Body    string
}

type Mailer struct {
}

func NewMailer() *Mailer {
	return &Mailer{}
}

func buildEmailMessage(from string, to []string, subject, body string) ([]byte, error) {
	// 防止邮件头注入：去除 From/To/Subject 中的换行符
	sanitizeHeader := func(s string) string {
		s = strings.ReplaceAll(s, "\r", "")
		s = strings.ReplaceAll(s, "\n", "")
		return s
	}

	cleanFrom := sanitizeHeader(from)
	cleanTo := make([]string, 0, len(to))
	for _, addr := range to {
		cleanTo = append(cleanTo, sanitizeHeader(addr))
	}
	cleanSubject := sanitizeHeader(subject)

	// 对 Subject 进行 MIME 编码，防止中文乱码或被拒收
	encodedSubject := mime.BEncoding.Encode("UTF-8", cleanSubject)
	// 添加 Date 头
	dateStr := time.Now().Format(time.RFC1123Z)
	toHeader := strings.Join(cleanTo, ", ")

	header := fmt.Sprintf("Date: %s\r\nFrom: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n",
		dateStr, cleanFrom, toHeader, encodedSubject)
	return []byte(header + body), nil
}
