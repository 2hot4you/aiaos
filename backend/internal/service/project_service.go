package service

import (
	"github.com/2hot4you/aiaos/backend/internal/domain"
	"github.com/2hot4you/aiaos/backend/internal/repository/postgres"
	"github.com/2hot4you/aiaos/backend/pkg/snowflake"
	"gorm.io/gorm"
)

type ProjectService struct {
	repo *postgres.ProjectRepo
}

func NewProjectService(repo *postgres.ProjectRepo) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) List(page, pageSize int, keyword, sortBy, sortOrder string) ([]domain.Project, int64, error) {
	return s.repo.List(page, pageSize, keyword, sortBy, sortOrder)
}

func (s *ProjectService) Create(name string, createdBy int64) (*domain.Project, error) {
	p := &domain.Project{
		ID:        snowflake.Generate(),
		Name:      name,
		CreatedBy: createdBy,
	}
	if err := s.repo.Create(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ProjectService) GetByID(id int64) (*domain.Project, error) {
	p, err := s.repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrProjectNotFound
		}
		return nil, err
	}
	return p, nil
}

func (s *ProjectService) Update(id int64, name string) (*domain.Project, error) {
	p, err := s.repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrProjectNotFound
		}
		return nil, err
	}

	if name != "" {
		p.Name = name
	}

	if err := s.repo.Update(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ProjectService) Delete(id int64) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return domain.ErrProjectNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}
