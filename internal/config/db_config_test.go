package config

import (
	"errors"
	"testing"

	"perfect-pic-server/internal/consts"
	"perfect-pic-server/internal/model"
	repo "perfect-pic-server/internal/repository"

	"gorm.io/gorm"
)

type mockSettingStore struct {
	findByKeyFn           func(key string) (*model.Setting, error)
	createFn              func(setting *model.Setting) error
	findByKeyCallCount    int
	createCallCount       int
	lastCreatedSettingKey string
}

func (m *mockSettingStore) InitializeDefaults(defaults []model.Setting) error {
	return nil
}

func (m *mockSettingStore) DeleteNotInKeys(allowedKeys []string) error {
	return nil
}

func (m *mockSettingStore) FindByKey(key string) (*model.Setting, error) {
	m.findByKeyCallCount++
	if m.findByKeyFn != nil {
		return m.findByKeyFn(key)
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockSettingStore) Create(setting *model.Setting) error {
	m.createCallCount++
	m.lastCreatedSettingKey = setting.Key
	if m.createFn != nil {
		return m.createFn(setting)
	}
	return nil
}

func (m *mockSettingStore) FindAll() ([]model.Setting, error) {
	return nil, nil
}

func (m *mockSettingStore) UpdateSettings(items []repo.UpdateSettingItem, maskedValue string) error {
	return nil
}

func TestDBConfigGetString_RecordNotFoundUsesDefaultValue(t *testing.T) {
	store := &mockSettingStore{
		findByKeyFn: func(key string) (*model.Setting, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	dbConfig := NewDBConfig(store)

	got := dbConfig.GetString(consts.ConfigSiteName)
	if got != "Perfect Pic" {
		t.Fatalf("expected default value %q, got %q", "Perfect Pic", got)
	}
	if store.createCallCount != 1 {
		t.Fatalf("expected create called once, got %d", store.createCallCount)
	}
	if store.lastCreatedSettingKey != consts.ConfigSiteName {
		t.Fatalf("expected created key %q, got %q", consts.ConfigSiteName, store.lastCreatedSettingKey)
	}

	_ = dbConfig.GetString(consts.ConfigSiteName)
	if store.findByKeyCallCount != 1 {
		t.Fatalf("expected cached read after first call, findByKey called %d times", store.findByKeyCallCount)
	}
}

func TestDBConfigGetString_NonNotFoundErrorShouldNotCreateDefault(t *testing.T) {
	store := &mockSettingStore{
		findByKeyFn: func(key string) (*model.Setting, error) {
			return nil, errors.New("database temporarily unavailable")
		},
	}
	dbConfig := NewDBConfig(store)

	got := dbConfig.GetString(consts.ConfigSiteName)
	if got != "" {
		t.Fatalf("expected empty string on db error, got %q", got)
	}
	if store.createCallCount != 0 {
		t.Fatalf("expected create not called on db error, got %d", store.createCallCount)
	}

	_ = dbConfig.GetString(consts.ConfigSiteName)
	if store.findByKeyCallCount != 2 {
		t.Fatalf("expected db error not cached, findByKey called %d times", store.findByKeyCallCount)
	}
}
