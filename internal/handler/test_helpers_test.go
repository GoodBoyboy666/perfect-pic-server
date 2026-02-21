package handler

import (
	"testing"

	"perfect-pic-server/internal/repository"
	"perfect-pic-server/internal/service"
	"perfect-pic-server/internal/testutils"

	"gorm.io/gorm"
)

var (
	testService *service.AppService
	testHandler *Handler
)

func setupTestDB(t *testing.T) *gorm.DB {
	gdb := testutils.SetupDB(t)
	testService = service.NewAppService(repository.NewRepositories(
		repository.NewUserRepository(gdb),
		repository.NewImageRepository(gdb),
		repository.NewSettingRepository(gdb),
		repository.NewSystemRepository(gdb),
	))
	testHandler = NewHandler(testService)
	testService.ClearCache()
	return gdb
}
