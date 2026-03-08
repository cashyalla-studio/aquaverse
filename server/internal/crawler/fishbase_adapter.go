package crawler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/cashyalla/aquaverse/internal/domain"
)

const (
	fishbaseAdapterAPIBase  = "https://fishbase.ropensci.org"
	fishbaseAdapterImageBase = "https://www.fishbase.se/images/species"
	fishbaseAdapterSummaryBase = "https://www.fishbase.se/summary"
	fishbaseAdapterPageSize = 100
)

// FishBaseAdapter FishBase rOpenSci REST API SourceAdapter 구현체.
// CC-BY 라이선스 데이터를 제공한다.
// rate limit: 요청 간 1초 딜레이.
type FishBaseAdapter struct {
	client    *http.Client
	logger    *slog.Logger
	userAgent string
}

// fishBaseAPISpecies FishBase /species 엔드포인트 응답 항목
type fishBaseAPISpecies struct {
	SpecCode         int     `json:"SpecCode"`
	Genus            string  `json:"Genus"`
	Species          string  `json:"Species"`
	FBname           string  `json:"FBname"`
	Family           string  `json:"Family"`
	Order            string  `json:"Order"`
	Class            string  `json:"Class"`
	Length           float64 `json:"Length"`
	Longevity        float64 `json:"Longevity"`
	PHMin            float64 `json:"pHMin"`
	PHMax            float64 `json:"pHMax"`
	TempMin          float64 `json:"TempMin"`
	TempMax          float64 `json:"TempMax"`
	PicPreferredName string  `json:"PicPreferredName"`
}

// fishBaseListResponse /species 목록 응답 래퍼
type fishBaseListResponse struct {
	Data []fishBaseAPISpecies `json:"data"`
}

// NewFishBaseAdapter FishBaseAdapter 생성자
func NewFishBaseAdapter(logger *slog.Logger, userAgent string) *FishBaseAdapter {
	return &FishBaseAdapter{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:    logger,
		userAgent: userAgent,
	}
}

// Name 소스 식별자 반환
func (a *FishBaseAdapter) Name() string {
	return "fishbase"
}

// Crawl 전체 어종 목록을 페이지 단위로 순회하며 RawCrawlData를 저장한다.
// API 연결 실패 시 (0, nil) 을 반환해 graceful degradation을 지원한다.
func (a *FishBaseAdapter) Crawl(ctx context.Context, jobID int64, save func(*domain.RawCrawlData) error) (int, error) {
	fetched := 0
	offset := 0

	for {
		species, err := a.fetchPage(ctx, offset, fishbaseAdapterPageSize)
		if err != nil {
			a.logger.Warn("fishbase_adapter: fetch page failed, stopping crawl",
				"offset", offset, "err", err)
			// graceful degradation: 연결 실패 시 지금까지 수집한 건수와 nil 반환
			return fetched, nil
		}
		if len(species) == 0 {
			break
		}

		for _, sp := range species {
			rawData, err := a.toRawCrawlData(jobID, sp)
			if err != nil {
				a.logger.Warn("fishbase_adapter: marshal failed",
					"spec_code", sp.SpecCode, "err", err)
				continue
			}
			if err := save(rawData); err != nil {
				a.logger.Error("fishbase_adapter: save failed",
					"spec_code", sp.SpecCode, "err", err)
				continue
			}
			fetched++
		}

		if len(species) < fishbaseAdapterPageSize {
			break
		}

		offset += fishbaseAdapterPageSize

		// rate limiting: 1초 딜레이
		select {
		case <-ctx.Done():
			return fetched, ctx.Err()
		case <-time.After(time.Second):
		}
	}

	return fetched, nil
}

// fetchPage /species?limit=N&offset=M 엔드포인트 호출
func (a *FishBaseAdapter) fetchPage(ctx context.Context, offset, limit int) ([]fishBaseAPISpecies, error) {
	url := fmt.Sprintf("%s/species?limit=%d&offset=%d", fishbaseAdapterAPIBase, limit, offset)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("fishbase_adapter: new request: %w", err)
	}
	req.Header.Set("User-Agent", a.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fishbase_adapter: http do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fishbase_adapter: unexpected status %d for %s", resp.StatusCode, url)
	}

	var result fishBaseListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("fishbase_adapter: decode response: %w", err)
	}

	return result.Data, nil
}

// toRawCrawlData FishBase 응답 항목을 domain.FishData로 변환한 뒤 RawCrawlData로 래핑한다.
// RawJSON은 domain.FishData를 JSON 직렬화한 값이어야 pipeline.parseCrawlData()에서 언마샬 가능하다.
func (a *FishBaseAdapter) toRawCrawlData(jobID int64, sp fishBaseAPISpecies) (*domain.RawCrawlData, error) {
	scientificName := sp.Genus + " " + sp.Species
	sourceURL := fmt.Sprintf("%s/%s-%s.html", fishbaseAdapterSummaryBase, sp.Genus, sp.Species)
	license := "CC BY"
	attribution := "FishBase"

	fish := domain.FishData{
		ScientificName:    scientificName,
		Genus:             sp.Genus,
		Species:           sp.Species,
		Family:            sp.Family,
		OrderName:         sp.Order,
		ClassName:         sp.Class,
		PrimaryCommonName: sp.FBname,
		PrimarySource:     "fishbase",
		SourceURL:         &sourceURL,
		License:           &license,
		Attribution:       &attribution,
		CreatureCategory:  string(domain.CategoryFish),
		PublishStatus:     domain.PublishStatusDraft,
	}

	// 수치 필드: 0 값은 미설정으로 간주하고 nil 유지
	if sp.Length > 0 {
		fish.MaxSizeCm = &sp.Length
	}
	if sp.Longevity > 0 {
		fish.LifespanYears = &sp.Longevity
	}
	if sp.PHMin > 0 {
		fish.PHMin = &sp.PHMin
	}
	if sp.PHMax > 0 {
		fish.PHMax = &sp.PHMax
	}
	if sp.TempMin > 0 {
		fish.TempMinC = &sp.TempMin
	}
	if sp.TempMax > 0 {
		fish.TempMaxC = &sp.TempMax
	}

	// 이미지 URL
	if sp.PicPreferredName != "" {
		imageURL := fmt.Sprintf("%s/%s", fishbaseAdapterImageBase, sp.PicPreferredName)
		fish.PrimaryImageURL = &imageURL
		imageAttribution := "FishBase"
		imageLicense := "CC BY"
		fish.ImageLicense = &imageLicense
		fish.ImageAuthor = &imageAttribution
	}

	rawJSON, err := json.Marshal(fish)
	if err != nil {
		return nil, fmt.Errorf("fishbase_adapter: marshal fish data: %w", err)
	}

	jobIDCopy := jobID
	return &domain.RawCrawlData{
		CrawlJobID:  &jobIDCopy,
		SourceName:  "fishbase",
		SourceID:    fmt.Sprintf("%d", sp.SpecCode),
		SourceURL:   &sourceURL,
		RawJSON:     rawJSON,
		ContentHash: hashContent(rawJSON),
		ParseStatus: "PENDING",
	}, nil
}
