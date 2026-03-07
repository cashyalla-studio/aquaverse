package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/cashyalla/aquaverse/internal/service"
	"github.com/jmoiron/sqlx"
)

type FishRepository struct {
	db *sqlx.DB
}

func NewFishRepository(db *sqlx.DB) *FishRepository {
	return &FishRepository{db: db}
}

func (r *FishRepository) List(ctx context.Context, filter service.FishFilter) ([]domain.FishListResponse, int, error) {
	args := []interface{}{}
	where := []string{"publish_status = 'PUBLISHED'"}
	idx := 1

	if filter.Family != "" {
		where = append(where, fmt.Sprintf("family = $%d", idx))
		args = append(args, filter.Family)
		idx++
	}
	if filter.CareLevel != "" {
		where = append(where, fmt.Sprintf("care_level = $%d", idx))
		args = append(args, filter.CareLevel)
		idx++
	}
	if filter.Search != "" {
		where = append(where, fmt.Sprintf(
			"(to_tsvector('simple', coalesce(scientific_name,'') || ' ' || coalesce(primary_common_name,'')) @@ plainto_tsquery('simple', $%d) OR scientific_name ILIKE $%d OR primary_common_name ILIKE $%d)",
			idx, idx+1, idx+2,
		))
		like := "%" + filter.Search + "%"
		args = append(args, filter.Search, like, like)
		idx += 3
	}

	whereClause := "WHERE " + strings.Join(where, " AND ")

	// 카운트
	var total int
	countQ := fmt.Sprintf("SELECT COUNT(*) FROM fish_data %s", whereClause)
	if err := r.db.GetContext(ctx, &total, countQ, args...); err != nil {
		return nil, 0, err
	}

	// 페이지네이션
	offset := (filter.Page - 1) * filter.Limit

	// 번역 조인 (locale=$1 고정, 나머지 필터는 $2~)
	// locale을 args 맨 앞에 배치하고 WHERE 절 파라미터 인덱스를 $2부터 시작하도록 재조정
	localeArg := string(filter.Locale)
	finalArgs := []interface{}{localeArg}
	finalArgs = append(finalArgs, args...)

	// LIMIT/OFFSET 인덱스: finalArgs 현재 길이 + 1, + 2
	limitIdx := len(finalArgs) + 1
	offsetIdx := len(finalArgs) + 2
	finalArgs = append(finalArgs, filter.Limit, offset)

	// whereClause의 파라미터 인덱스($1~)를 $2~로 쉬프트 (locale이 $1을 차지)
	shiftedWhere := shiftParamIndices(whereClause, 1)

	q := fmt.Sprintf(`
		SELECT
			f.id,
			f.scientific_name,
			COALESCE(t.common_name, f.primary_common_name) AS common_name,
			COALESCE(f.family, '') AS family,
			f.care_level,
			f.temperament,
			f.max_size_cm,
			f.min_tank_size_liters,
			f.primary_image_url,
			f.quality_score
		FROM fish_data f
		LEFT JOIN fish_translations t ON t.fish_data_id = f.id AND t.locale = $1
		%s
		ORDER BY f.quality_score DESC
		LIMIT $%d OFFSET $%d
	`, shiftedWhere, limitIdx, offsetIdx)

	var items []domain.FishListResponse
	if err := r.db.SelectContext(ctx, &items, q, finalArgs...); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *FishRepository) GetByID(ctx context.Context, id int64, locale domain.Locale) (*domain.FishData, error) {
	var fish domain.FishData
	q := `SELECT * FROM fish_data WHERE id = $1 AND publish_status = 'PUBLISHED'`
	if err := r.db.GetContext(ctx, &fish, q, id); err != nil {
		return nil, err
	}

	// 번역 로드
	var translations []domain.FishTranslation
	tq := `SELECT * FROM fish_translations WHERE fish_data_id = $1 AND locale = $2`
	if err := r.db.SelectContext(ctx, &translations, tq, id, string(locale)); err == nil {
		fish.Translations = translations
	}

	// 동의어 로드
	var synonyms []domain.FishSynonym
	sq := `SELECT * FROM fish_synonyms WHERE fish_data_id = $1 ORDER BY synonym_type`
	if err := r.db.SelectContext(ctx, &synonyms, sq, id); err == nil {
		fish.Synonyms = synonyms
	}

	return &fish, nil
}

func (r *FishRepository) Search(ctx context.Context, query string, locale domain.Locale) ([]domain.FishListResponse, error) {
	// FTS 우선: search_vector @@ tsquery
	ftsQ := `
		SELECT
			f.id,
			f.scientific_name,
			COALESCE(t.common_name, f.primary_common_name) AS common_name,
			COALESCE(f.family, '') AS family,
			f.care_level,
			f.temperament,
			f.max_size_cm,
			f.min_tank_size_liters,
			f.primary_image_url,
			f.quality_score
		FROM fish_data f
		LEFT JOIN fish_translations t ON t.fish_data_id = f.id AND t.locale = $1
		WHERE f.publish_status = 'PUBLISHED'
		  AND f.search_vector @@ to_tsquery('simple', unaccent($2) || ':*')
		ORDER BY
			ts_rank(f.search_vector, to_tsquery('simple', unaccent($2) || ':*')) DESC,
			f.quality_score DESC
		LIMIT 30
	`
	var items []domain.FishListResponse
	err := r.db.SelectContext(ctx, &items, ftsQ, string(locale), query)
	if err == nil && len(items) > 0 {
		return items, nil
	}

	// ILIKE fallback: FTS 결과가 없거나 search_vector 컬럼이 아직 없을 때
	like := "%" + query + "%"
	fallbackQ := `
		SELECT
			f.id,
			f.scientific_name,
			COALESCE(t.common_name, f.primary_common_name) AS common_name,
			COALESCE(f.family, '') AS family,
			f.care_level,
			f.temperament,
			f.max_size_cm,
			f.min_tank_size_liters,
			f.primary_image_url,
			f.quality_score
		FROM fish_data f
		LEFT JOIN fish_translations t ON t.fish_data_id = f.id AND t.locale = $1
		LEFT JOIN fish_synonyms s ON s.fish_data_id = f.id
		WHERE f.publish_status = 'PUBLISHED'
		  AND (
			  f.scientific_name ILIKE $2
			  OR f.primary_common_name ILIKE $2
			  OR t.common_name ILIKE $2
			  OR s.name ILIKE $2
		  )
		GROUP BY f.id, t.common_name
		ORDER BY
			CASE WHEN f.scientific_name ILIKE $3 THEN 0
			     WHEN f.primary_common_name ILIKE $3 THEN 1
			     ELSE 2 END,
			f.quality_score DESC
		LIMIT 30
	`
	items = nil
	if err2 := r.db.SelectContext(ctx, &items, fallbackQ, string(locale), like, query); err2 != nil {
		return nil, err2
	}
	return items, nil
}

func (r *FishRepository) ListFamilies(ctx context.Context) ([]string, error) {
	var families []string
	q := `SELECT DISTINCT family FROM fish_data WHERE publish_status = 'PUBLISHED' AND family IS NOT NULL ORDER BY family`
	if err := r.db.SelectContext(ctx, &families, q); err != nil {
		return nil, err
	}
	return families, nil
}

// Save 크롤러/파이프라인에서 사용
func (r *FishRepository) Save(ctx context.Context, fish *domain.FishData) error {
	q := `
		INSERT INTO fish_data (
			scientific_name, genus, species, family, order_name, class_name,
			primary_common_name, care_level, temperament,
			max_size_cm, lifespan_years,
			ph_min, ph_max, temp_min_c, temp_max_c, hardness_min_dgh, hardness_max_dgh,
			min_tank_size_liters, diet_type, diet_notes, breeding_notes, care_notes,
			primary_image_url, image_license, image_author,
			primary_source, source_url, license, license_url, attribution,
			publish_status, quality_score, ai_enriched
		) VALUES (
			:scientific_name, :genus, :species, :family, :order_name, :class_name,
			:primary_common_name, :care_level, :temperament,
			:max_size_cm, :lifespan_years,
			:ph_min, :ph_max, :temp_min_c, :temp_max_c, :hardness_min_dgh, :hardness_max_dgh,
			:min_tank_size_liters, :diet_type, :diet_notes, :breeding_notes, :care_notes,
			:primary_image_url, :image_license, :image_author,
			:primary_source, :source_url, :license, :license_url, :attribution,
			:publish_status, :quality_score, :ai_enriched
		)
		ON CONFLICT (scientific_name) DO UPDATE SET
			quality_score   = EXCLUDED.quality_score,
			publish_status  = EXCLUDED.publish_status,
			care_level      = COALESCE(EXCLUDED.care_level, fish_data.care_level),
			temperament     = COALESCE(EXCLUDED.temperament, fish_data.temperament),
			diet_notes      = COALESCE(EXCLUDED.diet_notes, fish_data.diet_notes),
			breeding_notes  = COALESCE(EXCLUDED.breeding_notes, fish_data.breeding_notes),
			care_notes      = COALESCE(EXCLUDED.care_notes, fish_data.care_notes),
			ai_enriched     = EXCLUDED.ai_enriched,
			updated_at      = NOW()
		RETURNING id
	`
	rows, err := r.db.NamedQueryContext(ctx, q, fish)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&fish.ID)
	}
	return nil
}

func (r *FishRepository) SaveTranslation(ctx context.Context, t *domain.FishTranslation) error {
	q := `
		INSERT INTO fish_translations (fish_data_id, locale, common_name, care_level_label, temperament_label, diet_notes, breeding_notes, care_notes, translation_source)
		VALUES (:fish_data_id, :locale, :common_name, :care_level_label, :temperament_label, :diet_notes, :breeding_notes, :care_notes, :translation_source)
		ON CONFLICT (fish_data_id, locale) DO UPDATE SET
			common_name       = COALESCE(EXCLUDED.common_name, fish_translations.common_name),
			diet_notes        = COALESCE(EXCLUDED.diet_notes, fish_translations.diet_notes),
			breeding_notes    = COALESCE(EXCLUDED.breeding_notes, fish_translations.breeding_notes),
			care_notes        = COALESCE(EXCLUDED.care_notes, fish_translations.care_notes),
			translation_source = EXCLUDED.translation_source,
			translated_at     = NOW()
	`
	_, err := r.db.NamedQueryContext(ctx, q, t)
	return err
}

func (r *FishRepository) SaveRawCrawlData(ctx context.Context, data *domain.RawCrawlData) error {
	q := `
		INSERT INTO raw_crawl_data (crawl_job_id, source_name, source_id, source_url, raw_json, content_hash, parse_status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (source_name, source_id) DO UPDATE SET
			raw_json     = EXCLUDED.raw_json,
			content_hash = EXCLUDED.content_hash,
			parse_status = 'PENDING',
			crawled_at   = NOW()
		RETURNING id
	`
	return r.db.QueryRowContext(ctx, q,
		data.CrawlJobID, data.SourceName, data.SourceID, data.SourceURL,
		data.RawJSON, data.ContentHash, data.ParseStatus,
	).Scan(&data.ID)
}

func (r *FishRepository) ListPendingCrawlData(ctx context.Context, limit int) ([]domain.RawCrawlData, error) {
	var items []domain.RawCrawlData
	q := `SELECT * FROM raw_crawl_data WHERE parse_status = 'PENDING' ORDER BY crawled_at ASC LIMIT $1`
	return items, r.db.SelectContext(ctx, &items, q, limit)
}

func (r *FishRepository) ListLowQualityFish(ctx context.Context, maxScore float64, limit int) ([]domain.FishData, error) {
	var items []domain.FishData
	q := `SELECT * FROM fish_data WHERE quality_score < $1 AND publish_status != 'REJECTED' ORDER BY quality_score ASC LIMIT $2`
	return items, r.db.SelectContext(ctx, &items, q, maxScore, limit)
}

func (r *FishRepository) UpdatePublishStatus(ctx context.Context, id int64, status domain.PublishStatus, score float64) error {
	q := `UPDATE fish_data SET publish_status = $1, quality_score = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.db.ExecContext(ctx, q, string(status), score, id)
	return err
}

// shiftParamIndices WHERE 절의 PostgreSQL 파라미터 인덱스($N)를 shift만큼 증가시킨다.
// locale을 $1로 고정한 뒤 기존 WHERE 절의 $1~을 $2~로 밀기 위해 사용한다.
func shiftParamIndices(clause string, shift int) string {
	// 뒤에서부터 처리해야 $9 -> $10 치환 시 $1을 재치환하는 문제를 방지한다.
	// 간단하게 최대 인덱스(99)부터 1까지 역순 치환한다.
	result := clause
	for i := 99; i >= 1; i-- {
		old := "$" + strconv.Itoa(i)
		new := "$" + strconv.Itoa(i+shift)
		result = strings.ReplaceAll(result, old, new)
	}
	return result
}
