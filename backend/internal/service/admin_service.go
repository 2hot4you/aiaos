package service

import (
	"github.com/2hot4you/aiaos/backend/internal/domain"
	"github.com/2hot4you/aiaos/backend/internal/repository/postgres"
	"github.com/2hot4you/aiaos/backend/pkg/config"
	"github.com/2hot4you/aiaos/backend/pkg/crypto"
	"github.com/2hot4you/aiaos/backend/pkg/snowflake"
	"gorm.io/gorm"
)

type AdminService struct {
	userRepo  *postgres.UserRepo
	modelRepo *postgres.ModelConfigRepo
	encCfg    config.EncryptionConfig
}

func NewAdminService(userRepo *postgres.UserRepo, modelRepo *postgres.ModelConfigRepo, encCfg config.EncryptionConfig) *AdminService {
	return &AdminService{
		userRepo:  userRepo,
		modelRepo: modelRepo,
		encCfg:    encCfg,
	}
}

// ---- User Management ----

func (s *AdminService) ListUsers(page, pageSize int, keyword string) ([]domain.User, int64, error) {
	return s.userRepo.List(page, pageSize, keyword)
}

func (s *AdminService) CreateUser(username, displayName, password, role string) (*domain.User, error) {
	// Check duplicate
	if existing, _ := s.userRepo.FindByUsername(username); existing != nil {
		return nil, domain.ErrUsernameExists
	}

	hash, err := crypto.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:           snowflake.Generate(),
		Username:     username,
		DisplayName:  displayName,
		PasswordHash: hash,
		Role:         role,
		Enabled:      true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AdminService) UpdateUser(id int64, displayName, role string) (*domain.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	if displayName != "" {
		user.DisplayName = displayName
	}
	if role != "" {
		user.Role = role
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AdminService) DeleteUser(id int64) error {
	return s.userRepo.SoftDelete(id)
}

func (s *AdminService) ResetPassword(id int64, newPassword string) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return domain.ErrUserNotFound
		}
		return err
	}

	hash, err := crypto.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = hash
	return s.userRepo.Update(user)
}

func (s *AdminService) UpdateUserStatus(id int64, enabled bool) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return domain.ErrUserNotFound
		}
		return err
	}

	user.Enabled = enabled
	return s.userRepo.Update(user)
}

// ---- Model Management ----

func (s *AdminService) ListModels() ([]domain.AIModelConfig, error) {
	return s.modelRepo.List()
}

func (s *AdminService) CreateModel(m *domain.AIModelConfig, apiKey string) (*domain.AIModelConfig, error) {
	m.ID = snowflake.Generate()

	enc, err := crypto.EncryptAES(apiKey, s.encCfg.Key)
	if err != nil {
		return nil, err
	}
	m.APIKeyEnc = enc

	// If is_default, clear other defaults of same type
	if m.IsDefault {
		_ = s.modelRepo.ClearDefault(m.ModelType)
	}

	if err := s.modelRepo.Create(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *AdminService) UpdateModel(id int64, updates map[string]interface{}, apiKey string) (*domain.AIModelConfig, error) {
	m, err := s.modelRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrModelNotFound
		}
		return nil, err
	}

	if v, ok := updates["name"].(string); ok && v != "" {
		m.Name = v
	}
	if v, ok := updates["model_type"].(string); ok && v != "" {
		m.ModelType = v
	}
	if v, ok := updates["provider"].(string); ok && v != "" {
		m.Provider = v
	}
	if v, ok := updates["endpoint"].(string); ok && v != "" {
		m.Endpoint = v
	}
	if v, ok := updates["model_identifier"].(string); ok && v != "" {
		m.ModelIdentifier = v
	}
	if v, ok := updates["max_concurrency"].(float64); ok {
		m.MaxConcurrency = int(v)
	}
	if v, ok := updates["timeout_seconds"].(float64); ok {
		m.TimeoutSeconds = int(v)
	}
	if v, ok := updates["is_default"].(bool); ok {
		if v {
			_ = s.modelRepo.ClearDefault(m.ModelType)
		}
		m.IsDefault = v
	}
	if v, ok := updates["enabled"].(bool); ok {
		m.Enabled = v
	}

	if apiKey != "" {
		enc, err := crypto.EncryptAES(apiKey, s.encCfg.Key)
		if err != nil {
			return nil, err
		}
		m.APIKeyEnc = enc
	}

	if err := s.modelRepo.Update(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *AdminService) DeleteModel(id int64) error {
	return s.modelRepo.Delete(id)
}
