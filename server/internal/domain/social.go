package domain

import "time"

type UserFollow struct {
	FollowerID  string    `json:"follower_id" db:"follower_id"`
	FollowingID string    `json:"following_id" db:"following_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type ActivityItem struct {
	ID         int64     `json:"id" db:"id"`
	ActorID    string    `json:"actor_id" db:"actor_id"`
	ActorName  string    `json:"actor_name" db:"actor_name"`
	Verb       string    `json:"verb" db:"verb"`
	ObjectType string    `json:"object_type" db:"object_type"`
	ObjectID   *int64    `json:"object_id" db:"object_id"`
	ObjectData []byte    `json:"object_data" db:"object_data"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type FollowSuggestion struct {
	UserID     string  `json:"user_id" db:"user_id"`
	Username   string  `json:"username" db:"username"`
	TrustScore float64 `json:"trust_score" db:"trust_score"`
	CommonFish string  `json:"common_fish" db:"common_fish"` // 공통 관심 어종
}
