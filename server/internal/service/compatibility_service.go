package service

import (
	"context"
	"fmt"
	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/jmoiron/sqlx"
)

type CompatibilityService struct {
	db *sqlx.DB
}

func NewCompatibilityService(db *sqlx.DB) *CompatibilityService {
	return &CompatibilityService{db: db}
}

// CheckCompatibility 두 어종 간 호환성 체크
func (s *CompatibilityService) CheckCompatibility(ctx context.Context, fishAID, fishBID int64) (*domain.CompatibilityResult, error) {
	// fish_a_id < fish_b_id 보장
	if fishAID > fishBID {
		fishAID, fishBID = fishBID, fishAID
	}

	var fc domain.FishCompatibility
	err := s.db.GetContext(ctx, &fc,
		`SELECT fish_a_id, fish_b_id, compatible, caution, COALESCE(reason,'') as reason
         FROM fish_compatibility WHERE fish_a_id=$1 AND fish_b_id=$2`,
		fishAID, fishBID,
	)
	if err == nil {
		return &domain.CompatibilityResult{
			Compatible: fc.Compatible,
			Caution:    fc.Caution,
			Reason:     fc.Reason,
			Source:     "database",
		}, nil
	}

	// DB에 없으면 Rule-based fallback: fish_data의 수질 파라미터 비교
	return s.ruleBasedCheck(ctx, fishAID, fishBID)
}

func (s *CompatibilityService) ruleBasedCheck(ctx context.Context, fishAID, fishBID int64) (*domain.CompatibilityResult, error) {
	type fishParams struct {
		ID          int64   `db:"id"`
		TempMin     float64 `db:"temp_min"`
		TempMax     float64 `db:"temp_max"`
		PhMin       float64 `db:"ph_min"`
		PhMax       float64 `db:"ph_max"`
		Temperament string  `db:"temperament"`
	}

	var fish [2]fishParams
	ids := []int64{fishAID, fishBID}
	for i, id := range ids {
		if err := s.db.GetContext(ctx, &fish[i],
			`SELECT id,
                    COALESCE(temp_min,20) as temp_min, COALESCE(temp_max,28) as temp_max,
                    COALESCE(ph_min,6.5) as ph_min, COALESCE(ph_max,7.5) as ph_max,
                    COALESCE(temperament,'peaceful') as temperament
             FROM fish_data WHERE id=$1`, id); err != nil {
			return &domain.CompatibilityResult{Compatible: true, Source: "rule", Reason: "데이터 부족으로 판단 불가"}, nil
		}
	}

	// 수온 범위 겹침 확인
	tempOverlapMin := max64(fish[0].TempMin, fish[1].TempMin)
	tempOverlapMax := min64(fish[0].TempMax, fish[1].TempMax)
	if tempOverlapMin > tempOverlapMax {
		return &domain.CompatibilityResult{
			Compatible: false,
			Source:     "rule",
			Reason:     fmt.Sprintf("수온 범위 불일치 (A: %.0f-%.0f°C, B: %.0f-%.0f°C)", fish[0].TempMin, fish[0].TempMax, fish[1].TempMin, fish[1].TempMax),
		}, nil
	}

	// pH 범위 겹침 확인
	phOverlapMin := max64(fish[0].PhMin, fish[1].PhMin)
	phOverlapMax := min64(fish[0].PhMax, fish[1].PhMax)
	if phOverlapMin > phOverlapMax {
		return &domain.CompatibilityResult{
			Compatible: false,
			Source:     "rule",
			Reason:     fmt.Sprintf("pH 범위 불일치 (A: %.1f-%.1f, B: %.1f-%.1f)", fish[0].PhMin, fish[0].PhMax, fish[1].PhMin, fish[1].PhMax),
		}, nil
	}

	// 기질 체크
	caution := false
	reason := ""
	if fish[0].Temperament == "aggressive" || fish[1].Temperament == "aggressive" {
		caution = true
		reason = "공격적 기질 어종 포함 — 합사 주의"
	}

	return &domain.CompatibilityResult{
		Compatible: true,
		Caution:    caution,
		Source:     "rule",
		Reason:     reason,
	}, nil
}

// GetCompatibleFish 특정 어종과 합사 가능한 어종 목록
func (s *CompatibilityService) GetCompatibleFish(ctx context.Context, fishID int64) ([]domain.FishRecommendation, error) {
	rows, err := s.db.QueryxContext(ctx, `
        SELECT
            fd.id as fish_id,
            fd.common_name as fish_name,
            COALESCE(fd.scientific_name,'') as scientific_name,
            COALESCE(fd.image_url,'') as image_url,
            fc.reason
        FROM fish_compatibility fc
        JOIN fish_data fd ON (
            CASE WHEN fc.fish_a_id = $1 THEN fc.fish_b_id ELSE fc.fish_a_id END = fd.id
        )
        WHERE (fc.fish_a_id = $1 OR fc.fish_b_id = $1) AND fc.compatible = true
        ORDER BY fc.caution ASC, fd.common_name ASC
        LIMIT 50
    `, fishID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.FishRecommendation
	for rows.Next() {
		var r domain.FishRecommendation
		if err := rows.StructScan(&r); err != nil {
			continue
		}
		result = append(result, r)
	}
	return result, nil
}

// RecommendForTank 수조 기반 추천 어종
func (s *CompatibilityService) RecommendForTank(ctx context.Context, tankID int64) ([]domain.FishRecommendation, error) {
	// 수조의 현재 어종 목록 가져오기
	var currentFishIDs []int64
	if err := s.db.SelectContext(ctx, &currentFishIDs,
		`SELECT fish_data_id FROM tank_inhabitants WHERE tank_id=$1`, tankID); err != nil {
		return nil, err
	}

	if len(currentFishIDs) == 0 {
		// 수조가 비어있으면 인기 어종 반환
		return s.getPopularFish(ctx)
	}

	// 수조 어종들과 호환되는 어종 중 현재 없는 것 추천
	type fishScore struct {
		FishID     int64
		FishName   string
		ScientName string
		ImageURL   string
		Score      float64
		Reason     string
	}

	scores := map[int64]*fishScore{}

	for _, cid := range currentFishIDs {
		recs, err := s.GetCompatibleFish(ctx, cid)
		if err != nil {
			continue
		}
		for _, r := range recs {
			// 이미 수조에 있는 어종 제외
			inTank := false
			for _, tid := range currentFishIDs {
				if r.FishID == tid {
					inTank = true
					break
				}
			}
			if inTank {
				continue
			}

			if _, exists := scores[r.FishID]; !exists {
				scores[r.FishID] = &fishScore{
					FishID:     r.FishID,
					FishName:   r.FishName,
					ScientName: r.ScientName,
					ImageURL:   r.ImageURL,
					Reason:     r.Reason,
				}
			}
			scores[r.FishID].Score++
		}
	}

	var result []domain.FishRecommendation
	for _, fs := range scores {
		result = append(result, domain.FishRecommendation{
			FishID:     fs.FishID,
			FishName:   fs.FishName,
			ScientName: fs.ScientName,
			ImageURL:   fs.ImageURL,
			Score:      fs.Score,
			Reason:     fmt.Sprintf("현재 수조 %d종과 호환", int(fs.Score)),
		})
	}

	// 점수 내림차순 정렬
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].Score > result[i].Score {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	if len(result) > 20 {
		result = result[:20]
	}
	return result, nil
}

func (s *CompatibilityService) getPopularFish(ctx context.Context) ([]domain.FishRecommendation, error) {
	rows, err := s.db.QueryxContext(ctx, `
        SELECT id as fish_id, common_name as fish_name,
               COALESCE(scientific_name,'') as scientific_name,
               COALESCE(image_url,'') as image_url,
               0.0 as score, '인기 어종' as reason
        FROM fish_data
        ORDER BY id ASC
        LIMIT 10
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.FishRecommendation
	for rows.Next() {
		var r domain.FishRecommendation
		if err := rows.StructScan(&r); err != nil {
			continue
		}
		result = append(result, r)
	}
	return result, nil
}

func max64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
