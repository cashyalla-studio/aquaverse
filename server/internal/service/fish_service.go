package service

import (
	"context"
	"fmt"

	"github.com/cashyalla/aquaverse/internal/domain"
)

// FishFilter = domain.FishFilter
// FishListResult = domain.FishListResult

type FishRepository interface {
	List(ctx context.Context, filter domain.FishFilter) ([]domain.FishListResponse, int, error)
	GetByID(ctx context.Context, id int64, locale domain.Locale) (*domain.FishData, error)
	Search(ctx context.Context, query string, locale domain.Locale) ([]domain.FishListResponse, error)
	ListFamilies(ctx context.Context) ([]string, error)
	ListCategories(ctx context.Context) ([]domain.CreatureCategoryInfo, error)
}

type FishService struct {
	repo  FishRepository
	cache FishCachePort
}

type FishCachePort interface {
	GetFish(ctx context.Context, key string) (*domain.FishData, error)
	SetFish(ctx context.Context, key string, fish *domain.FishData) error
	GetFishList(ctx context.Context, key string) (*domain.FishListResult, error)
	SetFishList(ctx context.Context, key string, result *domain.FishListResult) error
}

func NewFishService(repo FishRepository, cache FishCachePort) *FishService {
	return &FishService{repo: repo, cache: cache}
}

func (s *FishService) List(ctx context.Context, filter domain.FishFilter) (*domain.FishListResult, error) {
	cacheKey := fmt.Sprintf("fish:list:%s:%s:%s:%s:%s:%d:%d",
		filter.Locale, filter.Category, filter.Family, filter.CareLevel, filter.Search, filter.Page, filter.Limit)

	if cached, err := s.cache.GetFishList(ctx, cacheKey); err == nil {
		return cached, nil
	}

	items, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	result := &domain.FishListResult{
		Items:      items,
		TotalCount: total,
		Page:       filter.Page,
		Limit:      filter.Limit,
	}
	_ = s.cache.SetFishList(ctx, cacheKey, result)
	return result, nil
}

func (s *FishService) GetByID(ctx context.Context, id int64, locale domain.Locale) (*domain.FishData, error) {
	cacheKey := fmt.Sprintf("fish:%d:%s", id, locale)

	if cached, err := s.cache.GetFish(ctx, cacheKey); err == nil {
		return cached, nil
	}

	fish, err := s.repo.GetByID(ctx, id, locale)
	if err != nil {
		return nil, err
	}
	_ = s.cache.SetFish(ctx, cacheKey, fish)
	return fish, nil
}

func (s *FishService) Search(ctx context.Context, query string, locale domain.Locale) ([]domain.FishListResponse, error) {
	return s.repo.Search(ctx, query, locale)
}

func (s *FishService) ListFamilies(ctx context.Context) ([]string, error) {
	return s.repo.ListFamilies(ctx)
}

func (s *FishService) ListCategories(ctx context.Context) ([]domain.CreatureCategoryInfo, error) {
	return s.repo.ListCategories(ctx)
}
