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
