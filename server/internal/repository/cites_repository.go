package repository

import (
	"context"
	"database/sql"
	"strings"

	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/jmoiron/sqlx"
)

type CitesRepository struct {
	db *sqlx.DB
}

func NewCitesRepository(db *sqlx.DB) *CitesRepository {
	return &CitesRepository{db: db}
}

// CheckScientificName: 학명으로 CITES 체크
func (r *CitesRepository) CheckScientificName(ctx context.Context, scientificName string) (*domain.CitesCheckResult, error) {
	result := &domain.CitesCheckResult{}
	normalized := strings.TrimSpace(scientificName)
	if normalized == "" {
		return result, nil
	}

	// 정확한 학명 매칭 (대소문자 무관)
	var fish domain.CitesFish
	err := r.db.GetContext(ctx, &fish,
		`SELECT * FROM cites_fish WHERE LOWER(scientific_name) = LOWER($1)`,
		normalized,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == nil {
		result.Appendix = fish.Appendix
		result.IsBlocked = fish.IsBlocked
		result.HasWarning = true
		if fish.IsBlocked {
			result.Message = "이 어종은 CITES 부속서 " + string(fish.Appendix) + "에 등재되어 거래가 제한됩니다."
		} else {
			result.Message = "이 어종은 CITES 부속서 " + string(fish.Appendix) + "에 등재된 보호종입니다. 합법적 출처를 확인해 주세요."
		}
	}

	// 한국 생태계 교란종 체크
	var invasiveCount int
	r.db.QueryRowxContext(ctx,
		`SELECT COUNT(*) FROM invasive_species_kr WHERE LOWER(scientific_name) = LOWER($1)`,
		normalized,
	).Scan(&invasiveCount)
	if invasiveCount > 0 {
		result.IsInvasiveKR = true
		result.HasWarning = true
		if result.Message == "" {
			result.Message = "이 어종은 한국 생태계 교란종으로 지정되어 있습니다. 자연 방류는 불법입니다."
		}
	}

	return result, nil
}
