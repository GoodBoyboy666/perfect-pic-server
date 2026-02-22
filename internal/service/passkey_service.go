package service

import (
	"errors"
	"perfect-pic-server/internal/model"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"gorm.io/gorm"
)

// UserPasskey 是返回给前端的用户 Passkey 列表项。
type UserPasskey struct {
	ID           uint   `json:"id"`
	CredentialID string `json:"credential_id"`
	Name         string `json:"name"`
	CreatedAt    int64  `json:"created_at"`
}

// BeginPasskeyRegistration 为当前用户创建 Passkey 注册挑战并返回会话 ID。
func (s *AppService) BeginPasskeyRegistration(userID uint) (string, *protocol.CredentialCreation, error) {
	// 先校验容量上限，避免在已达上限时仍创建挑战造成无效流程。
	if err := s.ensureUserPasskeyCapacity(userID); err != nil {
		return "", nil, err
	}

	webauthnClient, err := s.createPasskeyWebAuthnClient()
	if err != nil {
		return "", nil, err
	}

	passkeyUser, err := s.loadPasskeyWebAuthnUser(userID, passkeyWebAuthnUserLoadModeRegistration)
	if err != nil {
		return "", nil, err
	}

	creation, sessionData, err := webauthnClient.BeginRegistration(
		passkeyUser,
		// 限制可接受的公钥算法集合，只允许较安全的推荐算法。
		webauthn.WithCredentialParameters(getPasskeyRecommendedCredentialParameters()),
		// 要求创建可发现凭据（Resident Key），用于“无需用户名”的 Passkey 登录。
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementRequired),
		// 将已绑定凭据放入排除列表，阻止同一凭据重复注册。
		webauthn.WithExclusions(webauthn.Credentials(passkeyUser.credentials).CredentialDescriptors()),
		// 让客户端返回 credential properties（例如 rk），便于前端感知凭据能力。
		webauthn.WithExtensions(protocol.AuthenticationExtensions{"credProps": true}),
	)
	if err != nil {
		return "", nil, NewInternalError("创建 Passkey 注册挑战失败")
	}

	// 服务端保存一次性会话，并仅把 session_id 发给前端。
	sessionID, err := storePasskeySession(passkeySessionRegistration, userID, sessionData)
	if err != nil {
		return "", nil, NewInternalError("创建 Passkey 注册会话失败")
	}

	return sessionID, creation, nil
}

// FinishPasskeyRegistration 校验并完成 Passkey 注册，随后持久化凭据。
func (s *AppService) FinishPasskeyRegistration(userID uint, sessionID string, credentialJSON []byte) error {
	// 读取并消费注册会话，同时校验该会话必须归属当前登录用户。
	sessionData, err := consumePasskeyRegistrationSession(sessionID, userID)
	if err != nil {
		return err
	}

	webauthnClient, err := s.createPasskeyWebAuthnClient()
	if err != nil {
		return err
	}

	passkeyUser, err := s.loadPasskeyWebAuthnUser(userID, passkeyWebAuthnUserLoadModeRegistration)
	if err != nil {
		return err
	}

	request, err := buildPasskeyCredentialRequest(credentialJSON)
	if err != nil {
		return err
	}

	// 使用服务端保存的 challenge/session 校验前端返回的 attestation。
	credential, err := webauthnClient.FinishRegistration(passkeyUser, *sessionData, request)
	if err != nil {
		return NewValidationError("Passkey 注册校验失败，请重试")
	}

	// 注册阶段显式拒绝不在白名单内的签名算法。
	credentialAlgorithm, err := extractPasskeyCredentialAlgorithm(credential)
	if err != nil || !isPasskeyAlgorithmAllowed(int64(credentialAlgorithm)) {
		return NewValidationError("Passkey 签名算法不被允许")
	}

	credentialID := encodePasskeyCredentialID(credential.ID)
	// 先按 credential_id 查重，区分“同账号重复绑定”和“被其他账号占用”。
	existing, findErr := s.repos.User.FindPasskeyCredentialByCredentialID(credentialID)
	if findErr == nil {
		if existing.UserID == userID {
			return NewConflictError("该 Passkey 已绑定")
		}
		return NewConflictError("该 Passkey 已被其他账号绑定")
	}
	if !errors.Is(findErr, gorm.ErrRecordNotFound) {
		return NewInternalError("保存 Passkey 失败")
	}

	// 注册完成前再次校验容量，防止并发场景下绕过上限。
	if err := s.ensureUserPasskeyCapacity(userID); err != nil {
		return err
	}

	// 持久化完整 credential（包含 signCount/flags/transport 等后续验签元数据）。
	serialized, err := marshalPasskeyCredential(credential)
	if err != nil {
		return NewInternalError("保存 Passkey 失败")
	}

	if err := s.repos.User.CreatePasskeyCredential(&model.PasskeyCredential{
		UserID:       userID,
		CredentialID: credentialID,
		Name:         buildDefaultPasskeyName(credentialID),
		Credential:   serialized,
	}); err != nil {
		if isPasskeyUniqueConstraintConflict(err) {
			return NewConflictError("该 Passkey 已绑定")
		}
		return NewInternalError("保存 Passkey 失败")
	}

	return nil
}

// ListUserPasskeys 返回指定用户已绑定的 Passkey 列表。
func (s *AppService) ListUserPasskeys(userID uint) ([]UserPasskey, error) {
	if _, err := s.GetUserByID(userID); err != nil {
		return nil, err
	}

	records, err := s.repos.User.ListPasskeyCredentialsByUserID(userID)
	if err != nil {
		return nil, NewInternalError("读取 Passkey 列表失败")
	}

	items := make([]UserPasskey, 0, len(records))
	for _, record := range records {
		items = append(items, UserPasskey{
			ID:           record.ID,
			CredentialID: record.CredentialID,
			Name:         record.Name,
			CreatedAt:    record.CreatedAt.Unix(),
		})
	}

	return items, nil
}

// DeleteUserPasskey 删除指定用户名下的某个 Passkey。
func (s *AppService) DeleteUserPasskey(userID uint, passkeyID uint) error {
	if passkeyID == 0 {
		return NewValidationError("无效的 Passkey ID")
	}

	if err := s.repos.User.DeletePasskeyCredentialByID(userID, passkeyID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NewNotFoundError("Passkey 不存在")
		}
		return NewInternalError("删除 Passkey 失败")
	}
	return nil
}

// UpdateUserPasskeyName 更新指定用户名下某个 Passkey 的显示名称。
func (s *AppService) UpdateUserPasskeyName(userID uint, passkeyID uint, name string) error {
	if passkeyID == 0 {
		return NewValidationError("无效的 Passkey ID")
	}

	normalizedName, err := normalizePasskeyName(name)
	if err != nil {
		return err
	}

	if err := s.repos.User.UpdatePasskeyCredentialNameByID(userID, passkeyID, normalizedName); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NewNotFoundError("Passkey 不存在")
		}
		return NewInternalError("更新 Passkey 名称失败")
	}
	return nil
}

// BeginPasskeyLogin 创建无用户名（discoverable）的 Passkey 登录挑战。
func (s *AppService) BeginPasskeyLogin() (string, *protocol.CredentialAssertion, error) {
	webauthnClient, err := s.createPasskeyWebAuthnClient()
	if err != nil {
		return "", nil, err
	}

	assertion, sessionData, err := webauthnClient.BeginDiscoverableLogin(
		// 登录场景偏向体验，优先请求 UV（支持设备会自动触发生物认证/PIN）。
		webauthn.WithUserVerification(protocol.VerificationPreferred),
	)
	if err != nil {
		return "", nil, NewInternalError("创建 Passkey 登录挑战失败")
	}

	// 登录挑战同样只在服务端保存明文 SessionData，前端仅持有 session_id。
	sessionID, err := storePasskeySession(passkeySessionLogin, 0, sessionData)
	if err != nil {
		return "", nil, NewInternalError("创建 Passkey 登录会话失败")
	}

	return sessionID, assertion, nil
}

// FinishPasskeyLogin 完成 Passkey 登录校验并签发 JWT。
//
//nolint:gocyclo
func (s *AppService) FinishPasskeyLogin(sessionID string, credentialJSON []byte) (string, error) {
	// 登录挑战一次性消费，防止 assertion 重放攻击。
	sessionData, err := consumePasskeyLoginSession(sessionID)
	if err != nil {
		return "", err
	}

	webauthnClient, err := s.createPasskeyWebAuthnClient()
	if err != nil {
		return "", err
	}

	request, err := buildPasskeyCredentialRequest(credentialJSON)
	if err != nil {
		return "", err
	}

	var resolvedUser *passkeyWebAuthnUser
	validatedUser, validatedCredential, err := webauthnClient.FinishPasskeyLogin(
		func(rawID, userHandle []byte) (webauthn.User, error) {
			// discoverable 流程下 userHandle 由认证器返回，这里按约定解析为 userID。
			userID, parseErr := parsePasskeyUserHandle(userHandle)
			if parseErr != nil {
				return nil, parseErr
			}

			// 登录校验阶段仅需用户凭据集合，不必提前查询完整用户资料。
			passkeyUser, loadErr := s.loadPasskeyWebAuthnUser(userID, passkeyWebAuthnUserLoadModeLogin)
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
		return "", NewUnauthorizedError("Passkey 登录失败")
	}

	passkeyUser, ok := validatedUser.(*passkeyWebAuthnUser)
	if !ok {
		// 正常情况下会是 *passkeyWebAuthnUser，这里保留兜底避免类型差异导致空指针。
		if resolvedUser == nil {
			return "", NewInternalError("Passkey 登录失败")
		}
		passkeyUser = resolvedUser
	}
	// 登录阶段同样校验算法白名单，阻断不符合策略的历史凭据。
	credentialAlgorithm, err := extractPasskeyCredentialAlgorithm(validatedCredential)
	if err != nil || !isPasskeyAlgorithmAllowed(int64(credentialAlgorithm)) {
		return "", NewUnauthorizedError("Passkey 签名算法不被允许")
	}

	// 将本次验证后更新过的 credential 写回库（尤其 signCount），用于后续反重放校验。
	serialized, err := marshalPasskeyCredential(validatedCredential)
	if err != nil {
		return "", NewInternalError("Passkey 登录失败")
	}

	if err := s.repos.User.UpdatePasskeyCredentialData(
		passkeyUser.userID,
		encodePasskeyCredentialID(validatedCredential.ID),
		serialized,
	); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", NewUnauthorizedError("Passkey 登录失败")
		}
		return "", NewInternalError("Passkey 登录失败")
	}

	// 验签通过后再查完整用户，复用统一登录准入策略（状态/邮箱验证/管理员规则等）。
	user, err := s.repos.User.FindByID(passkeyUser.userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", NewUnauthorizedError("Passkey 登录失败")
		}
		return "", NewInternalError("Passkey 登录失败")
	}

	// 统一通过既有发 token 逻辑签发 JWT，确保与密码登录行为一致。
	token, err := s.issueLoginToken(user)
	if err != nil {
		return "", err
	}
	return token, nil
}
