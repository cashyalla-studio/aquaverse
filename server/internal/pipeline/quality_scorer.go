package pipeline

import "github.com/cashyalla/aquaverse/internal/domain"

// QualityScorer 데이터 품질 점수 계산기
// coin-trading heatscore.go 패턴 재활용: 가중치 합산 + 정규화

// 품질 등급
const (
	QualityExcellent = 85.0 // 즉시 번역/게시
	QualityGood      = 70.0 // 번역 후 게시
	QualityFair      = 60.0 // AI 재가공 권장
	QualityPoor      = 40.0 // AI 재가공 필수
	// < 40: REJECT
)

// 필드별 가중치 (합계 = 100)
var fieldWeights = map[string]float64{
	"scientific_name":       10.0,
	"primary_common_name":    8.0,
	"family":                 5.0,
	"care_level":             8.0,
	"temperament":            7.0,
	"max_size_cm":            6.0,
	"min_tank_size_liters":   7.0,
	"ph_min":                 5.0,
	"ph_max":                 5.0,
	"temp_min_c":             5.0,
	"temp_max_c":             5.0,
	"diet_type":              6.0,
	"diet_notes":             4.0,
	"breeding_notes":         5.0,
	"care_notes":             6.0,
	"primary_image_url":      8.0,
	"lifespan_years":         3.0,
	"license":                2.0,
}

// 페널티
type penalty struct {
	condition func(*domain.FishData) bool
	amount    float64
}

var penalties = []penalty{
	{
		// pH 범위가 너무 넓음 (> 3)
		condition: func(f *domain.FishData) bool {
			return f.PHMin != nil && f.PHMax != nil && (*f.PHMax-*f.PHMin) > 3
		},
		amount: -5.0,
	},
	{
		// 온도 범위 불일치
		condition: func(f *domain.FishData) bool {
			return f.TempMinC != nil && f.TempMaxC != nil && *f.TempMaxC <= *f.TempMinC
		},
		amount: -3.0,
	},
	{
		// care_notes 너무 짧음
		condition: func(f *domain.FishData) bool {
			return f.CareNotes != nil && len(*f.CareNotes) < 50
		},
		amount: -3.0,
	},
}

// ScoreResult 품질 점수 결과
type ScoreResult struct {
	Score         float64            `json:"score"`
	Grade         string             `json:"grade"`
	FieldScores   map[string]float64 `json:"field_scores"`
	MissingFields []string           `json:"missing_fields"`
	Issues        []string           `json:"issues"`
	NeedsAI       bool               `json:"needs_ai"`
	ShouldReject  bool               `json:"should_reject"`
}

func ScoreFish(fish *domain.FishData) ScoreResult {
	result := ScoreResult{
		FieldScores: make(map[string]float64),
	}

	score := 0.0

	// 필드 존재 여부 점수
	check := map[string]bool{
		"scientific_name":       fish.ScientificName != "",
		"primary_common_name":   fish.PrimaryCommonName != "",
		"family":                fish.Family != "",
		"care_level":            fish.CareLevel != nil,
		"temperament":           fish.Temperament != nil,
		"max_size_cm":           fish.MaxSizeCm != nil,
		"min_tank_size_liters":  fish.MinTankSizeLiters != nil,
		"ph_min":                fish.PHMin != nil,
		"ph_max":                fish.PHMax != nil,
		"temp_min_c":            fish.TempMinC != nil,
		"temp_max_c":            fish.TempMaxC != nil,
		"diet_type":             fish.DietType != nil,
		"diet_notes":            fish.DietNotes != nil && *fish.DietNotes != "",
		"breeding_notes":        fish.BreedingNotes != nil && *fish.BreedingNotes != "",
		"care_notes":            fish.CareNotes != nil && *fish.CareNotes != "",
		"primary_image_url":     fish.PrimaryImageURL != nil,
		"lifespan_years":        fish.LifespanYears != nil,
		"license":               fish.License != nil && *fish.License != "",
	}

	for field, present := range check {
		w := fieldWeights[field]
		if present {
			result.FieldScores[field] = w
			score += w
		} else {
			result.FieldScores[field] = 0
			result.MissingFields = append(result.MissingFields, field)
		}
	}

	// 페널티 적용
	for _, p := range penalties {
		if p.condition(fish) {
			score += p.amount
			result.Issues = append(result.Issues, "penalty applied")
		}
	}

	if score < 0 {
		score = 0
	}

	result.Score = score
	result.Grade = gradeFromScore(score)
	result.NeedsAI = score < QualityFair
	result.ShouldReject = score < QualityPoor

	return result
}

func gradeFromScore(score float64) string {
	switch {
	case score >= QualityExcellent:
		return "EXCELLENT"
	case score >= QualityGood:
		return "GOOD"
	case score >= QualityFair:
		return "FAIR"
	case score >= QualityPoor:
		return "POOR"
	default:
		return "REJECT"
	}
}
