package admin

import (
	"errors"
	commonpkg "perfect-pic-server/internal/common"
	"time"

	"gorm.io/gorm"
)

// AdminDeleteUser 删除用户。
// hardDelete=true 时执行彻底删除；否则执行软删除并清理唯一字段占用。
func (c *UserManageUseCase) AdminDeleteUser(userID uint, hardDelete bool) error {
	if hardDelete {
		if err := c.imageService.DeleteUserFiles(userID); err != nil {
			return commonpkg.NewInternalError("删除用户失败")
		}
		// 显式删除该用户所有 Passkey 凭据，避免旧 SQLite 表外键缺失导致级联删除失效。
		if err := c.passkeyService.DeletePasskeyCredentialsByUserID(userID); err != nil {
			return commonpkg.NewInternalError("删除用户失败")
		}
		if err := c.userService.HardDeleteUserWithImages(userID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return commonpkg.NewNotFoundError("用户不存在")
			}
			return commonpkg.NewInternalError("删除用户失败")
		}
		return nil
	}

	if err := c.userService.SoftDeleteUser(userID, time.Now().Unix()); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return commonpkg.NewNotFoundError("用户不存在")
		}
		return commonpkg.NewInternalError("删除用户失败")
	}
	return nil
}
