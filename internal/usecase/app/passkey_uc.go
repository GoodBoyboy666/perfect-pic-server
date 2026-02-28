package app

import (
	"errors"
	commonpkg "perfect-pic-server/internal/common"
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/model"
	"strconv"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"gorm.io/gorm"
)

type passkeyWebAuthnUserLoadMode string

const (
	passkeyWebAuthnUserLoadModeRegistration passkeyWebAuthnUserLoadMode = "registration"
	passkeyWebAuthnUserLoadModeLogin        passkeyWebAuthnUserLoadMode = "login"
)

type passkeyWebAuthnUser struct {
	userID      uint
	username    string
	id          []byte
	credentials []webauthn.Credential
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

// BeginPasskeyRegistration 为当前用户创建 Passkey 注册挑战并返回会话 ID。
func (c *PasskeyUseCase) BeginPasskeyRegistration(userID uint) (string, *protocol.CredentialCreation, error) {
	// 先校验容量上限，避免在已达上限时仍创建挑战造成无效流程。
	if err := c.ensureUserPasskeyCapacity(userID); err != nil {
		return "", nil, err
	}

	webauthnClient, err := c.passkeyService.CreatePasskeyWebAuthnClient()
	if err != nil {
		return "", nil, err
	}

	passkeyUser, err := c.loadPasskeyWebAuthnUser(userID, passkeyWebAuthnUserLoadModeRegistration)
	if err != nil {
		return "", nil, err
	}

	creation, sessionData, err := webauthnClient.BeginRegistration(
		passkeyUser,
		// 限制可接受的公钥算法集合，只允许较安全的推荐算法。
		webauthn.WithCredentialParameters(c.passkeyService.GetPasskeyRecommendedCredentialParameters()),
		// 要求创建可发现凭据（Resident Key），用于“无需用户名”的 Passkey 登录。
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementRequired),
		// 将已绑定凭据放入排除列表，阻止同一凭据重复注册。
		webauthn.WithExclusions(webauthn.Credentials(passkeyUser.credentials).CredentialDescriptors()),
		// 让客户端返回 credential properties（例如 rk），便于前端感知凭据能力。
		webauthn.WithExtensions(protocol.AuthenticationExtensions{"credProps": true}),
	)
	if err != nil {
		return "", nil, commonpkg.NewInternalError("创建 Passkey 注册挑战失败")
	}

	// 服务端保存一次性会话，并仅把 session_id 发给前端。
	sessionID, err := c.passkeyService.StorePasskeySession(consts.PasskeySessionRegistration, userID, sessionData)
	if err != nil {
		return "", nil, commonpkg.NewInternalError("创建 Passkey 注册会话失败")
	}

	return sessionID, creation, nil
}

// FinishPasskeyRegistration 校验并完成 Passkey 注册，随后持久化凭据。
func (c *PasskeyUseCase) FinishPasskeyRegistration(userID uint, sessionID string, credentialJSON []byte) error {
	// 读取并消费注册会话，同时校验该会话必须归属当前登录用户。
	sessionData, err := c.passkeyService.ConsumePasskeyRegistrationSession(sessionID, userID)
	if err != nil {
		return err
	}

	webauthnClient, err := c.passkeyService.CreatePasskeyWebAuthnClient()
	if err != nil {
		return err
	}

	passkeyUser, err := c.loadPasskeyWebAuthnUser(userID, passkeyWebAuthnUserLoadModeRegistration)
	if err != nil {
		return err
	}

	request, err := c.passkeyService.BuildPasskeyCredentialRequest(credentialJSON)
	if err != nil {
		return err
	}

	// 使用服务端保存的 challenge/session 校验前端返回的 attestation。
	credential, err := webauthnClient.FinishRegistration(passkeyUser, *sessionData, request)
	if err != nil {
		return commonpkg.NewValidationError("Passkey 注册校验失败，请重试")
	}

	// 注册阶段显式拒绝不在白名单内的签名算法。
	credentialAlgorithm, err := c.passkeyService.ExtractPasskeyCredentialAlgorithm(credential)
	if err != nil || !c.passkeyService.IsPasskeyAlgorithmAllowed(int64(credentialAlgorithm)) {
		return commonpkg.NewValidationError("Passkey 签名算法不被允许")
	}

	credentialID := c.passkeyService.EncodePasskeyCredentialID(credential.ID)
	// 先按 credential_id 查重，区分“同账号重复绑定”和“被其他账号占用”。
	existing, findErr := c.passkeyStore.FindPasskeyCredentialByCredentialID(credentialID)
	if findErr == nil {
		if existing.UserID == userID {
			return commonpkg.NewConflictError("该 Passkey 已绑定")
		}
		return commonpkg.NewConflictError("该 Passkey 已被其他账号绑定")
	}
	if !errors.Is(findErr, gorm.ErrRecordNotFound) {
		return commonpkg.NewInternalError("保存 Passkey 失败")
	}

	// 注册完成前再次校验容量，防止并发场景下绕过上限。
	if err := c.ensureUserPasskeyCapacity(userID); err != nil {
		return err
	}

	// 持久化登录验签所需核心字段。
	serialized, err := c.passkeyService.MarshalPasskeyCredential(credential)
	if err != nil {
		return commonpkg.NewInternalError("保存 Passkey 失败")
	}

	if err := c.passkeyService.CreatePasskeyCredential(&model.PasskeyCredential{
		UserID:       userID,
		CredentialID: credentialID,
		Name:         c.passkeyService.BuildDefaultPasskeyName(credentialID),
		Credential:   serialized,
	}); err != nil {
		if c.passkeyService.IsPasskeyUniqueConstraintConflict(err) {
			return commonpkg.NewConflictError("该 Passkey 已绑定")
		}
		return commonpkg.NewInternalError("保存 Passkey 失败")
	}

	return nil
}

// BeginPasskeyLogin 创建无用户名（discoverable）的 Passkey 登录挑战。
func (c *PasskeyUseCase) BeginPasskeyLogin() (string, *protocol.CredentialAssertion, error) {
	webauthnClient, err := c.passkeyService.CreatePasskeyWebAuthnClient()
	if err != nil {
		return "", nil, err
	}

	assertion, sessionData, err := webauthnClient.BeginDiscoverableLogin(
		// 登录场景偏向体验，优先请求 UV（支持设备会自动触发生物认证/PIN）。
		webauthn.WithUserVerification(protocol.VerificationPreferred),
	)
	if err != nil {
		return "", nil, commonpkg.NewInternalError("创建 Passkey 登录挑战失败")
	}

	// 登录挑战同样只在服务端保存明文 SessionData，前端仅持有 session_id。
	sessionID, err := c.passkeyService.StorePasskeySession(consts.PasskeySessionLogin, 0, sessionData)
	if err != nil {
		return "", nil, commonpkg.NewInternalError("创建 Passkey 登录会话失败")
	}

	return sessionID, assertion, nil
}

// FinishPasskeyLogin 完成 Passkey 登录校验并签发 JWT。
//
//nolint:gocyclo
func (c *PasskeyUseCase) FinishPasskeyLogin(sessionID string, credentialJSON []byte) (string, error) {
	// 登录挑战一次性消费，防止 assertion 重放攻击。
	sessionData, err := c.passkeyService.ConsumePasskeyLoginSession(sessionID)
	if err != nil {
		return "", err
	}

	webauthnClient, err := c.passkeyService.CreatePasskeyWebAuthnClient()
	if err != nil {
		return "", err
	}

	request, err := c.passkeyService.BuildPasskeyCredentialRequest(credentialJSON)
	if err != nil {
		return "", err
	}

	var resolvedUser *passkeyWebAuthnUser
	validatedUser, validatedCredential, err := webauthnClient.FinishPasskeyLogin(
		func(rawID, userHandle []byte) (webauthn.User, error) {
			// discoverable 流程下 userHandle 由认证器返回，这里按约定解析为 userID。
			userID, parseErr := c.passkeyService.ParsePasskeyUserHandle(userHandle)
			if parseErr != nil {
				return nil, parseErr
			}

			// 登录校验阶段仅需用户凭据集合，不必提前查询完整用户资料。
			passkeyUser, loadErr := c.loadPasskeyWebAuthnUser(userID, passkeyWebAuthnUserLoadModeLogin)
			if loadErr != nil {
				return nil, loadErr
			}
			resolvedUser = passkeyUser
			_ = rawID // 库内部会结合 rawID 与凭据列表匹配，此处无需额外使用。
			return passkeyUser, nil
		},
		*sessionData,
		request,
	)
	if err != nil {
		return "", commonpkg.NewUnauthorizedError("Passkey 登录失败")
	}

	passkeyUser, ok := validatedUser.(*passkeyWebAuthnUser)
	if !ok {
		// 正常情况下会是 *passkeyWebAuthnUser，这里保留兜底避免类型差异导致空指针。
		if resolvedUser == nil {
			return "", commonpkg.NewInternalError("Passkey 登录失败")
		}
		passkeyUser = resolvedUser
	}
	// 登录阶段同样校验算法白名单，阻断不符合策略的历史凭据。
	credentialAlgorithm, err := c.passkeyService.ExtractPasskeyCredentialAlgorithm(validatedCredential)
	if err != nil || !c.passkeyService.IsPasskeyAlgorithmAllowed(int64(credentialAlgorithm)) {
		return "", commonpkg.NewUnauthorizedError("Passkey 签名算法不被允许")
	}

	// 将本次验证后更新过的 credential 写回库（尤其 signCount），用于后续反重放校验。
	serialized, err := c.passkeyService.MarshalPasskeyCredential(validatedCredential)
	if err != nil {
		return "", commonpkg.NewInternalError("Passkey 登录失败")
	}

	if err := c.passkeyService.UpdatePasskeyCredentialData(
		passkeyUser.userID,
		c.passkeyService.EncodePasskeyCredentialID(validatedCredential.ID),
		serialized,
	); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", commonpkg.NewUnauthorizedError("Passkey 登录失败")
		}
		return "", commonpkg.NewInternalError("Passkey 登录失败")
	}

	// 验签通过后再查完整用户，复用统一登录准入策略（状态/邮箱验证/管理员规则等）。
	user, err := c.userStore.FindByID(passkeyUser.userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", commonpkg.NewUnauthorizedError("Passkey 登录失败")
		}
		return "", commonpkg.NewInternalError("Passkey 登录失败")
	}

	// 统一通过既有发 token 逻辑签发 JWT，确保与密码登录行为一致。
	token, err := c.authService.IssueLoginToken(user)
	if err != nil {
		return "", err
	}
	return token, nil
}

// ensureUserPasskeyCapacity 检查用户 Passkey 是否达到上限。
func (c *PasskeyUseCase) ensureUserPasskeyCapacity(userID uint) error {
	// 以数据库实时计数为准，避免依赖客户端状态导致上限失效。
	count, err := c.passkeyStore.CountPasskeyCredentialsByUserID(userID)
	if err != nil {
		return commonpkg.NewInternalError("校验 Passkey 数量失败")
	}
	if count >= consts.MaxUserPasskeyCount {
		return commonpkg.NewConflictError("Passkey 数量已达上限（最多 10 个）")
	}
	return nil
}

// loadPasskeyWebAuthnUser 加载并构造 WebAuthn User，按场景决定是否附带用户资料。
func (c *PasskeyUseCase) loadPasskeyWebAuthnUser(
	userID uint,
	loadMode passkeyWebAuthnUserLoadMode,
) (*passkeyWebAuthnUser, error) {
	resolvedUserID := userID
	username := ""

	switch loadMode {
	case passkeyWebAuthnUserLoadModeRegistration:
		user, err := c.userStore.FindByID(userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, commonpkg.NewNotFoundError("用户不存在")
			}
			return nil, commonpkg.NewInternalError("读取用户信息失败")
		}
		resolvedUserID = user.ID
		username = user.Username
	case passkeyWebAuthnUserLoadModeLogin:
		// 登录校验只需要稳定 userHandle 与凭据集合，不依赖用户名展示信息。
	default:
		return nil, commonpkg.NewInternalError("Passkey 用户加载模式无效")
	}

	credentials, err := c.passkeyService.LoadUserPasskeyCredentials(resolvedUserID)
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
