package domain

import (
	"time"
)

// CareLevel 사육 난이도
type CareLevel string

const (
	CareLevelBeginner     CareLevel = "BEGINNER"
	CareLevelIntermediate CareLevel = "INTERMEDIATE"
	CareLevelExpert       CareLevel = "EXPERT"
)

// Temperament 성격
type Temperament string

const (
	TemperamentPeaceful      Temperament = "PEACEFUL"
	TemperamentSemiAggressive Temperament = "SEMI_AGGRESSIVE"
	TemperamentAggressive    Temperament = "AGGRESSIVE"
)

// DietType 식성
type DietType string

const (
	DietOmnivore  DietType = "OMNIVORE"
	DietCarnivore DietType = "CARNIVORE"
	DietHerbivore DietType = "HERBIVORE"
)

// CreatureCategory 생물 카테고리
type CreatureCategory string

const (
	CategoryFish      CreatureCategory = "fish"
	CategoryReptile   CreatureCategory = "reptile"
	CategoryAmphibian CreatureCategory = "amphibian"
	CategoryInsect    CreatureCategory = "insect"
	CategoryArachnid  CreatureCategory = "arachnid"
	CategoryBird      CreatureCategory = "bird"
	CategoryMammal    CreatureCategory = "mammal"
)

// CreatureCategoryInfo 카테고리 정보
type CreatureCategoryInfo struct {
	Code      string `db:"code"       json:"code"`
	NameKo    string `db:"name_ko"    json:"name_ko"`
	NameEn    string `db:"name_en"    json:"name_en"`
	IconEmoji string `db:"icon_emoji" json:"icon_emoji"`
	SortOrder int    `db:"sort_order" json:"sort_order"`
	IsActive  bool   `db:"is_active"  json:"is_active"`
}

// SpeciesExtraAttributes 카테고리별 추가 속성
type SpeciesExtraAttributes struct {
	FishDataID       int64    `db:"fish_data_id"       json:"fish_data_id"`
	CreatureCategory string   `db:"creature_category"  json:"creature_category"`
	HumidityMin      *float64 `db:"humidity_min"       json:"humidity_min,omitempty"`
	HumidityMax      *float64 `db:"humidity_max"       json:"humidity_max,omitempty"`
	UVRequirement    *string  `db:"uv_requirement"     json:"uv_requirement,omitempty"`
	BaskingTempC     *float64 `db:"basking_temp_c"     json:"basking_temp_c,omitempty"`
	SubstrateType    *string  `db:"substrate_type"     json:"substrate_type,omitempty"`
	EnclosureType    *string  `db:"enclosure_type"     json:"enclosure_type,omitempty"`
	ColonySizeMin    *int     `db:"colony_size_min"    json:"colony_size_min,omitempty"`
	ColonySizeMax    *int     `db:"colony_size_max"    json:"colony_size_max,omitempty"`
	QueenRequired    bool     `db:"queen_required"     json:"queen_required,omitempty"`
	VenomLevel       *string  `db:"venom_level"        json:"venom_level,omitempty"`
	LifespanYearsMin *int     `db:"lifespan_years_min" json:"lifespan_years_min,omitempty"`
	LifespanYearsMax *int     `db:"lifespan_years_max" json:"lifespan_years_max,omitempty"`
	AdultSizeCm      *float64 `db:"adult_size_cm"      json:"adult_size_cm,omitempty"`
	LegalStatusKr    string   `db:"legal_status_kr"    json:"legal_status_kr"`
	CITESAppendix    *string  `db:"cites_appendix"     json:"cites_appendix,omitempty"`
	Notes            *string  `db:"notes"              json:"notes,omitempty"`
}

// PublishStatus 게시 상태
type PublishStatus string

const (
	PublishStatusDraft         PublishStatus = "DRAFT"
	PublishStatusAIProcessing  PublishStatus = "AI_PROCESSING"
	PublishStatusTranslating   PublishStatus = "TRANSLATING"
	PublishStatusPublished     PublishStatus = "PUBLISHED"
	PublishStatusRejected      PublishStatus = "REJECTED"
)

// FishData 열대어 백과사전 핵심 모델
type FishData struct {
	ID int64 `db:"id"`

	// 분류학
	ScientificName string  `db:"scientific_name"`
	Genus          string  `db:"genus"`
	Species        string  `db:"species"`
	Family         string  `db:"family"`
	OrderName      string  `db:"order_name"`
	ClassName      string  `db:"class_name"`

	// 기본 정보
	PrimaryCommonName string      `db:"primary_common_name"`
	CareLevel         *CareLevel  `db:"care_level"`
	Temperament       *Temperament `db:"temperament"`

	// 크기/수명
	MaxSizeCm    *float64 `db:"max_size_cm"`
	LifespanYears *float64 `db:"lifespan_years"`

	// 수질 요건
	PHMin          *float64 `db:"ph_min"`
	PHMax          *float64 `db:"ph_max"`
	TempMinC       *float64 `db:"temp_min_c"`
	TempMaxC       *float64 `db:"temp_max_c"`
	HardnessMinDGH *float64 `db:"hardness_min_dgh"`
	HardnessMaxDGH *float64 `db:"hardness_max_dgh"`

	// 사육 정보
	MinTankSizeLiters *int     `db:"min_tank_size_liters"`
	DietType          *DietType `db:"diet_type"`
	DietNotes         *string  `db:"diet_notes"`
	BreedingNotes     *string  `db:"breeding_notes"`
	CareNotes         *string  `db:"care_notes"`

	// 이미지
	PrimaryImageURL *string `db:"primary_image_url"`
	ImageLicense    *string `db:"image_license"`
	ImageAuthor     *string `db:"image_author"`

	// 출처 및 저작권
	PrimarySource string  `db:"primary_source"`
	SourceURL     *string `db:"source_url"`
	License       *string `db:"license"`
	LicenseURL    *string `db:"license_url"`
	Attribution   *string `db:"attribution"`

	// 카테고리
	CreatureCategory string `db:"creature_category" json:"creature_category"`

	// 상태
	PublishStatus PublishStatus `db:"publish_status"`
	QualityScore  float64       `db:"quality_score"`
	AIEnriched    bool          `db:"ai_enriched"`
	AIEnrichedAt  *time.Time    `db:"ai_enriched_at"`

	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	PublishedAt *time.Time `db:"published_at"`

	// 조인 데이터
	Translations []FishTranslation `db:"-"`
	Synonyms     []FishSynonym     `db:"-"`
}

// FishTranslation 다국어 번역
type FishTranslation struct {
	ID              int64     `db:"id"`
	FishDataID      int64     `db:"fish_data_id"`
	Locale          Locale    `db:"locale"`
	CommonName      *string   `db:"common_name"`
	CareLevelLabel  *string   `db:"care_level_label"`
	TemperamentLabel *string  `db:"temperament_label"`
	DietNotes       *string   `db:"diet_notes"`
	BreedingNotes   *string   `db:"breeding_notes"`
	CareNotes       *string   `db:"care_notes"`
	TranslationSource string  `db:"translation_source"` // AI, HUMAN, COMMUNITY
	TranslatedAt    time.Time `db:"translated_at"`
	Verified        bool      `db:"verified"`
}

// FishSynonym 동의어/별명
type FishSynonym struct {
	ID          int64   `db:"id"`
	FishDataID  int64   `db:"fish_data_id"`
	SynonymType string  `db:"synonym_type"` // SCIENTIFIC, COMMON, TRADE
	Name        string  `db:"name"`
	Locale      *Locale `db:"locale"`
	Source      *string `db:"source"`
}

// CrawlJob 크롤링 작업 추적
type CrawlJob struct {
	ID             int64      `db:"id"`
	SourceName     string     `db:"source_name"`
	SourceURL      string     `db:"source_url"`
	JobType        string     `db:"job_type"` // FULL, INCREMENTAL, SINGLE
	Status         string     `db:"status"`   // PENDING, RUNNING, COMPLETED, FAILED
	ItemsFound     int        `db:"items_found"`
	ItemsProcessed int        `db:"items_processed"`
	ItemsFailed    int        `db:"items_failed"`
	ErrorMessage   *string    `db:"error_message"`
	StartedAt      *time.Time `db:"started_at"`
	CompletedAt    *time.Time `db:"completed_at"`
	CreatedAt      time.Time  `db:"created_at"`
}

// RawCrawlData 원시 크롤링 데이터
type RawCrawlData struct {
	ID           int64      `db:"id"`
	CrawlJobID   *int64     `db:"crawl_job_id"`
	SourceName   string     `db:"source_name"`
	SourceID     string     `db:"source_id"`
	SourceURL    *string    `db:"source_url"`
	RawJSON      []byte     `db:"raw_json"`
	ContentHash  string     `db:"content_hash"`
	ParseStatus  string     `db:"parse_status"` // PENDING, PROCESSED, FAILED, SKIPPED
	FishDataID   *int64     `db:"fish_data_id"`
	CrawledAt    time.Time  `db:"crawled_at"`
	ProcessedAt  *time.Time `db:"processed_at"`
}

// FishListResponse API 응답용
type FishListResponse struct {
	ID                int64     `json:"id"`
	ScientificName    string    `json:"scientific_name"`
	CommonName        string    `json:"common_name"`
	Family            string    `json:"family"`
	CareLevel         *string   `json:"care_level,omitempty"`
	Temperament       *string   `json:"temperament,omitempty"`
	MaxSizeCm         *float64  `json:"max_size_cm,omitempty"`
	MinTankSizeLiters *int      `json:"min_tank_size_liters,omitempty"`
	PrimaryImageURL   *string   `json:"primary_image_url,omitempty"`
	QualityScore      float64   `json:"quality_score"`
}

// FishFilter 어종 목록 필터
type FishFilter struct {
	Category  string
	Family    string
	CareLevel string
	Search    string
	Locale    Locale
	Page      int
	Limit     int
}

// FishListResult 어종 목록 결과
type FishListResult struct {
	Items      []FishListResponse `json:"items"`
	TotalCount int                `json:"total_count"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
}
