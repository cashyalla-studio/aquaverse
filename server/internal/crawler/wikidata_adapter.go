package crawler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/cashyalla/aquaverse/internal/domain"
)

const (
	wikidataSPARQLEndpoint = "https://query.wikidata.org/sparql"
	wikidataSPARQLTimeout  = 60 * time.Second

	// 다국어 공통명을 가진 주요 어종을 수집하는 SPARQL 쿼리.
	// P31=Q16521(taxon), P225(scientific name), P1843(taxon common name), P171(parent taxon)
	wikidataSPARQLQuery = `
SELECT ?taxon ?sciName ?commonName ?lang WHERE {
  ?taxon wdt:P31 wd:Q16521.
  ?taxon wdt:P225 ?sciName.
  ?taxon wdt:P1843 ?commonName.
  BIND(LANG(?commonName) AS ?lang)
  FILTER(?lang IN ("ko","en","ja","zh","de","fr","es","pt","ar","he"))
}
LIMIT 50000`
)

// WikidataAdapter Wikidata SPARQL SourceAdapter 구현체.
// 다국어 일반명을 가진 분류군의 scientific_name 기준 그룹 데이터를 수집한다.
// 라이선스: CC0 (Wikidata).
type WikidataAdapter struct {
	client    *http.Client
	logger    *slog.Logger
	userAgent string
}

// wikidataSPARQLResponse SPARQL JSON 응답 구조
type wikidataSPARQLResponse struct {
	Results struct {
		Bindings []wikidataBinding `json:"bindings"`
	} `json:"results"`
}

// wikidataBinding SPARQL 결과 한 행
type wikidataBinding struct {
	Taxon struct {
		Value string `json:"value"` // e.g. "http://www.wikidata.org/entity/Q183951"
	} `json:"taxon"`
	SciName struct {
		Value string `json:"value"`
	} `json:"sciName"`
	CommonName struct {
		Value string `json:"value"`
		Lang  string `json:"xml:lang"`
	} `json:"commonName"`
	Lang struct {
		Value string `json:"value"`
	} `json:"lang"`
}

// wikidataGroup scientific_name 기준으로 그룹핑한 중간 데이터
type wikidataGroup struct {
	QID         string            // Wikidata QID (예: "Q183951")
	ScientificName string
	Names       map[string]string // lang -> name
}

// NewWikidataAdapter WikidataAdapter 생성자
func NewWikidataAdapter(logger *slog.Logger, userAgent string) *WikidataAdapter {
	return &WikidataAdapter{
		client: &http.Client{
			Timeout: wikidataSPARQLTimeout,
		},
		logger:    logger,
		userAgent: userAgent,
	}
}

// Name 소스 식별자 반환
func (a *WikidataAdapter) Name() string {
	return "wikidata"
}

// Crawl SPARQL 쿼리로 다국어 공통명 데이터를 수집하고 scientific_name 기준으로 그룹핑해 저장한다.
// SPARQL 응답 실패 시 (0, nil)을 반환해 graceful degradation을 지원한다.
func (a *WikidataAdapter) Crawl(ctx context.Context, jobID int64, save func(*domain.RawCrawlData) error) (int, error) {
	bindings, err := a.executeSPARQL(ctx)
	if err != nil {
		a.logger.Warn("wikidata_adapter: SPARQL query failed, skipping",
			"err", err)
		// graceful degradation
		return 0, nil
	}

	// scientific_name 기준으로 그룹핑
	groups := a.groupByScientificName(bindings)
	a.logger.Info("wikidata_adapter: grouped species",
		"species_count", len(groups),
		"total_bindings", len(bindings))

	fetched := 0
	for _, group := range groups {
		rawData, err := a.toRawCrawlData(jobID, group)
		if err != nil {
			a.logger.Warn("wikidata_adapter: marshal failed",
				"scientific_name", group.ScientificName, "err", err)
			continue
		}
		if err := save(rawData); err != nil {
			a.logger.Error("wikidata_adapter: save failed",
				"scientific_name", group.ScientificName, "err", err)
			continue
		}
		fetched++
	}

	return fetched, nil
}

// executeSPARQL SPARQL 엔드포인트에 쿼리를 보내고 바인딩 목록을 반환한다.
func (a *WikidataAdapter) executeSPARQL(ctx context.Context) ([]wikidataBinding, error) {
	params := url.Values{}
	params.Set("format", "json")
	params.Set("query", wikidataSPARQLQuery)
	requestURL := wikidataSPARQLEndpoint + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("wikidata_adapter: new request: %w", err)
	}
	// Wikidata는 User-Agent 헤더 필수
	req.Header.Set("User-Agent", a.userAgent)
	req.Header.Set("Accept", "application/sparql-results+json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("wikidata_adapter: http do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wikidata_adapter: unexpected status %d", resp.StatusCode)
	}

	var sparqlResp wikidataSPARQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&sparqlResp); err != nil {
		return nil, fmt.Errorf("wikidata_adapter: decode SPARQL response: %w", err)
	}

	return sparqlResp.Results.Bindings, nil
}

// extractQID Wikidata 엔티티 URI에서 QID를 추출한다.
// 예: "http://www.wikidata.org/entity/Q183951" -> "Q183951"
func extractQID(entityURI string) string {
	for i := len(entityURI) - 1; i >= 0; i-- {
		if entityURI[i] == '/' {
			return entityURI[i+1:]
		}
	}
	return entityURI
}

// groupByScientificName SPARQL 바인딩을 scientific_name 기준으로 그룹핑한다.
func (a *WikidataAdapter) groupByScientificName(bindings []wikidataBinding) map[string]*wikidataGroup {
	groups := make(map[string]*wikidataGroup)

	for _, b := range bindings {
		sciName := b.SciName.Value
		if sciName == "" {
			continue
		}

		lang := b.CommonName.Lang
		if lang == "" {
			lang = b.Lang.Value
		}
		name := b.CommonName.Value

		if _, exists := groups[sciName]; !exists {
			qid := extractQID(b.Taxon.Value)
			groups[sciName] = &wikidataGroup{
				QID:            qid,
				ScientificName: sciName,
				Names:          make(map[string]string),
			}
		}

		if lang != "" && name != "" {
			// 같은 언어에 이름이 여러 개면 첫 번째 값 유지
			if _, alreadySet := groups[sciName].Names[lang]; !alreadySet {
				groups[sciName].Names[lang] = name
			}
		}
	}

	return groups
}

// toRawCrawlData wikidataGroup을 domain.FishData로 변환한 뒤 RawCrawlData로 래핑한다.
// RawJSON은 domain.FishData JSON이어야 pipeline.parseCrawlData()에서 언마샬 가능하다.
// creature_category는 SPARQL에서 P171 체인 추적이 복잡하므로 기본값 "fish"로 설정한다.
func (a *WikidataAdapter) toRawCrawlData(jobID int64, group *wikidataGroup) (*domain.RawCrawlData, error) {
	sourceURL := fmt.Sprintf("https://www.wikidata.org/wiki/%s", group.QID)
	license := "CC0"
	attribution := "Wikidata"

	// 영어 일반명 우선 사용
	commonName := group.Names["en"]

	// 다국어 이름 메모 생성 (CareNotes 필드에 JSON 직렬화해서 임시 보관)
	// 파이프라인 AI 보완 단계에서 fish_translations로 분리된다.
	var careNotes *string
	if len(group.Names) > 0 {
		namesJSON, err := json.Marshal(group.Names)
		if err == nil {
			note := fmt.Sprintf("wikidata_multilang:%s", string(namesJSON))
			careNotes = &note
		}
	}

	fish := domain.FishData{
		ScientificName:    group.ScientificName,
		PrimaryCommonName: commonName,
		PrimarySource:     "wikidata",
		SourceURL:         &sourceURL,
		License:           &license,
		Attribution:       &attribution,
		CreatureCategory:  string(domain.CategoryFish), // 기본값
		PublishStatus:     domain.PublishStatusDraft,
		CareNotes:         careNotes,
	}

	rawJSON, err := json.Marshal(fish)
	if err != nil {
		return nil, fmt.Errorf("wikidata_adapter: marshal fish data: %w", err)
	}

	jobIDCopy := jobID
	return &domain.RawCrawlData{
		CrawlJobID:  &jobIDCopy,
		SourceName:  "wikidata",
		SourceID:    group.QID,
		SourceURL:   &sourceURL,
		RawJSON:     rawJSON,
		ContentHash: hashContent(rawJSON),
		ParseStatus: "PENDING",
	}, nil
}
