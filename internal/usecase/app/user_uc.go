package app

import (
	"fmt"
	commonpkg "perfect-pic-server/internal/common"
	"perfect-pic-server/internal/consts"
	moduledto "perfect-pic-server/internal/dto"
	"perfect-pic-server/internal/pkg/validator"

	"golang.org/x/crypto/bcrypt"
)

// RequestEmailChange 发起邮箱修改流程并异步发送验证邮件。
func (c *UserUseCase) RequestEmailChange(userID uint, password, newEmail string) error {
	if ok, msg := validator.ValidateEmail(newEmail); !ok {
		return commonpkg.NewValidationError(msg)
	}

	user, err := c.userStore.FindByID(userID)
	if err != nil {
		return commonpkg.NewInternalError("用户不存在")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return commonpkg.NewForbiddenError("密码错误")
	}

	if user.Email == newEmail {
		return commonpkg.NewValidationError("新邮箱不能与当前邮箱相同")
	}

	emailTaken, err := c.userService.IsEmailTaken(newEmail, nil, true)
	if err != nil {
		return commonpkg.NewInternalError("检查邮箱占用失败")
	}
	if emailTaken {
		return commonpkg.NewConflictError("该邮箱已被使用")
	}

	token, err := c.userService.GenerateEmailChangeToken(user.ID, user.Email, newEmail)
	if err != nil {
		return commonpkg.NewInternalError("生成验证链接失败")
	}

	baseURL := c.dbConfig.GetString(consts.ConfigBaseURL)
	if baseURL == "" {
		baseURL = "http://localhost"
	}
	if len(baseURL) > 0 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}
	verifyURL := fmt.Sprintf("%s/auth/email-change-verify?token=%s", baseURL, token)

	if c.emailService.ShouldSendEmail() {
		go func() {
			_ = c.emailService.SendEmailChangeVerification(newEmail, user.Username, user.Email, newEmail, verifyURL)
		}()
	}

	return nil
}

// UpdateUsernameAndGenerateToken 修改用户名并签发新登录令牌。
func (c *UserUseCase) UpdateUsernameAndGenerateToken(userID uint, username string) (string, error) {
	if err := c.userService.UpdateUser(userID, moduledto.UpdateUserRequest{Username: &username}, false); err != nil {
		return "", err
	}
	user, err := c.userStore.FindByID(userID)
	if err != nil {
		return "", commonpkg.NewInternalError("更新失败")
	}
	return c.authService.IssueLoginToken(user)
}

// UpdatePasswordByOldPassword 使用旧密码校验后更新密码。
func (c *UserUseCase) UpdatePasswordByOldPassword(userID uint, oldPassword, newPassword string) error {
	return c.userService.UpdatePasswordByOldPassword(userID, oldPassword, newPassword)
}
