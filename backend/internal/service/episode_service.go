package service

import (
	"github.com/2hot4you/aiaos/backend/internal/domain"
	"github.com/2hot4you/aiaos/backend/internal/repository/postgres"
	"github.com/2hot4you/aiaos/backend/pkg/snowflake"
	"gorm.io/gorm"
)

type EpisodeService struct {
	repo       *postgres.EpisodeRepo
	seasonRepo *postgres.SeasonRepo
}

func NewEpisodeService(repo *postgres.EpisodeRepo, seasonRepo *postgres.SeasonRepo) *EpisodeService {
	return &EpisodeService{repo: repo, seasonRepo: seasonRepo}
}

func (s *EpisodeService) ListBySeason(seasonID int64) ([]domain.Episode, error) {
	if _, err := s.seasonRepo.FindByID(seasonID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrSeasonNotFound
		}
		return nil, err
	}
	return s.repo.ListBySeason(seasonID)
}

func (s *EpisodeService) Create(seasonID int64, title string, createdBy int64) (*domain.Episode, error) {
	if _, err := s.seasonRepo.FindByID(seasonID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrSeasonNotFound
		}
		return nil, err
	}

	maxOrder, _ := s.repo.GetMaxSortOrder(seasonID)

	ep := &domain.Episode{
		ID:        snowflake.Generate(),
		SeasonID:  seasonID,
		Title:     title,
		SortOrder: maxOrder + 1,
		CreatedBy: createdBy,
	}

	if err := s.repo.Create(ep); err != nil {
		return nil, err
	}
	return ep, nil
}

func (s *EpisodeService) GetByID(id int64) (*domain.Episode, error) {
	ep, err := s.repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrEpisodeNotFound
		}
		return nil, err
	}
	return ep, nil
}

func (s *EpisodeService) Update(id int64, title string) (*domain.Episode, error) {
	ep, err := s.repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrEpisodeNotFound
		}
		return nil, err
	}

	if title != "" {
		ep.Title = title
	}

	if err := s.repo.Update(ep); err != nil {
		return nil, err
	}
	return ep, nil
}

func (s *EpisodeService) Delete(id int64) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if err == gorm.ErrRecordNotFound {
			return domain.ErrEpisodeNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}
