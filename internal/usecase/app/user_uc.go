package app

import (
	"fmt"
	commonpkg "perfect-pic-server/internal/common"
	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/utils"

	"golang.org/x/crypto/bcrypt"
)

// RequestEmailChange 发起邮箱修改流程并异步发送验证邮件。
func (c *UserUseCase) RequestEmailChange(userID uint, password, newEmail string) error {
	if ok, msg := utils.ValidateEmail(newEmail); !ok {
		return commonpkg.NewValidationError(msg)
	}

	user, err := c.userStore.FindByID(userID)
	if err != nil {
		return commonpkg.NewNotFoundError("用户不存在")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return commonpkg.NewForbiddenError("密码错误")
	}

	if user.Email == newEmail {
		return commonpkg.NewValidationError("新邮箱不能与当前邮箱相同")
	}

	emailTaken, err := c.userService.IsEmailTaken(newEmail, nil, true)
	if err != nil {
		return commonpkg.NewInternalError("生成验证链接失败")
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