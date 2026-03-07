package domain

import "time"

type SpeciesCandidate struct {
	Name           string  `json:"name"`
	ScientificName string  `json:"scientific_name"`
	Confidence     float64 `json:"confidence"`  // 0.0 ~ 1.0
	Description    string  `json:"description"` // 특징 설명
	FishDataID     *int64  `json:"fish_data_id,omitempty"` // DB 매칭 ID
}

type SpeciesIdentifyResult struct {
	ID           int64              `json:"id"`
	Candidates   []SpeciesCandidate `json:"candidates"`
	ProcessingMs int                `json:"processing_ms"`
	CreatedAt    time.Time          `json:"created_at"`
}
