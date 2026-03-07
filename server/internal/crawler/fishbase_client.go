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

const fishbaseAPIBase = "https://fishbase.ropensci.org"

// FishBaseClient FishBase REST API 클라이언트 (CC-BY 라이선스)
type FishBaseClient struct {
	client      *http.Client
	rateLimiter <-chan time.Time
	logger      *slog.Logger
	userAgent   string
}

type FishBaseSpecies struct {
	SpecCode            int     `json:"SpecCode"`
	Genus               string  `json:"Genus"`
	Species             string  `json:"Species"`
	FBname              string  `json:"FBname"` // 영어 일반명
	Family              string  `json:"Family"`
	Order               string  `json:"Order"`
	Class               string  `json:"Class"`
	PicPreferredName    string  `json:"PicPreferredName"`
	PHMin               float64 `json:"pHMin"`
	PHMax               float64 `json:"pHMax"`
	TempMin             float64 `json:"TempMin"`
	TempMax             float64 `json:"TempMax"`
	LongevityWild       float64 `json:"LongevityWild"`
	LengthMax           float64 `json:"LengthMax"`
	Dangerous           string  `json:"Dangerous"`
	AquariumNote        string  `json:"AquariumNote"`
	DietTroph           float64 `json:"DietTroph"` // 영양 단계 (2=초식, 3+=육식)
}

type FishBaseEcology struct {
	SpecCode    int    `json:"SpecCode"`
	CoralReefs  int    `json:"CoralReefs"`
	Fresh       int    `json:"Fresh"`
	Brackish    int    `json:"Brackish"`
	Saltwater   int    `json:"Saltwater"`
	Aquarium    string `json:"Aquarium"` // "highly recommended"/"commercial" 등
}

type FishBaseSpawning struct {
	SpecCode    int    `json:"SpecCode"`
	FecundityMin float64 `json:"FecundityMin"`
	FecundityMax float64 `json:"FecundityMax"`
	SpawningGround string `json:"SpawningGround"`
}

func NewFishBaseClient(logger *slog.Logger, userAgent string, reqPerMin int) *FishBaseClient {
	// 분당 reqPerMin 제한
	ticker := time.NewTicker(time.Minute / time.Duration(reqPerMin))
	return &FishBaseClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		rateLimiter: ticker.C,
		logger:      logger,
		userAgent:   userAgent,
	}
}

// ListAquariumSpecies 수조 적합 어종 목록 조회
func (c *FishBaseClient) ListAquariumSpecies(ctx context.Context, offset, limit int) ([]FishBaseSpecies, error) {
	url := fmt.Sprintf("%s/species?aquarium=highly+recommended&offset=%d&limit=%d", fishbaseAPIBase, offset, limit)
	return c.fetchSpecies(ctx, url)
}

// GetSpecies 특정 어종 상세 조회
func (c *FishBaseClient) GetSpecies(ctx context.Context, specCode int) (*FishBaseSpecies, error) {
	url := fmt.Sprintf("%s/species/%d", fishbaseAPIBase, specCode)

	<-c.rateLimiter // rate limit 준수

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fishbase API error: %d", resp.StatusCode)
	}

	var species FishBaseSpecies
	if err := json.NewDecoder(resp.Body).Decode(&species); err != nil {
		return nil, err
	}
	return &species, nil
}

// GetEcology 생태 정보 조회
func (c *FishBaseClient) GetEcology(ctx context.Context, specCode int) (*FishBaseEcology, error) {
	url := fmt.Sprintf("%s/ecology?speccode=%d", fishbaseAPIBase, specCode)
	<-c.rateLimiter

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []FishBaseEcology `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Data) == 0 {
		return nil, nil
	}
	return &result.Data[0], nil
}

// ToRawCrawlData FishBase 응답을 RawCrawlData로 변환
func (c *FishBaseClient) ToRawCrawlData(species FishBaseSpecies) *domain.RawCrawlData {
	raw, _ := json.Marshal(species)
	sourceID := fmt.Sprintf("%d", species.SpecCode)
	sourceURL := fmt.Sprintf("%s/species/%d", fishbaseAPIBase, species.SpecCode)

	return &domain.RawCrawlData{
		SourceName:  "fishbase",
		SourceID:    sourceID,
		SourceURL:   &sourceURL,
		RawJSON:     raw,
		ContentHash: hashContent(raw),
		ParseStatus: "PENDING",
	}
}

func (c *FishBaseClient) fetchSpecies(ctx context.Context, url string) ([]FishBaseSpecies, error) {
	<-c.rateLimiter

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fishbase API error: %d for %s", resp.StatusCode, url)
	}

	var result struct {
		Data []FishBaseSpecies `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	c.logger.Info("fishbase species fetched", "count", len(result.Data), "url", url)
	return result.Data, nil
}
