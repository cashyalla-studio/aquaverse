package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type SpeciesIdentifyService struct {
	db     *sqlx.DB
	apiKey string
}

func NewSpeciesIdentifyService(db *sqlx.DB, apiKey string) *SpeciesIdentifyService {
	return &SpeciesIdentifyService{db: db, apiKey: apiKey}
}

// IdentifyFromBase64 Base64 이미지로 어종 식별
func (s *SpeciesIdentifyService) IdentifyFromBase64(
	ctx context.Context,
	userID *uuid.UUID,
	base64Image, mediaType string,
) (*domain.SpeciesIdentifyResult, error) {
	start := time.Now()

	prompt := `이 물고기 사진을 분석하여 어종을 식별해주세요.

반드시 다음 JSON 형식으로만 응답하세요 (다른 텍스트 없이):
{
  "candidates": [
    {
      "name": "한국어 일반명",
      "scientific_name": "학명",
      "confidence": 0.95,
      "description": "이 어종을 식별한 근거 (특징 2-3가지)"
    },
    {
      "name": "두 번째 후보",
      "scientific_name": "학명",
      "confidence": 0.70,
      "description": "특징 설명"
    },
    {
      "name": "세 번째 후보",
      "scientific_name": "학명",
      "confidence": 0.30,
      "description": "특징 설명"
    }
  ]
}

가능한 경우 최대 3개의 후보를 제시하고, 확신도(confidence)는 0.0~1.0 사이 값으로 표현하세요.
물고기가 아닌 경우: {"candidates": []}`

	var responseText string
	var err error

	if s.apiKey != "" {
		responseText, err = callClaudeWithImage(ctx, s.apiKey, base64Image, mediaType, prompt)
		if err != nil {
			return nil, fmt.Errorf("claude api error: %w", err)
		}
	} else {
		// Mock 응답 (API 키 없을 때)
		responseText = `{"candidates": [{"name": "네온테트라", "scientific_name": "Paracheirodon innesi", "confidence": 0.90, "description": "파란색과 붉은색의 선명한 색상, 작은 체형 (최대 4cm), 카리브해 남미 산"}, {"name": "카디날 테트라", "scientific_name": "Paracheirodon axelrodi", "confidence": 0.60, "description": "네온테트라와 유사하나 붉은 줄이 더 길게 이어짐"}, {"name": "글로우라이트 테트라", "scientific_name": "Hemigrammus erythrozonus", "confidence": 0.20, "description": "주황빛 선 하나가 특징적인 테트라류"}]}`
	}

	processingMs := int(time.Since(start).Milliseconds())

	// JSON 파싱
	// 응답에서 JSON 추출 (Claude가 마크다운 블록으로 감싸는 경우 처리)
	jsonStr := responseText
	if idx := strings.Index(responseText, "{"); idx > 0 {
		jsonStr = responseText[idx:]
	}
	if idx := strings.LastIndex(jsonStr, "}"); idx >= 0 {
		jsonStr = jsonStr[:idx+1]
	}

	var parsed struct {
		Candidates []domain.SpeciesCandidate `json:"candidates"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	// DB에서 후보 어종 ID 매핑 (정확도 높은 순으로)
	for i, c := range parsed.Candidates {
		var fishID int64
		err := s.db.QueryRowContext(ctx, `
            SELECT id FROM fish_data
            WHERE scientific_name ILIKE $1 OR primary_common_name ILIKE $2
            LIMIT 1
        `, c.ScientificName, c.Name).Scan(&fishID)
		if err == nil && fishID > 0 {
			parsed.Candidates[i].FishDataID = &fishID
		}
	}

	// 로그 저장
	candidatesJSON, _ := json.Marshal(parsed.Candidates)
	var topFishID *int64
	if len(parsed.Candidates) > 0 {
		topFishID = parsed.Candidates[0].FishDataID
	}

	var logID int64
	var userIDVal interface{}
	if userID != nil {
		userIDVal = *userID
	}
	s.db.QueryRowContext(ctx, `
        INSERT INTO species_identification_log
            (user_id, candidates, top_fish_id, processing_ms)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `, userIDVal, candidatesJSON, topFishID, processingMs).Scan(&logID)

	return &domain.SpeciesIdentifyResult{
		ID:           logID,
		Candidates:   parsed.Candidates,
		ProcessingMs: processingMs,
		CreatedAt:    time.Now(),
	}, nil
}
