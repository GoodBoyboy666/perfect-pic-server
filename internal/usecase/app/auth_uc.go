package app

import (
	"errors"
	"fmt"
	"perfect-pic-server/internal/common/httpx"
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/model"
	"perfect-pic-server/internal/utils"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// LoginUser 执行登录鉴权并返回登录令牌。
func (c *AuthUseCase) LoginUser(username, password string) (string, error) {
	user, err := c.userStore.FindByUsername(username)
	if err != nil {
		return "", httpx.NewAuthError(httpx.AuthErrorUnauthorized, "用户名或密码错误")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", httpx.NewAuthError(httpx.AuthErrorUnauthorized, "用户名或密码错误")
	}

	return c.authService.IssueLoginToken(user)
}

// RegisterUser 执行用户注册并异步发送邮箱验证邮件。
//
//nolint:gocyclo
func (c *AuthUseCase) RegisterUser(username, password, email string) error {
	// 系统未初始化时禁止注册：避免在还未创建管理员/完成基础配置时产生普通用户。
	if !c.initService.IsSystemInitialized() {
		return httpx.NewAuthError(httpx.AuthErrorForbidden, "系统尚未初始化，请先完成初始化")
	}

	if ok, msg := utils.ValidatePassword(password); !ok {
		return httpx.NewAuthError(httpx.AuthErrorValidation, msg)
	}

	if ok, msg := utils.ValidateUsername(username); !ok {
		return httpx.NewAuthError(httpx.AuthErrorValidation, msg)
	}

	if ok, msg := utils.ValidateEmail(email); !ok {
		return httpx.NewAuthError(httpx.AuthErrorValidation, msg)
	}

	if !c.dbConfig.GetBool(consts.ConfigAllowRegister) {
		return httpx.NewAuthError(httpx.AuthErrorForbidden, "注册功能已关闭")
	}

	usernameTaken, err := c.userService.IsUsernameTaken(username, nil, true)
	if err != nil {
		return httpx.NewAuthError(httpx.AuthErrorInternal, "注册失败，请稍后重试")
	}
	if usernameTaken {
		return httpx.NewAuthError(httpx.AuthErrorConflict, "用户名已存在")
	}

	emailTaken, err := c.userService.IsEmailTaken(email, nil, true)
	if err != nil {
		return httpx.NewAuthError(httpx.AuthErrorInternal, "注册失败，请稍后重试")
	}
	if emailTaken {
		return httpx.NewAuthError(httpx.AuthErrorConflict, "邮箱已被注册")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return httpx.NewAuthError(httpx.AuthErrorInternal, "密码加密失败")
	}

	newUser := model.User{
		Username:      username,
		Password:      string(hashedPassword),
		Email:         email,
		EmailVerified: false,
		Admin:         false,
		Avatar:        "",
	}

	if err := c.userService.CreateUser(&newUser); err != nil {
		return httpx.NewAuthError(httpx.AuthErrorInternal, "注册失败，请稍后重试")
	}

	verifyToken, err := utils.GenerateEmailToken(newUser.ID, newUser.Email, 30*time.Minute)
	if err != nil {
		return httpx.NewAuthError(httpx.AuthErrorInternal, "注册失败，请稍后重试")
	}

	baseURL := c.dbConfig.GetString(consts.ConfigBaseURL)
	if baseURL == "" {
		baseURL = "http://localhost"
	}
	if len(baseURL) > 0 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}

	verifyURL := fmt.Sprintf("%s/auth/email-verify?token=%s", baseURL, verifyToken)
	if c.emailService.ShouldSendEmail() {
		go func() {
			_ = c.emailService.SendVerificationEmail(newUser.Email, newUser.Username, verifyURL)
		}()
	}

	return nil
}

// VerifyEmail 验证邮箱激活令牌。
// 返回值第一个参数为 true 表示该邮箱已是验证状态。
func (c *AuthUseCase) VerifyEmail(token string) (bool, error) {
	claims, err := utils.ParseEmailToken(token)
	if err != nil {
		return false, httpx.NewAuthError(httpx.AuthErrorValidation, "验证链接已失效或不正确")
	}

	if claims.Type != "email_verify" {
		return false, httpx.NewAuthError(httpx.AuthErrorValidation, "无效的验证 Token 类型")
	}

	user, err := c.userStore.FindByID(claims.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, httpx.NewAuthError(httpx.AuthErrorNotFound, "用户不存在")
		}
		return false, httpx.NewAuthError(httpx.AuthErrorInternal, "验证失败，请稍后重试")
	}

	if user.Email != claims.Email {
		return false, httpx.NewAuthError(httpx.AuthErrorValidation, "邮箱不匹配，请重新发起验证")
	}

	if user.EmailVerified {
		return true, nil
	}

	user.EmailVerified = true
	if err := c.userService.SaveUser(user); err != nil {
		return false, httpx.NewAuthError(httpx.AuthErrorInternal, "验证失败，请稍后重试")
	}

	return false, nil
}

// VerifyEmailChange 验证邮箱变更令牌并更新邮箱。
func (c *AuthUseCase) VerifyEmailChange(token string) error {
	payload, ok := c.userService.VerifyEmailChangeToken(token)
	if !ok {
		return httpx.NewAuthError(httpx.AuthErrorValidation, "验证链接已失效或不正确")
	}

	user, err := c.userStore.FindByID(payload.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return httpx.NewAuthError(httpx.AuthErrorNotFound, "用户不存在")
		}
		return httpx.NewAuthError(httpx.AuthErrorInternal, "邮箱修改失败，请稍后重试")
	}

	if user.Email != payload.OldEmail {
		return httpx.NewAuthError(httpx.AuthErrorValidation, "您的当前邮箱已变更，该验证链接已失效")
	}

	excludeID := payload.UserID
	emailTaken, err := c.userService.IsEmailTaken(payload.NewEmail, &excludeID, true)
	if err != nil {
		return httpx.NewAuthError(httpx.AuthErrorInternal, "邮箱修改失败，请稍后重试")
	}
	if emailTaken {
		return httpx.NewAuthError(httpx.AuthErrorConflict, "新邮箱已被其他用户占用，无法修改")
	}

	user.Email = payload.NewEmail
	user.EmailVerified = true
	if err := c.userService.SaveUser(user); err != nil {
		return httpx.NewAuthError(httpx.AuthErrorInternal, "邮箱修改失败，请稍后重试")
	}

	return nil
}

// RequestPasswordReset 发起忘记密码流程并异步发送重置邮件。
func (c *AuthUseCase) RequestPasswordReset(email string) error {
	user, err := c.userStore.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return httpx.NewAuthError(httpx.AuthErrorInternal, "生成重置链接失败，请稍后重试")
	}

	if user.Status == 2 || user.Status == 3 {
		return httpx.NewAuthError(httpx.AuthErrorForbidden, "该账号已被封禁或停用，无法重置密码")
	}

	token, err := c.userService.GenerateForgetPasswordToken(user.ID)
	if err != nil {
		return httpx.NewAuthError(httpx.AuthErrorInternal, "生成重置链接失败，请稍后重试")
	}

	baseURL := c.dbConfig.GetString(consts.ConfigBaseURL)
	if baseURL == "" {
		baseURL = "http://localhost"
	}
	if len(baseURL) > 0 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}
	resetURL := fmt.Sprintf("%s/auth/reset-password?token=%s", baseURL, token)

	if c.emailService.ShouldSendEmail() {
		go func() {
			_ = c.emailService.SendPasswordResetEmail(user.Email, user.Username, resetURL)
		}()
	}

	return nil
}

// ResetPassword 使用重置令牌设置新密码。
func (c *AuthUseCase) ResetPassword(token, newPassword string) error {
	if ok, msg := utils.ValidatePassword(newPassword); !ok {
		return httpx.NewAuthError(httpx.AuthErrorValidation, msg)
	}

	userID, valid := c.userService.VerifyForgetPasswordToken(token)
	if !valid {
		return httpx.NewAuthError(httpx.AuthErrorValidation, "重置链接无效或已过期")
	}

	user, err := c.userStore.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return httpx.NewAuthError(httpx.AuthErrorNotFound, "用户不存在")
		}
		return httpx.NewAuthError(httpx.AuthErrorInternal, "密码重置失败")
	}

	if user.Status == 2 || user.Status == 3 {
		return httpx.NewAuthError(httpx.AuthErrorForbidden, "该账号已被封禁或停用")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return httpx.NewAuthError(httpx.AuthErrorInternal, "密码加密失败")
	}

	user.Password = string(hashedPassword)
	user.EmailVerified = true

	if err := c.userService.SaveUser(user); err != nil {
		return httpx.NewAuthError(httpx.AuthErrorInternal, "密码重置失败")
	}

	return nil
}
