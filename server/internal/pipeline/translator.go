package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/cashyalla/aquaverse/internal/domain"
)

// translationPriority 번역 우선순위 (비용 절감을 위해 수요 높은 언어 우선)
var translationPriority = []domain.Locale{
	domain.LocaleKO,
	domain.LocaleENUS,
	domain.LocaleJA,
	domain.LocaleZHCN,
	domain.LocaleZHTW,
	domain.LocaleDE,
	domain.LocaleFRFR,
	domain.LocaleFRCA,
	domain.LocaleES,
	domain.LocalePT,
	domain.LocaleENGB,
	domain.LocaleENAU,
	domain.LocaleAR,
	domain.LocaleHE,
}

// localeDisplayNames Claude 프롬프트에서 사용할 로케일 표시 이름
var localeDisplayNames = map[domain.Locale]string{
	domain.LocaleKO:   "Korean (ko)",
	domain.LocaleENUS: "English US (en-US)",
	domain.LocaleENGB: "English UK (en-GB)",
	domain.LocaleENAU: "English Australian (en-AU)",
	domain.LocaleJA:   "Japanese (ja)",
	domain.LocaleZHCN: "Chinese Simplified (zh-CN)",
	domain.LocaleZHTW: "Chinese Traditional (zh-TW)",
	domain.LocaleDE:   "German (de)",
	domain.LocaleFRFR: "French (fr-FR)",
	domain.LocaleFRCA: "French Canadian (fr-CA)",
	domain.LocaleES:   "Spanish (es)",
	domain.LocalePT:   "Portuguese (pt)",
	domain.LocaleAR:   "Arabic (ar)",
	domain.LocaleHE:   "Hebrew (he)",
}

// TranslationResult 단일 로케일 번역 결과
type TranslationResult struct {
	Locale        string `json:"locale"`
	CommonName    string `json:"common_name"`
	CareNotes     string `json:"care_notes"`
	DietNotes     string `json:"diet_notes"`
	BreedingNotes string `json:"breeding_notes"`
}

// Translator Claude API를 이용한 다국어 배치 번역
type Translator struct {
	client *http.Client
	apiKey string
	model  string
	logger *slog.Logger
}

// NewTranslator Translator 생성자
func NewTranslator(apiKey, model string, logger *slog.Logger) *Translator {
	return &Translator{
		client: &http.Client{Timeout: 90 * time.Second},
		apiKey: apiKey,
		model:  model,
		logger: logger,
	}
}

// TranslateFish 한 번의 Claude API 호출로 최대 5개 언어를 배치 번역한다.
// ANTHROPIC_API_KEY 미설정 시 nil, nil 반환 (graceful).
func (t *Translator) TranslateFish(ctx context.Context, fish *domain.FishData, locales []domain.Locale) ([]TranslationResult, error) {
	if t.apiKey == "" {
		t.logger.Info("ANTHROPIC_API_KEY not set, skipping translation",
			"fish", fish.ScientificName)
		return nil, nil
	}
	if len(locales) == 0 {
		return nil, nil
	}

	// 최대 5개 로케일씩 배치 처리
	if len(locales) > 5 {
		locales = locales[:5]
	}

	localeList := make([]string, 0, len(locales))
	for _, l := range locales {
		if name, ok := localeDisplayNames[l]; ok {
			localeList = append(localeList, name)
		} else {
			localeList = append(localeList, string(l))
		}
	}

	// 소스 데이터 구성
	sourceData := map[string]interface{}{
		"scientific_name":     fish.ScientificName,
		"primary_common_name": fish.PrimaryCommonName,
	}
	if fish.CareNotes != nil {
		sourceData["care_notes"] = *fish.CareNotes
	}
	if fish.DietNotes != nil {
		sourceData["diet_notes"] = *fish.DietNotes
	}
	if fish.BreedingNotes != nil {
		sourceData["breeding_notes"] = *fish.BreedingNotes
	}

	sourceJSON, _ := json.MarshalIndent(sourceData, "", "  ")

	// 출력 예시 구성
	exampleLocale := string(locales[0])
	prompt := fmt.Sprintf(`Translate the following tropical fish care information into these languages: %s.

Source data (English):
%s

Return a JSON array where each element represents one locale translation.
Each element must have exactly these fields:
- "locale": the locale code (e.g. "ko", "en-US")
- "common_name": translated common name for the fish
- "care_notes": translated care notes (keep empty string if source is empty)
- "diet_notes": translated diet notes (keep empty string if source is empty)
- "breeding_notes": translated breeding notes (keep empty string if source is empty)

Example output format:
[
  {
    "locale": "%s",
    "common_name": "...",
    "care_notes": "...",
    "diet_notes": "...",
    "breeding_notes": "..."
  }
]

Respond with valid JSON array only. No explanations, no markdown fences.`,
		strings.Join(localeList, ", "),
		string(sourceJSON),
		exampleLocale,
	)

	systemMsg := `You are an expert marine biologist and aquarium specialist fluent in all requested languages.
Translate fish care information accurately, preserving technical aquarium terminology.
For RTL languages (Arabic, Hebrew), ensure proper text direction.
Always respond with valid JSON only.`

	reqBody := claudeRequest{
		Model:     t.model,
		MaxTokens: 4096,
		System:    systemMsg,
		Messages: []claudeMessage{
			{Role: "user", Content: prompt},
		},
	}

	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicAPIBase+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create translation request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", t.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("claude translation API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("claude translation API error: %d", resp.StatusCode)
	}

	var claudeResp claudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
		return nil, fmt.Errorf("decode claude response: %w", err)
	}
	if len(claudeResp.Content) == 0 {
		return nil, fmt.Errorf("empty claude translation response")
	}

	rawText := claudeResp.Content[0].Text

	// JSON 파싱
	var results []TranslationResult
	if err := json.Unmarshal([]byte(rawText), &results); err != nil {
		t.logger.Warn("failed to parse translation response",
			"fish", fish.ScientificName,
			"response", rawText,
			"err", err)
		return nil, fmt.Errorf("parse translation JSON: %w", err)
	}

	t.logger.Info("batch translation completed",
		"fish", fish.ScientificName,
		"locales", len(results))

	return results, nil
}

// TranslateAllMissing 번역이 없는 로케일을 찾아서 5개 단위 배치로 번역한다.
// 완성된 FishTranslation 슬라이스를 반환한다.
func (t *Translator) TranslateAllMissing(
	ctx context.Context,
	fishID int64,
	fish *domain.FishData,
	existingLocales []domain.Locale,
) ([]domain.FishTranslation, error) {
	if t.apiKey == "" {
		t.logger.Info("ANTHROPIC_API_KEY not set, skipping TranslateAllMissing",
			"fish", fish.ScientificName)
		return nil, nil
	}

	// 기존 번역 셋 구성
	existing := make(map[domain.Locale]bool, len(existingLocales))
	for _, l := range existingLocales {
		existing[l] = true
	}

	// 번역이 필요한 로케일 수집 (우선순위 순서 유지)
	missing := make([]domain.Locale, 0, len(translationPriority))
	for _, l := range translationPriority {
		if !existing[l] {
			missing = append(missing, l)
		}
	}

	if len(missing) == 0 {
		t.logger.Info("all locales already translated",
			"fish", fish.ScientificName)
		return nil, nil
	}

	t.logger.Info("starting translation for missing locales",
		"fish", fish.ScientificName,
		"missing_count", len(missing))

	var allTranslations []domain.FishTranslation

	// 5개 단위 배치 처리
	batchSize := 5
	for i := 0; i < len(missing); i += batchSize {
		end := i + batchSize
		if end > len(missing) {
			end = len(missing)
		}
		batch := missing[i:end]

		results, err := t.TranslateFish(ctx, fish, batch)
		if err != nil {
			t.logger.Warn("batch translation failed",
				"fish", fish.ScientificName,
				"batch_start", i,
				"err", err)
			continue
		}
		if results == nil {
			break
		}

		now := time.Now()
		for _, r := range results {
			tr := domain.FishTranslation{
				FishDataID:        fishID,
				Locale:            domain.Locale(r.Locale),
				TranslationSource: "ai",
				TranslatedAt:      now,
				Verified:          false,
			}
			if r.CommonName != "" {
				s := r.CommonName
				tr.CommonName = &s
			}
			if r.CareNotes != "" {
				s := r.CareNotes
				tr.CareNotes = &s
			}
			if r.DietNotes != "" {
				s := r.DietNotes
				tr.DietNotes = &s
			}
			if r.BreedingNotes != "" {
				s := r.BreedingNotes
				tr.BreedingNotes = &s
			}
			allTranslations = append(allTranslations, tr)
		}
	}

	t.logger.Info("TranslateAllMissing completed",
		"fish", fish.ScientificName,
		"translations_created", len(allTranslations))

	return allTranslations, nil
}
