package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/cashyalla/aquaverse/internal/domain"
)

const anthropicAPIBase = "https://api.anthropic.com/v1"

// AIEnricher Claude API를 이용한 열대어 데이터 자동 가공
type AIEnricher struct {
	client    *http.Client
	apiKey    string
	model     string
	maxTokens int
	logger    *slog.Logger
}

type claudeRequest struct {
	Model     string           `json:"model"`
	MaxTokens int              `json:"max_tokens"`
	Messages  []claudeMessage  `json:"messages"`
	System    string           `json:"system"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

// EnrichedFishData AI가 채워주는 필드들
type EnrichedFishData struct {
	CareLevel         *string  `json:"care_level,omitempty"`
	Temperament       *string  `json:"temperament,omitempty"`
	MinTankSizeLiters *int     `json:"min_tank_size_liters,omitempty"`
	DietType          *string  `json:"diet_type,omitempty"`
	DietNotes         *string  `json:"diet_notes,omitempty"`
	BreedingNotes     *string  `json:"breeding_notes,omitempty"`
	CareNotes         *string  `json:"care_notes,omitempty"`
	PHMin             *float64 `json:"ph_min,omitempty"`
	PHMax             *float64 `json:"ph_max,omitempty"`
	TempMinC          *float64 `json:"temp_min_c,omitempty"`
	TempMaxC          *float64 `json:"temp_max_c,omitempty"`
	LifespanYears     *float64 `json:"lifespan_years,omitempty"`
	MaxSizeCm         *float64 `json:"max_size_cm,omitempty"`
}

const systemPrompt = `You are an expert tropical fish biologist and aquarium specialist with 20+ years of experience.
Your task is to analyze raw fish data and fill in missing fields based on your expert knowledge.
Always respond with valid JSON only, no explanations.
Use metric units (cm, liters, Celsius).
Care level: BEGINNER, INTERMEDIATE, or EXPERT only.
Temperament: PEACEFUL, SEMI_AGGRESSIVE, or AGGRESSIVE only.
Diet type: OMNIVORE, CARNIVORE, or HERBIVORE only.`

func NewAIEnricher(apiKey, model string, maxTokens int, logger *slog.Logger) *AIEnricher {
	return &AIEnricher{
		client:    &http.Client{Timeout: 60 * time.Second},
		apiKey:    apiKey,
		model:     model,
		maxTokens: maxTokens,
		logger:    logger,
	}
}

// Enrich 낮은 품질 데이터를 AI로 보완
func (e *AIEnricher) Enrich(ctx context.Context, fish *domain.FishData, missing []string) (*EnrichedFishData, error) {
	rawData, _ := json.Marshal(map[string]interface{}{
		"scientific_name":     fish.ScientificName,
		"genus":               fish.Genus,
		"species":             fish.Species,
		"family":              fish.Family,
		"primary_common_name": fish.PrimaryCommonName,
		"ph_min":              fish.PHMin,
		"ph_max":              fish.PHMax,
		"temp_min_c":          fish.TempMinC,
		"temp_max_c":          fish.TempMaxC,
		"max_size_cm":         fish.MaxSizeCm,
		"diet_notes":          fish.DietNotes,
		"care_notes":          fish.CareNotes,
	})

	prompt := fmt.Sprintf(`Analyze this tropical fish data and fill in ONLY the missing fields.

Raw data:
%s

Missing fields to fill: %v

Respond with JSON containing only the missing fields you can confidently fill.
Example response format:
{
  "care_level": "BEGINNER",
  "temperament": "PEACEFUL",
  "min_tank_size_liters": 60,
  "diet_type": "OMNIVORE",
  "care_notes": "Hardy fish suitable for beginners...",
  "breeding_notes": "Egg scatterer that breeds readily..."
}`, string(rawData), missing)

	reqBody := claudeRequest{
		Model:     e.model,
		MaxTokens: e.maxTokens,
		System:    systemPrompt,
		Messages: []claudeMessage{
			{Role: "user", Content: prompt},
		},
	}

	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicAPIBase+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", e.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("claude API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("claude API error: %d", resp.StatusCode)
	}

	var claudeResp claudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
		return nil, err
	}
	if len(claudeResp.Content) == 0 {
		return nil, fmt.Errorf("empty claude response")
	}

	var enriched EnrichedFishData
	if err := json.Unmarshal([]byte(claudeResp.Content[0].Text), &enriched); err != nil {
		e.logger.Warn("failed to parse AI enrichment response",
			"fish", fish.ScientificName, "response", claudeResp.Content[0].Text)
		return nil, err
	}

	e.logger.Info("AI enrichment completed",
		"fish", fish.ScientificName,
		"missing_filled", len(missing))

	return &enriched, nil
}

// ApplyEnrichment AI 결과를 FishData에 반영
func ApplyEnrichment(fish *domain.FishData, enriched *EnrichedFishData) {
	if enriched.CareLevel != nil && fish.CareLevel == nil {
		cl := domain.CareLevel(*enriched.CareLevel)
		fish.CareLevel = &cl
	}
	if enriched.Temperament != nil && fish.Temperament == nil {
		t := domain.Temperament(*enriched.Temperament)
		fish.Temperament = &t
	}
	if enriched.MinTankSizeLiters != nil && fish.MinTankSizeLiters == nil {
		fish.MinTankSizeLiters = enriched.MinTankSizeLiters
	}
	if enriched.DietType != nil && fish.DietType == nil {
		dt := domain.DietType(*enriched.DietType)
		fish.DietType = &dt
	}
	if enriched.DietNotes != nil && (fish.DietNotes == nil || *fish.DietNotes == "") {
		fish.DietNotes = enriched.DietNotes
	}
	if enriched.BreedingNotes != nil && (fish.BreedingNotes == nil || *fish.BreedingNotes == "") {
		fish.BreedingNotes = enriched.BreedingNotes
	}
	if enriched.CareNotes != nil && (fish.CareNotes == nil || *fish.CareNotes == "") {
		fish.CareNotes = enriched.CareNotes
	}
	if enriched.PHMin != nil && fish.PHMin == nil {
		fish.PHMin = enriched.PHMin
	}
	if enriched.PHMax != nil && fish.PHMax == nil {
		fish.PHMax = enriched.PHMax
	}
	if enriched.TempMinC != nil && fish.TempMinC == nil {
		fish.TempMinC = enriched.TempMinC
	}
	if enriched.TempMaxC != nil && fish.TempMaxC == nil {
		fish.TempMaxC = enriched.TempMaxC
	}
	if enriched.LifespanYears != nil && fish.LifespanYears == nil {
		fish.LifespanYears = enriched.LifespanYears
	}
	if enriched.MaxSizeCm != nil && fish.MaxSizeCm == nil {
		fish.MaxSizeCm = enriched.MaxSizeCm
	}
}
