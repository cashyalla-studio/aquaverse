package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type TankDoctorService struct {
	db         *sqlx.DB
	rdb        *redis.Client
	httpClient *http.Client
	apiKey     string
}

func NewTankDoctorService(db *sqlx.DB, rdb *redis.Client) *TankDoctorService {
	return &TankDoctorService{
		db:         db,
		rdb:        rdb,
		httpClient: &http.Client{Timeout: 60 * time.Second},
		apiKey:     os.Getenv("ANTHROPIC_API_KEY"),
	}
}

// RecordWaterParams 수질 파라미터 기록
func (s *TankDoctorService) RecordWaterParams(ctx context.Context, params domain.WaterParams) (*domain.WaterParams, error) {
	var id int64
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO tank_water_params (tank_id, temp_c, ph, ammonia_ppm, nitrite_ppm, nitrate_ppm, gh_dgh, kh_dkh, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id`,
		params.TankID, params.TempC, params.PH, params.AmmoniaPPM,
		params.NitritePPM, params.NitratePPM, params.GhDgh, params.KhDkh, params.Notes,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	params.ID = id
	return &params, nil
}

// GetWaterHistory 수질 이력 조회
func (s *TankDoctorService) GetWaterHistory(ctx context.Context, tankID int64, limit int) ([]domain.WaterParams, error) {
	var params []domain.WaterParams
	err := s.db.SelectContext(ctx, &params, `
		SELECT id, tank_id,
		       temp_c, ph, ammonia_ppm, nitrite_ppm, nitrate_ppm, gh_dgh, kh_dkh,
		       recorded_at, COALESCE(notes,'') as notes
		FROM tank_water_params
		WHERE tank_id=$1
		ORDER BY recorded_at DESC
		LIMIT $2
	`, tankID, limit)
	return params, err
}

// inhabitant 수조 어종 정보
type inhabitant struct {
	FishName string  `db:"fish_name"`
	Quantity int     `db:"quantity"`
	TempMin  float64 `db:"temp_min"`
	TempMax  float64 `db:"temp_max"`
	PhMin    float64 `db:"ph_min"`
	PhMax    float64 `db:"ph_max"`
}

// DiagnoseTank 수조 AI 진단 (Redis 5분 캐시)
func (s *TankDoctorService) DiagnoseTank(ctx context.Context, tankID int64) (*domain.TankDiagnosis, error) {
	cacheKey := fmt.Sprintf("tank:diagnosis:%d", tankID)

	// Redis 캐시 확인
	cached, err := s.rdb.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var diag domain.TankDiagnosis
		if json.Unmarshal(cached, &diag) == nil {
			return &diag, nil
		}
	}

	// 최신 수질 파라미터 조회
	var params domain.WaterParams
	if err := s.db.GetContext(ctx, &params, `
		SELECT id, tank_id, temp_c, ph, ammonia_ppm, nitrite_ppm, nitrate_ppm, gh_dgh, kh_dkh,
		       recorded_at, COALESCE(notes,'') as notes
		FROM tank_water_params WHERE tank_id=$1 ORDER BY recorded_at DESC LIMIT 1
	`, tankID); err != nil {
		return nil, fmt.Errorf("수질 데이터 없음: 먼저 수질을 기록해주세요")
	}

	// 수조 어종 목록 조회
	var inhabitants []inhabitant
	s.db.SelectContext(ctx, &inhabitants, `
		SELECT fd.common_name as fish_name, ti.quantity,
		       COALESCE(fd.temp_min,20) as temp_min, COALESCE(fd.temp_max,28) as temp_max,
		       COALESCE(fd.ph_min,6.5) as ph_min, COALESCE(fd.ph_max,7.5) as ph_max
		FROM tank_inhabitants ti
		JOIN fish_data fd ON fd.id = ti.fish_data_id
		WHERE ti.tank_id=$1
	`, tankID)

	// 프롬프트 구성
	prompt := s.buildPrompt(params, inhabitants)

	// Claude API 호출
	diag, err := s.callClaude(ctx, tankID, prompt)
	if err != nil {
		return nil, err
	}

	// Redis 캐시 저장 (5분)
	if data, err := json.Marshal(diag); err == nil {
		s.rdb.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	return diag, nil
}

func (s *TankDoctorService) buildPrompt(params domain.WaterParams, inhabitants []inhabitant) string {
	buf := &bytes.Buffer{}
	buf.WriteString("다음 수조의 수질 데이터와 어종 정보를 분석해주세요.\n\n")
	buf.WriteString("## 수질 현황\n")
	if params.TempC != nil {
		fmt.Fprintf(buf, "- 수온: %.1f°C\n", *params.TempC)
	}
	if params.PH != nil {
		fmt.Fprintf(buf, "- pH: %.1f\n", *params.PH)
	}
	if params.AmmoniaPPM != nil {
		fmt.Fprintf(buf, "- 암모니아: %.2f ppm\n", *params.AmmoniaPPM)
	}
	if params.NitritePPM != nil {
		fmt.Fprintf(buf, "- 아질산: %.2f ppm\n", *params.NitritePPM)
	}
	if params.NitratePPM != nil {
		fmt.Fprintf(buf, "- 질산: %.2f ppm\n", *params.NitratePPM)
	}
	if params.GhDgh != nil {
		fmt.Fprintf(buf, "- GH: %.1f dGH\n", *params.GhDgh)
	}
	if params.KhDkh != nil {
		fmt.Fprintf(buf, "- KH: %.1f dKH\n", *params.KhDkh)
	}

	if len(inhabitants) > 0 {
		buf.WriteString("\n## 수조 어종\n")
		for _, inh := range inhabitants {
			fmt.Fprintf(buf, "- %s x%d (적정 수온: %.0f~%.0f°C, pH: %.1f~%.1f)\n",
				inh.FishName, inh.Quantity, inh.TempMin, inh.TempMax, inh.PhMin, inh.PhMax)
		}
	}

	buf.WriteString("\n## 반드시 JSON 형식으로만 응답하세요:\n")
	buf.WriteString(`{
  "summary": "한 줄 전체 진단 요약",
  "fish_states": [
    {"fish_name": "어종명", "status": "good|warning|danger", "issue": "문제 설명 (없으면 빈 문자열)", "suggestion": "조치 방법 (없으면 빈 문자열)"}
  ],
  "actions": ["즉시 해야 할 조치 1", "조치 2"]
}`)
	return buf.String()
}

func (s *TankDoctorService) callClaude(ctx context.Context, tankID int64, prompt string) (*domain.TankDiagnosis, error) {
	if s.apiKey == "" {
		// API 키 없으면 mock 응답 반환
		return &domain.TankDiagnosis{
			TankID:     tankID,
			Summary:    "Claude API 키가 설정되지 않았습니다. ANTHROPIC_API_KEY 환경변수를 설정하세요.",
			FishStates: []domain.FishStatus{},
			Actions:    []string{"ANTHROPIC_API_KEY 환경변수 설정"},
			CreatedAt:  time.Now().Format(time.RFC3339),
		}, nil
	}

	reqBody := map[string]interface{}{
		"model":      "claude-haiku-4-5-20251001",
		"max_tokens": 1024,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", s.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Claude API 오류: %d %s", resp.StatusCode, string(body))
	}

	var apiResp struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil || len(apiResp.Content) == 0 {
		return nil, fmt.Errorf("Claude API 응답 파싱 오류")
	}

	text := apiResp.Content[0].Text
	// JSON 추출 (```json ... ``` 코드블록 제거)
	if start := bytes.Index([]byte(text), []byte("{")); start >= 0 {
		if end := bytes.LastIndex([]byte(text), []byte("}")); end >= start {
			text = text[start : end+1]
		}
	}

	var result struct {
		Summary    string `json:"summary"`
		FishStates []struct {
			FishName   string `json:"fish_name"`
			Status     string `json:"status"`
			Issue      string `json:"issue"`
			Suggestion string `json:"suggestion"`
		} `json:"fish_states"`
		Actions []string `json:"actions"`
	}

	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("진단 결과 파싱 오류: %w", err)
	}

	diag := &domain.TankDiagnosis{
		TankID:    tankID,
		Summary:   result.Summary,
		Actions:   result.Actions,
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	for _, fs := range result.FishStates {
		diag.FishStates = append(diag.FishStates, domain.FishStatus{
			FishName:   fs.FishName,
			Status:     fs.Status,
			Issue:      fs.Issue,
			Suggestion: fs.Suggestion,
		})
	}

	return diag, nil
}
