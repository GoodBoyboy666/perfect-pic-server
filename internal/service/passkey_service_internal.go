package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	commonpkg "perfect-pic-server/internal/common"
	"perfect-pic-server/internal/consts"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/protocol/webauthncbor"
	"github.com/go-webauthn/webauthn/protocol/webauthncose"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/redis/go-redis/v9"
)

type passkeySessionEntry struct {
	PasskeySessionType consts.PasskeySessionType
	UserID             uint
	SessionData        webauthn.SessionData
	ExpiresAt          time.Time
}

type passkeyStoredCredential struct {
	ID              []byte                            `json:"id"`
	PublicKey       []byte                            `json:"publicKey"`
	AttestationType string                            `json:"attestationType"`
	Transport       []protocol.AuthenticatorTransport `json:"transport"`
	Flags           webauthn.CredentialFlags          `json:"flags"`
	Authenticator   webauthn.Authenticator            `json:"authenticator"`
}

var passkeySessionStore sync.Map

var passkeyAllowedCOSEAlgorithms = map[webauthncose.COSEAlgorithmIdentifier]struct{}{
	webauthncose.AlgEdDSA: {},
	webauthncose.AlgES256: {},
	webauthncose.AlgRS256: {},
}

// CreatePasskeyWebAuthnClient 根据系统配置构建 WebAuthn 客户端。
func (s *PasskeyService) CreatePasskeyWebAuthnClient() (*webauthn.WebAuthn, error) {
	baseURL := strings.TrimSpace(s.dbConfig.GetString(consts.ConfigBaseURL))
	if baseURL == "" {
		baseURL = "http://localhost"
	}

	// base_url 同时决定 RP ID 与 Origin，必须是完整可解析的绝对 URL。
	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil || parsedBaseURL.Scheme == "" || parsedBaseURL.Host == "" || parsedBaseURL.Hostname() == "" {
		return nil, commonpkg.NewValidationError("系统 base_url 配置无效，无法启用 Passkey")
	}

	siteName := strings.TrimSpace(s.dbConfig.GetString(consts.ConfigSiteName))
	if siteName == "" {
		siteName = "Perfect Pic"
	}

	return webauthn.New(&webauthn.Config{
		RPDisplayName: siteName,
		// RPID 必须是 host（不含端口/协议），认证器会严格校验。
		RPID: parsedBaseURL.Hostname(),
		// RPOrigins 需要精确包含协议+主机（含端口），用于浏览器端 origin 校验。
		RPOrigins: []string{parsedBaseURL.Scheme + "://" + parsedBaseURL.Host},
	})
}

// LoadUserPasskeyCredentials 读取并反序列化用户的 Passkey 凭据集合。
func (s *PasskeyService) LoadUserPasskeyCredentials(userID uint) ([]webauthn.Credential, error) {
	records, err := s.passkeyStore.ListPasskeyCredentialsByUserID(userID)
	if err != nil {
		return nil, commonpkg.NewInternalError("读取 Passkey 失败")
	}

	credentials := make([]webauthn.Credential, 0, len(records))
	for _, record := range records {
		var credential webauthn.Credential
		// DB 中保存的是完整 credential JSON，登录/注册排除都依赖其完整反序列化结果。
		if err := json.Unmarshal([]byte(record.Credential), &credential); err != nil {
			return nil, commonpkg.NewInternalError("Passkey 数据损坏，请重新绑定")
		}
		credentials = append(credentials, credential)
	}

	return credentials, nil
}

// StorePasskeySession 保存 Passkey 挑战会话并返回一次性会话 ID。
func (s *PasskeyService) StorePasskeySession(sessionType consts.PasskeySessionType, userID uint, session *webauthn.SessionData) (string, error) {
	if session == nil {
		return "", errors.New("passkey session is nil")
	}

	sessionID, err := generatePasskeySessionID()
	if err != nil {
		return "", err
	}

	expireAt := time.Now().Add(consts.PasskeySessionTTL)
	// 显式同步 Expires，确保后续库侧校验与本地过期策略一致。
	session.Expires = expireAt
	entry := passkeySessionEntry{
		PasskeySessionType: sessionType,
		UserID:             userID,
		SessionData:        *session,
		ExpiresAt:          expireAt,
	}

	// Redis 可用时优先写入 Redis，支持多实例共享会话。
	if storePasskeySessionInRedis(sessionID, entry) {
		return sessionID, nil
	}

	// Redis 不可用或写入失败时回退本地内存。
	storePasskeySessionInMemory(sessionID, entry)
	return sessionID, nil
}

func storePasskeySessionInRedis(sessionID string, entry passkeySessionEntry) bool {
	redisClient := GetRedisClient()
	if redisClient == nil {
		return false
	}

	payload, err := json.Marshal(entry)
	if err != nil {
		log.Printf("⚠️ Redis 写入 Passkey 会话失败，序列化异常，回退内存会话: %v", err)
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	key := RedisKey("passkey", "session", sessionID)
	if err := redisClient.Set(ctx, key, payload, consts.PasskeySessionTTL).Err(); err != nil {
		log.Printf("⚠️ Redis 写入 Passkey 会话失败，回退内存会话: %v", err)
		return false
	}

	return true
}

func storePasskeySessionInMemory(sessionID string, entry passkeySessionEntry) {
	// 每次写入前顺带清理过期会话，控制内存占用。
	cleanupExpiredPasskeySessions()
	passkeySessionStore.Store(sessionID, entry)
}

// ConsumePasskeyLoginSession 读取并消费登录会话，仅返回 WebAuthn 校验所需的 SessionData。
func (s *PasskeyService) ConsumePasskeyLoginSession(sessionID string) (*webauthn.SessionData, error) {
	entry, err := consumePasskeySessionEntry(sessionID, consts.PasskeySessionLogin)
	if err != nil {
		return nil, err
	}
	return &entry.SessionData, nil
}

// ConsumePasskeyRegistrationSession 读取并消费注册会话，并校验会话归属用户。
func (s *PasskeyService) ConsumePasskeyRegistrationSession(sessionID string, userID uint) (*webauthn.SessionData, error) {
	entry, err := consumePasskeySessionEntry(sessionID, consts.PasskeySessionRegistration)
	if err != nil {
		return nil, err
	}
	// 注册会话必须与当前登录用户一致，避免跨账号完成绑定。
	if entry.UserID != userID {
		return nil, commonpkg.NewForbiddenError("无权完成该 Passkey 注册会话")
	}
	return &entry.SessionData, nil
}

// consumePasskeySessionEntry 读取并消费底层会话条目，负责类型与过期校验。
func consumePasskeySessionEntry(sessionID string, expectedType consts.PasskeySessionType) (*passkeySessionEntry, error) {
	if strings.TrimSpace(sessionID) == "" {
		return nil, commonpkg.NewValidationError("session_id 不能为空")
	}

	// Redis 可用时优先从 Redis 原子读取并删除；未命中再回退本地内存。
	entry, err := consumePasskeySessionEntryFromRedis(sessionID)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		entry, err = consumePasskeySessionEntryFromMemory(sessionID)
		if err != nil {
			return nil, err
		}
	}

	// 防止把“注册会话”拿去走“登录校验”或反向混用。
	if entry.PasskeySessionType != expectedType {
		return nil, commonpkg.NewValidationError("Passkey 会话类型不匹配")
	}
	if time.Now().After(entry.ExpiresAt) {
		return nil, commonpkg.NewValidationError("Passkey 会话已过期，请重新发起")
	}

	return entry, nil
}

func consumePasskeySessionEntryFromRedis(sessionID string) (*passkeySessionEntry, error) {
	redisClient := GetRedisClient()
	if redisClient == nil {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	key := RedisKey("passkey", "session", sessionID)
	payload, err := redisClient.GetDel(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		// Redis 异常时回退本地内存，避免影响单机兼容路径。
		log.Printf("⚠️ Redis 读取 Passkey 会话失败，回退内存会话: %v", err)
		return nil, nil
	}

	if strings.TrimSpace(payload) == "" {
		return nil, nil
	}

	var entry passkeySessionEntry
	if err := json.Unmarshal([]byte(payload), &entry); err != nil {
		return nil, commonpkg.NewInternalError("Passkey 会话数据异常")
	}

	return &entry, nil
}

func consumePasskeySessionEntryFromMemory(sessionID string) (*passkeySessionEntry, error) {
	// LoadAndDelete 保证会话只可使用一次，天然抵御挑战重放。
	raw, ok := passkeySessionStore.LoadAndDelete(sessionID)
	if !ok {
		return nil, commonpkg.NewValidationError("Passkey 会话不存在或已过期，请重新发起")
	}

	entry, ok := raw.(passkeySessionEntry)
	if !ok {
		return nil, commonpkg.NewInternalError("Passkey 会话数据异常")
	}

	return &entry, nil
}

// cleanupExpiredPasskeySessions 清理内存中已过期的会话记录。
func cleanupExpiredPasskeySessions() {
	now := time.Now()
	passkeySessionStore.Range(func(key, value interface{}) bool {
		entry, ok := value.(passkeySessionEntry)
		if !ok || now.After(entry.ExpiresAt) {
			passkeySessionStore.Delete(key)
		}
		return true
	})
}

// generatePasskeySessionID 生成高熵的一次性会话 ID。
func generatePasskeySessionID() (string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(randomBytes), nil
}

// BuildPasskeyCredentialRequest 将前端凭据 JSON 包装成 WebAuthn 库可处理的 HTTP 请求。
func (s *PasskeyService) BuildPasskeyCredentialRequest(credentialJSON []byte) (*http.Request, error) {
	trimmed := bytes.TrimSpace(credentialJSON)
	if len(trimmed) == 0 {
		return nil, commonpkg.NewValidationError("credential 不能为空")
	}

	request, err := http.NewRequest(http.MethodPost, "/", bytes.NewReader(trimmed))
	if err != nil {
		return nil, commonpkg.NewInternalError("Passkey 请求构造失败")
	}
	request.Header.Set("Content-Type", "application/json")
	return request, nil
}

// ParsePasskeyUserHandle 将 discoverable 登录返回的 userHandle 解析为用户 ID。
func (s *PasskeyService) ParsePasskeyUserHandle(userHandle []byte) (uint, error) {
	if len(userHandle) == 0 {
		return 0, errors.New("user handle is empty")
	}

	// userHandle 由 WebAuthnID() 写入十进制 userID，这里按同一约定解析。
	parsed, err := strconv.ParseUint(string(userHandle), 10, 64)
	if err != nil || parsed == 0 {
		return 0, errors.New("invalid user handle")
	}
	if parsed > uint64(^uint(0)) {
		return 0, errors.New("user handle overflows uint")
	}
	return uint(parsed), nil
}

// EncodePasskeyCredentialID 将凭据 ID 编码为可存储字符串。
func (s *PasskeyService) EncodePasskeyCredentialID(credentialID []byte) string {
	return base64.RawURLEncoding.EncodeToString(credentialID)
}

// GetPasskeyRecommendedCredentialParameters 返回注册阶段允许的签名算法列表。
func (s *PasskeyService) GetPasskeyRecommendedCredentialParameters() []protocol.CredentialParameter {
	return webauthn.CredentialParametersRecommendedL3()
}

// IsPasskeyAlgorithmAllowed 判断凭据算法是否在系统允许的安全白名单中。
func (s *PasskeyService) IsPasskeyAlgorithmAllowed(algorithm int64) bool {
	_, ok := passkeyAllowedCOSEAlgorithms[webauthncose.COSEAlgorithmIdentifier(algorithm)]
	return ok
}

// ExtractPasskeyCredentialAlgorithm 从凭据中提取 COSE 算法标识。
// 部分浏览器不会回填 Attestation.PublicKeyAlgorithm，因此需要回退解析 credential.PublicKey。
func (s *PasskeyService) ExtractPasskeyCredentialAlgorithm(credential *webauthn.Credential) (webauthncose.COSEAlgorithmIdentifier, error) {
	if credential == nil {
		return 0, errors.New("credential is nil")
	}

	if credential.Attestation.PublicKeyAlgorithm != 0 {
		return webauthncose.COSEAlgorithmIdentifier(credential.Attestation.PublicKeyAlgorithm), nil
	}

	var publicKey webauthncose.PublicKeyData
	if err := webauthncbor.Unmarshal(credential.PublicKey, &publicKey); err != nil {
		return 0, err
	}
	return webauthncose.COSEAlgorithmIdentifier(publicKey.Algorithm), nil
}

// BuildDefaultPasskeyName 根据凭据 ID 构造默认名称，便于用户首次识别。
func (s *PasskeyService) BuildDefaultPasskeyName(credentialID string) string {
	short := credentialID
	if len(short) > 8 {
		short = short[:8]
	}
	return "Passkey-" + short
}

// normalizePasskeyName 清洗并校验用户输入的 Passkey 名称。
func normalizePasskeyName(name string) (string, error) {
	normalized := strings.TrimSpace(name)
	if normalized == "" {
		return "", commonpkg.NewValidationError("Passkey 名称不能为空")
	}
	if utf8.RuneCountInString(normalized) > consts.PasskeyNameMaxRunes {
		return "", commonpkg.NewValidationError("Passkey 名称长度不能超过 64 个字符")
	}
	return normalized, nil
}

// MarshalPasskeyCredential 将凭据对象序列化为存储用 JSON（不包含 Attestation 大字段）。
func (s *PasskeyService) MarshalPasskeyCredential(credential *webauthn.Credential) (string, error) {
	if credential == nil {
		return "", errors.New("credential is nil")
	}

	storedCredential := passkeyStoredCredential{
		ID:              credential.ID,
		PublicKey:       credential.PublicKey,
		AttestationType: credential.AttestationType,
		Transport:       credential.Transport,
		Flags:           credential.Flags,
		Authenticator:   credential.Authenticator,
	}

	raw, err := json.Marshal(storedCredential)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

// IsPasskeyUniqueConstraintConflict 判断数据库错误是否属于唯一约束冲突。
func (s *PasskeyService) IsPasskeyUniqueConstraintConflict(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique") || strings.Contains(msg, "duplicate")
}
