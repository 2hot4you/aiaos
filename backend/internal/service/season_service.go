package service

import (
	"github.com/2hot4you/aiaos/backend/internal/domain"
	"github.com/2hot4you/aiaos/backend/internal/repository/postgres"
	"github.com/2hot4you/aiaos/backend/pkg/snowflake"
	"gorm.io/gorm"
)

type SeasonService struct {
	repo        *postgres.SeasonRepo
	projectRepo *postgres.ProjectRepo
}

func NewSeasonService(repo *postgres.SeasonRepo, projectRepo *postgres.ProjectRepo) *SeasonService {
	return &SeasonService{repo: repo, projectRepo: projectRepo}
}

func (s *SeasonService) ListByProject(projectID int64) ([]domain.Season, error) {
	// Verify project exists
	if _, err := s.projectRepo.FindByID(projectID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrProjectNotFound
		}
		return nil, err
	}
	return s.repo.ListByProject(projectID)
}

func (s *SeasonService) Create(projectID int64, title string, createdBy int64) (*domain.Season, error) {
	// Verify project exists
	if _, err := s.projectRepo.FindByID(projectID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrProjectNotFound
		}
		return nil, err
	}

	maxOrder, _ := s.repo.GetMaxSortOrder(projectID)

	season := &domain.Season{
		ID:        snowflake.Generate(),
		ProjectID: projectID,
		Title:     title,
		SortOrder: maxOrder + 1,
		CreatedBy: createdBy,
	}

	if err := s.repo.Create(season); err != nil {
		return nil, err
	}
	return season, nil
}

func (s *SeasonService) Update(id int64, title string) (*domain.Season, error) {
	season, err := s.repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrSeasonNotFound
		}
		return nil, err
	}

	if title != "" {
		season.Title = title
	}

	if err := s.repo.Update(season); err != nil {
		return nil, err
	}
	return season, nil
}

func (s *SeasonService) Delete(id int64) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if err == gorm.ErrRecordNotFound {
			return domain.ErrSeasonNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}
