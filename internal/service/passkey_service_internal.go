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
	"gorm.io/gorm"
)

type passkeySessionType string

type passkeyWebAuthnUserLoadMode string

const (
	passkeySessionRegistration passkeySessionType = "registration"
	passkeySessionLogin        passkeySessionType = "login"
	passkeySessionTTL                             = 5 * time.Minute
	maxUserPasskeyCount                           = 10
	passkeyNameMaxRunes                           = 64
)

const (
	passkeyWebAuthnUserLoadModeRegistration passkeyWebAuthnUserLoadMode = "registration"
	passkeyWebAuthnUserLoadModeLogin        passkeyWebAuthnUserLoadMode = "login"
)

type passkeySessionEntry struct {
	PasskeySessionType passkeySessionType
	UserID             uint
	SessionData        webauthn.SessionData
	ExpiresAt          time.Time
}

type passkeyWebAuthnUser struct {
	userID      uint
	username    string
	id          []byte
	credentials []webauthn.Credential
}

var passkeySessionStore sync.Map

var passkeyAllowedCOSEAlgorithms = map[webauthncose.COSEAlgorithmIdentifier]struct{}{
	webauthncose.AlgEdDSA: {},
	webauthncose.AlgES256: {},
	webauthncose.AlgRS256: {},
}

func (u *passkeyWebAuthnUser) WebAuthnID() []byte {
	return u.id
}

func (u *passkeyWebAuthnUser) WebAuthnName() string {
	return u.username
}

func (u *passkeyWebAuthnUser) WebAuthnDisplayName() string {
	return u.username
}

func (u *passkeyWebAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	return u.credentials
}

// createPasskeyWebAuthnClient 根据系统配置构建 WebAuthn 客户端。
func (s *AppService) createPasskeyWebAuthnClient() (*webauthn.WebAuthn, error) {
	baseURL := strings.TrimSpace(s.GetString(consts.ConfigBaseURL))
	if baseURL == "" {
		baseURL = "http://localhost"
	}

	// base_url 同时决定 RP ID 与 Origin，必须是完整可解析的绝对 URL。
	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil || parsedBaseURL.Scheme == "" || parsedBaseURL.Host == "" || parsedBaseURL.Hostname() == "" {
		return nil, NewValidationError("系统 base_url 配置无效，无法启用 Passkey")
	}

	siteName := strings.TrimSpace(s.GetString(consts.ConfigSiteName))
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

// loadPasskeyWebAuthnUser 加载并构造 WebAuthn User，按场景决定是否附带用户资料。
func (s *AppService) loadPasskeyWebAuthnUser(
	userID uint,
	loadMode passkeyWebAuthnUserLoadMode,
) (*passkeyWebAuthnUser, error) {
	resolvedUserID := userID
	username := ""

	switch loadMode {
	case passkeyWebAuthnUserLoadModeRegistration:
		user, err := s.repos.User.FindByID(userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, NewNotFoundError("用户不存在")
			}
			return nil, NewInternalError("读取用户信息失败")
		}
		resolvedUserID = user.ID
		username = user.Username
	case passkeyWebAuthnUserLoadModeLogin:
		// 登录校验只需要稳定 userHandle 与凭据集合，不依赖用户名展示信息。
	default:
		return nil, NewInternalError("Passkey 用户加载模式无效")
	}

	credentials, err := s.loadUserPasskeyCredentials(resolvedUserID)
	if err != nil {
		return nil, err
	}

	return &passkeyWebAuthnUser{
		userID:   resolvedUserID,
		username: username,
		// userHandle 使用十进制 userID 字节串，登录回调中可无歧义反解析。
		id:          []byte(strconv.FormatUint(uint64(resolvedUserID), 10)),
		credentials: credentials,
	}, nil
}

// loadUserPasskeyCredentials 读取并反序列化用户的 Passkey 凭据集合。
func (s *AppService) loadUserPasskeyCredentials(userID uint) ([]webauthn.Credential, error) {
	records, err := s.repos.User.ListPasskeyCredentialsByUserID(userID)
	if err != nil {
		return nil, NewInternalError("读取 Passkey 失败")
	}

	credentials := make([]webauthn.Credential, 0, len(records))
	for _, record := range records {
		var credential webauthn.Credential
		// DB 中保存的是完整 credential JSON，登录/注册排除都依赖其完整反序列化结果。
		if err := json.Unmarshal([]byte(record.Credential), &credential); err != nil {
			return nil, NewInternalError("Passkey 数据损坏，请重新绑定")
		}
		credentials = append(credentials, credential)
	}

	return credentials, nil
}

// storePasskeySession 保存 Passkey 挑战会话并返回一次性会话 ID。
func storePasskeySession(sessionType passkeySessionType, userID uint, session *webauthn.SessionData) (string, error) {
	if session == nil {
		return "", errors.New("passkey session is nil")
	}

	sessionID, err := generatePasskeySessionID()
	if err != nil {
		return "", err
	}

	expireAt := time.Now().Add(passkeySessionTTL)
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
	if err := redisClient.Set(ctx, key, payload, passkeySessionTTL).Err(); err != nil {
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

// consumePasskeyLoginSession 读取并消费登录会话，仅返回 WebAuthn 校验所需的 SessionData。
func consumePasskeyLoginSession(sessionID string) (*webauthn.SessionData, error) {
	entry, err := consumePasskeySessionEntry(sessionID, passkeySessionLogin)
	if err != nil {
		return nil, err
	}
	return &entry.SessionData, nil
}

// consumePasskeyRegistrationSession 读取并消费注册会话，并校验会话归属用户。
func consumePasskeyRegistrationSession(sessionID string, userID uint) (*webauthn.SessionData, error) {
	entry, err := consumePasskeySessionEntry(sessionID, passkeySessionRegistration)
	if err != nil {
		return nil, err
	}
	// 注册会话必须与当前登录用户一致，避免跨账号完成绑定。
	if entry.UserID != userID {
		return nil, NewForbiddenError("无权完成该 Passkey 注册会话")
	}
	return &entry.SessionData, nil
}

// consumePasskeySessionEntry 读取并消费底层会话条目，负责类型与过期校验。
func consumePasskeySessionEntry(sessionID string, expectedType passkeySessionType) (*passkeySessionEntry, error) {
	if strings.TrimSpace(sessionID) == "" {
		return nil, NewValidationError("session_id 不能为空")
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
		return nil, NewValidationError("Passkey 会话类型不匹配")
	}
	if time.Now().After(entry.ExpiresAt) {
		return nil, NewValidationError("Passkey 会话已过期，请重新发起")
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
		return nil, NewInternalError("Passkey 会话数据异常")
	}

	return &entry, nil
}

func consumePasskeySessionEntryFromMemory(sessionID string) (*passkeySessionEntry, error) {
	// LoadAndDelete 保证会话只可使用一次，天然抵御挑战重放。
	raw, ok := passkeySessionStore.LoadAndDelete(sessionID)
	if !ok {
		return nil, NewValidationError("Passkey 会话不存在或已过期，请重新发起")
	}

	entry, ok := raw.(passkeySessionEntry)
	if !ok {
		return nil, NewInternalError("Passkey 会话数据异常")
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

// buildPasskeyCredentialRequest 将前端凭据 JSON 包装成 WebAuthn 库可处理的 HTTP 请求。
func buildPasskeyCredentialRequest(credentialJSON []byte) (*http.Request, error) {
	trimmed := bytes.TrimSpace(credentialJSON)
	if len(trimmed) == 0 {
		return nil, NewValidationError("credential 不能为空")
	}

	request, err := http.NewRequest(http.MethodPost, "/", bytes.NewReader(trimmed))
	if err != nil {
		return nil, NewInternalError("Passkey 请求构造失败")
	}
	request.Header.Set("Content-Type", "application/json")
	return request, nil
}

// parsePasskeyUserHandle 将 discoverable 登录返回的 userHandle 解析为用户 ID。
func parsePasskeyUserHandle(userHandle []byte) (uint, error) {
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

// encodePasskeyCredentialID 将凭据 ID 编码为可存储字符串。
func encodePasskeyCredentialID(credentialID []byte) string {
	return base64.RawURLEncoding.EncodeToString(credentialID)
}

// getPasskeyRecommendedCredentialParameters 返回注册阶段允许的签名算法列表。
func getPasskeyRecommendedCredentialParameters() []protocol.CredentialParameter {
	return webauthn.CredentialParametersRecommendedL3()
}

// isPasskeyAlgorithmAllowed 判断凭据算法是否在系统允许的安全白名单中。
func isPasskeyAlgorithmAllowed(algorithm int64) bool {
	_, ok := passkeyAllowedCOSEAlgorithms[webauthncose.COSEAlgorithmIdentifier(algorithm)]
	return ok
}

// extractPasskeyCredentialAlgorithm 从凭据中提取 COSE 算法标识。
// 部分浏览器不会回填 Attestation.PublicKeyAlgorithm，因此需要回退解析 credential.PublicKey。
func extractPasskeyCredentialAlgorithm(credential *webauthn.Credential) (webauthncose.COSEAlgorithmIdentifier, error) {
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

// buildDefaultPasskeyName 根据凭据 ID 构造默认名称，便于用户首次识别。
func buildDefaultPasskeyName(credentialID string) string {
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
		return "", NewValidationError("Passkey 名称不能为空")
	}
	if utf8.RuneCountInString(normalized) > passkeyNameMaxRunes {
		return "", NewValidationError("Passkey 名称长度不能超过 64 个字符")
	}
	return normalized, nil
}

// marshalPasskeyCredential 将凭据对象序列化为 JSON 字符串。
func marshalPasskeyCredential(credential *webauthn.Credential) (string, error) {
	raw, err := json.Marshal(credential)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

// isPasskeyUniqueConstraintConflict 判断数据库错误是否属于唯一约束冲突。
func isPasskeyUniqueConstraintConflict(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique") || strings.Contains(msg, "duplicate")
}

// ensureUserPasskeyCapacity 检查用户 Passkey 是否达到上限。
func (s *AppService) ensureUserPasskeyCapacity(userID uint) error {
	// 以数据库实时计数为准，避免依赖客户端状态导致上限失效。
	count, err := s.repos.User.CountPasskeyCredentialsByUserID(userID)
	if err != nil {
		return NewInternalError("校验 Passkey 数量失败")
	}
	if count >= maxUserPasskeyCount {
		return NewConflictError("Passkey 数量已达上限（最多 10 个）")
	}
	return nil
}
