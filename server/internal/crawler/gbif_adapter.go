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
	gbifAPIBase      = "https://api.gbif.org/v1"
	gbifAnimaliaKey  = 1   // kingdom Animalia GBIF taxon key
	gbifPageSize     = 300
	gbifRateDelay    = 500 * time.Millisecond
)

// GBIFAdapter GBIF Species API SourceAdapter 구현체.
// CC0 라이선스 데이터를 제공한다.
// rate limit: 요청 간 500ms 딜레이.
type GBIFAdapter struct {
	client    *http.Client
	logger    *slog.Logger
	userAgent string
}

// gbifSpeciesResult GBIF /species/search 응답 항목
type gbifSpeciesResult struct {
	Key             int                `json:"key"`
	ScientificName  string             `json:"scientificName"`
	CanonicalName   string             `json:"canonicalName"`
	Genus           string             `json:"genus"`
	SpecificEpithet string             `json:"specificEpithet"`
	Family          string             `json:"family"`
	Order           string             `json:"order"`
	Class           string             `json:"class"`
	Kingdom         string             `json:"kingdom"`
	VernacularNames []gbifVernacularName `json:"vernacularNames"`
}

// gbifVernacularName GBIF 다국어 일반명 항목
type gbifVernacularName struct {
	VernacularName string `json:"vernacularName"`
	Language       string `json:"language"`
}

// gbifSearchResponse GBIF /species/search 응답 래퍼
type gbifSearchResponse struct {
	Results      []gbifSpeciesResult `json:"results"`
	EndOfRecords bool                `json:"endOfRecords"`
}

// NewGBIFAdapter GBIFAdapter 생성자
func NewGBIFAdapter(logger *slog.Logger, userAgent string) *GBIFAdapter {
	return &GBIFAdapter{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:    logger,
		userAgent: userAgent,
	}
}

// Name 소스 식별자 반환
func (a *GBIFAdapter) Name() string {
	return "gbif"
}

// Crawl kingdom Animalia의 SPECIES 랭크 분류군을 페이지 단위로 순회해 저장한다.
// API 연결 실패 시 (0, nil)을 반환해 graceful degradation을 지원한다.
func (a *GBIFAdapter) Crawl(ctx context.Context, jobID int64, save func(*domain.RawCrawlData) error) (int, error) {
	fetched := 0
	offset := 0

	for {
		resp, err := a.fetchPage(ctx, offset, gbifPageSize)
		if err != nil {
			a.logger.Warn("gbif_adapter: fetch page failed, stopping crawl",
				"offset", offset, "err", err)
			// graceful degradation
			return fetched, nil
		}

		for _, result := range resp.Results {
			rawData, err := a.toRawCrawlData(jobID, result)
			if err != nil {
				a.logger.Warn("gbif_adapter: marshal failed",
					"key", result.Key, "err", err)
				continue
			}
			if err := save(rawData); err != nil {
				a.logger.Error("gbif_adapter: save failed",
					"key", result.Key, "err", err)
				continue
			}
			fetched++
		}

		if resp.EndOfRecords || len(resp.Results) == 0 {
			break
		}

		offset += gbifPageSize

		// rate limiting: 500ms 딜레이
		select {
		case <-ctx.Done():
			return fetched, ctx.Err()
		case <-time.After(gbifRateDelay):
		}
	}

	return fetched, nil
}

// fetchPage GBIF /species/search 엔드포인트 호출
func (a *GBIFAdapter) fetchPage(ctx context.Context, offset, limit int) (*gbifSearchResponse, error) {
	url := fmt.Sprintf(
		"%s/species/search?rank=SPECIES&highertaxonKey=%d&limit=%d&offset=%d",
		gbifAPIBase, gbifAnimaliaKey, limit, offset,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("gbif_adapter: new request: %w", err)
	}
	req.Header.Set("User-Agent", a.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gbif_adapter: http do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gbif_adapter: unexpected status %d for %s", resp.StatusCode, url)
	}

	var result gbifSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("gbif_adapter: decode response: %w", err)
	}

	return &result, nil
}

// gbifClassToCategory GBIF class 필드를 CreatureCategory로 변환한다.
func gbifClassToCategory(class string) string {
	switch class {
	case "Actinopterygii", "Chondrichthyes", "Sarcopterygii":
		return string(domain.CategoryFish)
	case "Reptilia":
		return string(domain.CategoryReptile)
	case "Amphibia":
		return string(domain.CategoryAmphibian)
	case "Insecta":
		return string(domain.CategoryInsect)
	case "Arachnida":
		return string(domain.CategoryArachnid)
	case "Aves":
		return string(domain.CategoryBird)
	case "Mammalia":
		return string(domain.CategoryMammal)
	default:
		// 분류 불명: 기본값 "fish"로 처리 (파이프라인에서 AI 보완 가능)
		return string(domain.CategoryFish)
	}
}

// gbifPrimaryCommonName vernacularNames 목록에서 영어(eng) 일반명을 추출한다.
func gbifPrimaryCommonName(names []gbifVernacularName) string {
	for _, vn := range names {
		if vn.Language == "eng" && vn.VernacularName != "" {
			return vn.VernacularName
		}
	}
	return ""
}

// toRawCrawlData GBIF 응답 항목을 domain.FishData로 변환한 뒤 RawCrawlData로 래핑한다.
func (a *GBIFAdapter) toRawCrawlData(jobID int64, result gbifSpeciesResult) (*domain.RawCrawlData, error) {
	// canonicalName 우선, 없으면 scientificName 사용
	scientificName := result.CanonicalName
	if scientificName == "" {
		scientificName = result.ScientificName
	}

	sourceURL := fmt.Sprintf("https://www.gbif.org/species/%d", result.Key)
	license := "CC0"
	attribution := "GBIF"
	category := gbifClassToCategory(result.Class)
	commonName := gbifPrimaryCommonName(result.VernacularNames)

	fish := domain.FishData{
		ScientificName:    scientificName,
		Genus:             result.Genus,
		Species:           result.SpecificEpithet,
		Family:            result.Family,
		OrderName:         result.Order,
		ClassName:         result.Class,
		PrimaryCommonName: commonName,
		PrimarySource:     "gbif",
		SourceURL:         &sourceURL,
		License:           &license,
		Attribution:       &attribution,
		CreatureCategory:  category,
		PublishStatus:     domain.PublishStatusDraft,
	}

	rawJSON, err := json.Marshal(fish)
	if err != nil {
		return nil, fmt.Errorf("gbif_adapter: marshal fish data: %w", err)
	}

	jobIDCopy := jobID
	sourceIDStr := fmt.Sprintf("%d", result.Key)
	return &domain.RawCrawlData{
		CrawlJobID:  &jobIDCopy,
		SourceName:  "gbif",
		SourceID:    sourceIDStr,
		SourceURL:   &sourceURL,
		RawJSON:     rawJSON,
		ContentHash: hashContent(rawJSON),
		ParseStatus: "PENDING",
	}, nil
}
