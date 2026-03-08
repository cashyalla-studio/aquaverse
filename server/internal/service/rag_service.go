package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// RAGService pgvector 기반 RAG 챗봇 서비스
type RAGService struct {
	db         *sqlx.DB
	rdb        *redis.Client
	httpClient *http.Client
	apiKey     string
}

func NewRAGService(db *sqlx.DB, rdb *redis.Client) *RAGService {
	return &RAGService{
		db:         db,
		rdb:        rdb,
		httpClient: &http.Client{Timeout: 60 * time.Second},
		apiKey:     os.Getenv("ANTHROPIC_API_KEY"),
	}
}

// ─── 공개 API ────────────────────────────────────────────────────────────────

// Answer 질문을 받아 pgvector RAG로 답변한다.
// sessionID가 빈 문자열이면 새 세션을 생성한다.
func (s *RAGService) Answer(ctx context.Context, question, sessionID, userID string) (*domain.AskResponse, error) {
	// 1. 세션 조회 또는 생성
	sid, err := s.ensureSession(ctx, sessionID, userID)
	if err != nil {
		return nil, fmt.Errorf("세션 오류: %w", err)
	}

	// 2. pgvector 유사도 검색
	sources, contextText, err := s.retrieveContext(ctx, question)
	if err != nil {
		// pgvector 미설치 등의 fallback: 컨텍스트 없이 진행
		contextText = ""
		sources = nil
	}

	// 3. Claude API 호출
	answer, err := s.callClaudeRAG(ctx, question, contextText)
	if err != nil {
		return nil, fmt.Errorf("AI 응답 오류: %w", err)
	}

	// 4. 메시지 저장
	if saveErr := s.saveMessages(ctx, sid, question, answer); saveErr != nil {
		// 저장 실패는 무시하고 응답은 반환
		_ = saveErr
	}

	// 5. 세션 last_msg_at 갱신
	_, _ = s.db.ExecContext(ctx,
		`UPDATE rag_sessions SET last_msg_at = NOW() WHERE id = $1`, sid)

	return &domain.AskResponse{
		Answer:    answer,
		SessionID: sid,
		Sources:   sources,
	}, nil
}

// IndexFish fish_data 테이블에서 어종 정보를 가져와 임베딩을 생성한 뒤 fish_embeddings에 저장한다.
func (s *RAGService) IndexFish(ctx context.Context, fishID int64) error {
	// 어종 정보 조회
	var row struct {
		ID                int64   `db:"id"`
		PrimaryCommonName string  `db:"primary_common_name"`
		ScientificName    string  `db:"scientific_name"`
		Family            string  `db:"family"`
		CareNotes         *string `db:"care_notes"`
		DietNotes         *string `db:"diet_notes"`
		BreedingNotes     *string `db:"breeding_notes"`
		PHMin             *float64 `db:"ph_min"`
		PHMax             *float64 `db:"ph_max"`
		TempMinC          *float64 `db:"temp_min_c"`
		TempMaxC          *float64 `db:"temp_max_c"`
		MaxSizeCm         *float64 `db:"max_size_cm"`
	}
	if err := s.db.GetContext(ctx, &row, `
		SELECT id, primary_common_name, scientific_name, family,
		       care_notes, diet_notes, breeding_notes,
		       ph_min, ph_max, temp_min_c, temp_max_c, max_size_cm
		FROM fish_data WHERE id = $1
	`, fishID); err != nil {
		return fmt.Errorf("어종 조회 실패 (id=%d): %w", fishID, err)
	}

	// 임베딩할 텍스트 구성
	content := s.buildFishContent(row.PrimaryCommonName, row.ScientificName, row.Family,
		row.CareNotes, row.DietNotes, row.BreedingNotes,
		row.PHMin, row.PHMax, row.TempMinC, row.TempMaxC, row.MaxSizeCm)

	// 임베딩 생성 (API 키 없으면 mock zeros)
	embedding, err := s.getEmbedding(ctx, content)
	if err != nil {
		return fmt.Errorf("임베딩 생성 실패: %w", err)
	}

	// fish_embeddings UPSERT
	vectorLiteral := float32SliceToVectorLiteral(embedding)
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO fish_embeddings (fish_id, embedding, content, updated_at)
		VALUES ($1, $2::vector, $3, NOW())
		ON CONFLICT (fish_id) DO UPDATE
		  SET embedding  = EXCLUDED.embedding,
		      content    = EXCLUDED.content,
		      updated_at = NOW()
	`, fishID, vectorLiteral, content)
	return err
}

// GetSessionsByUser 사용자의 채팅 세션 목록을 반환한다.
func (s *RAGService) GetSessionsByUser(ctx context.Context, userID string) ([]domain.RAGSession, error) {
	var sessions []domain.RAGSession
	err := s.db.SelectContext(ctx, &sessions, `
		SELECT id, user_id, created_at, last_msg_at
		FROM rag_sessions
		WHERE user_id = $1
		ORDER BY last_msg_at DESC
		LIMIT 50
	`, userID)
	return sessions, err
}

// GetMessages 세션의 메시지 목록을 반환한다.
func (s *RAGService) GetMessages(ctx context.Context, sessionID string) ([]domain.RAGMessage, error) {
	var messages []domain.RAGMessage
	err := s.db.SelectContext(ctx, &messages, `
		SELECT id, session_id, role, content, created_at
		FROM rag_messages
		WHERE session_id = $1
		ORDER BY created_at ASC
		LIMIT 200
	`, sessionID)
	return messages, err
}

// ─── 내부 헬퍼 ───────────────────────────────────────────────────────────────

// ensureSession 기존 세션을 확인하거나 새 세션을 생성하고 UUID 문자열을 반환한다.
func (s *RAGService) ensureSession(ctx context.Context, sessionID, userID string) (string, error) {
	if sessionID != "" {
		// 세션이 실제 존재하는지 확인
		var exists bool
		err := s.db.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM rag_sessions WHERE id = $1)`, sessionID,
		).Scan(&exists)
		if err == nil && exists {
			return sessionID, nil
		}
	}

	// 새 세션 생성
	var newID string
	var err error
	if userID != "" {
		err = s.db.QueryRowContext(ctx,
			`INSERT INTO rag_sessions (user_id) VALUES ($1) RETURNING id`, userID,
		).Scan(&newID)
	} else {
		err = s.db.QueryRowContext(ctx,
			`INSERT INTO rag_sessions DEFAULT VALUES RETURNING id`,
		).Scan(&newID)
	}
	return newID, err
}

// retrieveContext pgvector 코사인 유사도 검색으로 관련 어종 정보를 가져온다.
// pgvector extension이 없거나 임베딩 테이블이 비어 있으면 빈 컨텍스트를 반환한다.
func (s *RAGService) retrieveContext(ctx context.Context, question string) ([]domain.FishSource, string, error) {
	// 질문 임베딩 생성
	qEmbedding, err := s.getEmbedding(ctx, question)
	if err != nil {
		return nil, "", err
	}

	vectorLiteral := float32SliceToVectorLiteral(qEmbedding)

	rows, err := s.db.QueryContext(ctx, `
		SELECT fe.fish_id, fe.content,
		       1 - (fe.embedding <=> $1::vector) AS similarity,
		       fd.primary_common_name
		FROM fish_embeddings fe
		JOIN fish_data fd ON fd.id = fe.fish_id
		ORDER BY fe.embedding <=> $1::vector
		LIMIT 5
	`, vectorLiteral)
	if err != nil {
		// pgvector 미설치 등 — fallback
		return nil, "", fmt.Errorf("pgvector 검색 실패: %w", err)
	}
	defer rows.Close()

	var sources []domain.FishSource
	var contextBuf bytes.Buffer

	for rows.Next() {
		var fishID int64
		var content string
		var similarity float64
		var name string
		if err := rows.Scan(&fishID, &content, &similarity, &name); err != nil {
			continue
		}
		// 유사도 0.5 이상인 결과만 컨텍스트로 사용
		if similarity >= 0.5 {
			fmt.Fprintf(&contextBuf, "### %s\n%s\n\n", name, content)
			sources = append(sources, domain.FishSource{ID: fishID, Name: name})
		}
	}

	return sources, contextBuf.String(), nil
}

// callClaudeRAG Claude API에 RAG 프롬프트를 전송하고 답변 텍스트를 반환한다.
func (s *RAGService) callClaudeRAG(ctx context.Context, question, contextText string) (string, error) {
	if s.apiKey == "" {
		return "죄송합니다. AI 기능이 현재 비활성화되어 있습니다.", nil
	}

	var prompt strings.Builder
	prompt.WriteString("당신은 반려생물 전문 AI 어시스턴트 Finara입니다.\n")
	if contextText != "" {
		prompt.WriteString("다음 어종 정보를 참고하여 질문에 답변하세요:\n\n")
		prompt.WriteString(contextText)
		prompt.WriteString("---\n\n")
	}
	prompt.WriteString("질문: ")
	prompt.WriteString(question)

	reqBody := map[string]interface{}{
		"model":      "claude-haiku-4-5-20251001",
		"max_tokens": 1024,
		"messages": []map[string]string{
			{"role": "user", "content": prompt.String()},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", s.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Claude API 오류: %d %s", resp.StatusCode, string(body))
	}

	var apiResp struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil || len(apiResp.Content) == 0 {
		return "", fmt.Errorf("Claude API 응답 파싱 오류")
	}
	return apiResp.Content[0].Text, nil
}

// getEmbedding 텍스트에 대한 1536차원 임베딩 벡터를 반환한다.
// ANTHROPIC_API_KEY가 없거나 Embeddings API가 실패하면 zero 벡터를 반환한다.
func (s *RAGService) getEmbedding(ctx context.Context, text string) ([]float32, error) {
	const dim = 1536

	if s.apiKey == "" {
		return make([]float32, dim), nil
	}

	reqBody := map[string]interface{}{
		"model": "voyage-3",
		"input": text,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/embeddings", bytes.NewReader(bodyBytes))
	if err != nil {
		return make([]float32, dim), nil
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", s.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return make([]float32, dim), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// 임베딩 API 실패 → zero 벡터 fallback
		return make([]float32, dim), nil
	}

	var apiResp struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &apiResp); err != nil || len(apiResp.Data) == 0 {
		return make([]float32, dim), nil
	}

	emb := apiResp.Data[0].Embedding
	// 차원 수 보정
	if len(emb) < dim {
		padded := make([]float32, dim)
		copy(padded, emb)
		return padded, nil
	}
	return emb[:dim], nil
}

// saveMessages 사용자 질문과 어시스턴트 답변을 DB에 저장한다.
func (s *RAGService) saveMessages(ctx context.Context, sessionID, question, answer string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO rag_messages (session_id, role, content)
		VALUES ($1, 'user',      $2),
		       ($1, 'assistant', $3)
	`, sessionID, question, answer)
	return err
}

// buildFishContent 어종 정보로부터 임베딩할 텍스트를 구성한다.
func (s *RAGService) buildFishContent(
	commonName, scientificName, family string,
	careNotes, dietNotes, breedingNotes *string,
	phMin, phMax, tempMin, tempMax, maxSize *float64,
) string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "어종명: %s (%s)\n", commonName, scientificName)
	fmt.Fprintf(&buf, "과(Family): %s\n", family)

	if phMin != nil && phMax != nil {
		fmt.Fprintf(&buf, "적정 pH: %.1f ~ %.1f\n", *phMin, *phMax)
	}
	if tempMin != nil && tempMax != nil {
		fmt.Fprintf(&buf, "적정 수온: %.0f ~ %.0f°C\n", *tempMin, *tempMax)
	}
	if maxSize != nil {
		fmt.Fprintf(&buf, "최대 크기: %.1f cm\n", *maxSize)
	}
	if careNotes != nil && *careNotes != "" {
		fmt.Fprintf(&buf, "사육 정보: %s\n", *careNotes)
	}
	if dietNotes != nil && *dietNotes != "" {
		fmt.Fprintf(&buf, "식이 정보: %s\n", *dietNotes)
	}
	if breedingNotes != nil && *breedingNotes != "" {
		fmt.Fprintf(&buf, "번식 정보: %s\n", *breedingNotes)
	}
	return buf.String()
}

// float32SliceToVectorLiteral []float32 슬라이스를 PostgreSQL vector 리터럴 문자열로 변환한다.
// 예: [0.1, 0.2, 0.3] → "[0.1,0.2,0.3]"
func float32SliceToVectorLiteral(v []float32) string {
	var buf bytes.Buffer
	buf.WriteByte('[')
	for i, f := range v {
		if i > 0 {
			buf.WriteByte(',')
		}
		fmt.Fprintf(&buf, "%g", f)
	}
	buf.WriteByte(']')
	return buf.String()
}
