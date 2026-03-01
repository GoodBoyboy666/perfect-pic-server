package service

import (
	"errors"
	commonpkg "perfect-pic-server/internal/common"
	moduledto "perfect-pic-server/internal/dto"

	"gorm.io/gorm"
)

// ListUserPasskeys 返回指定用户已绑定的 Passkey 列表。
func (s *PasskeyService) ListUserPasskeys(userID uint) ([]moduledto.UserPasskeyResponse, error) {
	//if _, err := s.findUserByID(userID); err != nil {
	//	if errors.Is(err, gorm.ErrRecordNotFound) {
	//		return nil, commonpkg.NewNotFoundError("用户不存在")
	//	}
	//	return nil, commonpkg.NewInternalError("读取用户信息失败")
	//}

	records, err := s.passkeyStore.ListPasskeyCredentialsByUserID(userID)
	if err != nil {
		return nil, commonpkg.NewInternalError("读取 Passkey 列表失败")
	}

	items := make([]moduledto.UserPasskeyResponse, 0, len(records))
	for _, record := range records {
		items = append(items, moduledto.UserPasskeyResponse{
			ID:           record.ID,
			CredentialID: record.CredentialID,
			Name:         record.Name,
			CreatedAt:    record.CreatedAt.Unix(),
		})
	}

	return items, nil
}

// DeleteUserPasskey 删除指定用户名下的某个 Passkey。
func (s *PasskeyService) DeleteUserPasskey(userID uint, passkeyID uint) error {
	if passkeyID == 0 {
		return commonpkg.NewValidationError("无效的 Passkey ID")
	}

	if err := s.passkeyStore.DeletePasskeyCredentialByID(userID, passkeyID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return commonpkg.NewNotFoundError("Passkey 不存在")
		}
		return commonpkg.NewInternalError("删除 Passkey 失败")
	}
	return nil
}

// UpdateUserPasskeyName 更新指定用户名下某个 Passkey 的显示名称。
func (s *PasskeyService) UpdateUserPasskeyName(userID uint, passkeyID uint, name string) error {
	if passkeyID == 0 {
		return commonpkg.NewValidationError("无效的 Passkey ID")
	}

	normalizedName, err := normalizePasskeyName(name)
	if err != nil {
		return err
	}

	if err := s.passkeyStore.UpdatePasskeyCredentialNameByID(userID, passkeyID, normalizedName); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return commonpkg.NewNotFoundError("Passkey 不存在")
		}
		return commonpkg.NewInternalError("更新 Passkey 名称失败")
	}
	return nil
}
