package jwt

import (
	"testing"
	"time"
)

func testJWTService(duration time.Duration) *JWT {
	return NewJWT(&Config{
		JWTSecret: []byte("test_secret"),
		Duration:  duration,
	})
}

// 测试内容：验证登录令牌生成与解析的完整往返流程。
func TestLoginToken_RoundTrip(t *testing.T) {
	svc := testJWTService(time.Hour)
	token, err := svc.GenerateLoginToken(123, "alice", true)
	if err != nil {
		t.Fatalf("GenerateLoginToken 错误: %v", err)
	}
	claims, err := svc.ParseLoginToken(token)
	if err != nil {
		t.Fatalf("ParseLoginToken 错误: %v", err)
	}
	if claims.ID != 123 || claims.Username != "alice" || claims.Admin != true || claims.Type != "login" {
		t.Fatalf("非预期 claims: %+v", claims)
	}
}

// 测试内容：验证解析登录令牌时会拒绝错误类型的令牌。
func TestParseLoginToken_RejectsWrongType(t *testing.T) {
	svc := testJWTService(time.Hour)
	emailToken, err := svc.GenerateEmailToken(1, "a@example.com")
	if err != nil {
		t.Fatalf("GenerateEmailToken 错误: %v", err)
	}
	_, err = svc.ParseLoginToken(emailToken)
	if err == nil {
		t.Fatalf("期望错误的令牌类型返回错误")
	}
}

// 测试内容：验证过期的登录令牌会被解析为错误。
func TestParseLoginToken_Expired(t *testing.T) {
	svc := testJWTService(-1 * time.Second)
	token, err := svc.GenerateLoginToken(1, "alice", false)
	if err != nil {
		t.Fatalf("GenerateLoginToken 错误: %v", err)
	}
	_, err = svc.ParseLoginToken(token)
	if err == nil {
		t.Fatalf("期望返回令牌过期错误")
	}
}

// 测试内容：验证邮箱验证令牌生成与解析的往返流程。
func TestEmailToken_RoundTrip(t *testing.T) {
	svc := testJWTService(time.Hour)
	token, err := svc.GenerateEmailToken(1, "a@example.com")
	if err != nil {
		t.Fatalf("GenerateEmailToken: %v", err)
	}
	claims, err := svc.ParseEmailToken(token)
	if err != nil {
		t.Fatalf("ParseEmailToken: %v", err)
	}
	if claims.Email != "a@example.com" || claims.Type != "email_verify" {
		t.Fatalf("非预期 claims: %+v", claims)
	}
}
