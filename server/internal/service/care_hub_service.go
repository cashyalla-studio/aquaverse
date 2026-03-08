package service

import (
	"context"
	"fmt"
	"time"

	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// CareHubService 케어 허브 서비스 - 케어 일정, 기록, 스트릭 관리
type CareHubService struct {
	db *sqlx.DB
}

func NewCareHubService(db *sqlx.DB) *CareHubService {
	return &CareHubService{db: db}
}

// CreateSchedule 케어 일정 생성
func (s *CareHubService) CreateSchedule(ctx context.Context, tankID int64, userID uuid.UUID, req domain.CreateScheduleRequest) (*domain.CareSchedule, error) {
	nextDue, err := time.Parse(time.RFC3339, req.NextDueAt)
	if err != nil {
		return nil, fmt.Errorf("next_due_at 형식 오류 (RFC3339 필요): %w", err)
	}

	var schedule domain.CareSchedule
	err = s.db.QueryRowContext(ctx, `
		INSERT INTO care_schedules
			(tank_id, user_id, schedule_type, title, description, frequency, interval_days, next_due_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, tank_id, user_id, schedule_type, title, description, frequency,
		          interval_days, next_due_at, last_done_at, is_active, created_at
	`,
		tankID, userID, req.ScheduleType, req.Title, req.Description,
		req.Frequency, req.IntervalDays, nextDue,
	).Scan(
		&schedule.ID, &schedule.TankID, &schedule.UserID,
		&schedule.ScheduleType, &schedule.Title, &schedule.Description,
		&schedule.Frequency, &schedule.IntervalDays, &schedule.NextDueAt,
		&schedule.LastDoneAt, &schedule.IsActive, &schedule.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &schedule, nil
}

// ListSchedules 수조의 케어 일정 목록 조회 (is_active=true만)
func (s *CareHubService) ListSchedules(ctx context.Context, tankID int64) ([]domain.CareSchedule, error) {
	var schedules []domain.CareSchedule
	err := s.db.SelectContext(ctx, &schedules, `
		SELECT id, tank_id, user_id, schedule_type, title, description, frequency,
		       interval_days, next_due_at, last_done_at, is_active, created_at
		FROM care_schedules
		WHERE tank_id = $1 AND is_active = TRUE
		ORDER BY next_due_at ASC
	`, tankID)
	if err != nil {
		return nil, err
	}
	return schedules, nil
}

// UpdateSchedule 케어 일정 수정 (소유자 확인 포함)
func (s *CareHubService) UpdateSchedule(ctx context.Context, scheduleID int64, userID uuid.UUID, req domain.UpdateScheduleRequest) (*domain.CareSchedule, error) {
	// 소유자 확인
	var ownerID uuid.UUID
	if err := s.db.QueryRowContext(ctx,
		`SELECT user_id FROM care_schedules WHERE id = $1`, scheduleID,
	).Scan(&ownerID); err != nil {
		return nil, fmt.Errorf("일정을 찾을 수 없습니다")
	}
	if ownerID != userID {
		return nil, fmt.Errorf("수정 권한이 없습니다")
	}

	// 동적 업데이트: 변경된 필드만 반영
	var nextDue *time.Time
	if req.NextDueAt != nil {
		t, err := time.Parse(time.RFC3339, *req.NextDueAt)
		if err != nil {
			return nil, fmt.Errorf("next_due_at 형식 오류: %w", err)
		}
		nextDue = &t
	}

	var schedule domain.CareSchedule
	err := s.db.QueryRowContext(ctx, `
		UPDATE care_schedules
		SET
			title         = COALESCE($1, title),
			description   = COALESCE($2, description),
			frequency     = COALESCE($3, frequency),
			interval_days = COALESCE($4, interval_days),
			next_due_at   = COALESCE($5, next_due_at),
			is_active     = COALESCE($6, is_active)
		WHERE id = $7
		RETURNING id, tank_id, user_id, schedule_type, title, description, frequency,
		          interval_days, next_due_at, last_done_at, is_active, created_at
	`,
		req.Title, req.Description, req.Frequency, req.IntervalDays, nextDue, req.IsActive,
		scheduleID,
	).Scan(
		&schedule.ID, &schedule.TankID, &schedule.UserID,
		&schedule.ScheduleType, &schedule.Title, &schedule.Description,
		&schedule.Frequency, &schedule.IntervalDays, &schedule.NextDueAt,
		&schedule.LastDoneAt, &schedule.IsActive, &schedule.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &schedule, nil
}

// DeleteSchedule 케어 일정 삭제 (소유자 확인 포함)
func (s *CareHubService) DeleteSchedule(ctx context.Context, scheduleID int64, userID uuid.UUID) error {
	result, err := s.db.ExecContext(ctx,
		`DELETE FROM care_schedules WHERE id = $1 AND user_id = $2`,
		scheduleID, userID,
	)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("일정을 찾을 수 없거나 삭제 권한이 없습니다")
	}
	return nil
}

// CompleteSchedule 케어 완료 처리 - 로그 생성, next_due_at 갱신, 스트릭 업데이트
func (s *CareHubService) CompleteSchedule(ctx context.Context, scheduleID int64, userID uuid.UUID, req domain.CompleteScheduleRequest) (*domain.CareLog, error) {
	// 일정 조회
	var schedule domain.CareSchedule
	err := s.db.QueryRowContext(ctx, `
		SELECT id, tank_id, user_id, schedule_type, title, frequency, interval_days, next_due_at, is_active
		FROM care_schedules WHERE id = $1
	`, scheduleID).Scan(
		&schedule.ID, &schedule.TankID, &schedule.UserID,
		&schedule.ScheduleType, &schedule.Title, &schedule.Frequency,
		&schedule.IntervalDays, &schedule.NextDueAt, &schedule.IsActive,
	)
	if err != nil {
		return nil, fmt.Errorf("일정을 찾을 수 없습니다")
	}
	if schedule.UserID != userID {
		return nil, fmt.Errorf("완료 처리 권한이 없습니다")
	}
	if !schedule.IsActive {
		return nil, fmt.Errorf("비활성화된 일정입니다")
	}

	now := time.Now()

	// 다음 due_at 계산
	nextDue := s.calcNextDue(schedule.Frequency, schedule.IntervalDays, now)

	// 트랜잭션으로 케어 로그 생성 + 일정 업데이트 + 스트릭 갱신
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 케어 로그 삽입
	var log domain.CareLog
	err = tx.QueryRowContext(ctx, `
		INSERT INTO care_logs (schedule_id, tank_id, user_id, care_type, notes, photo_url, done_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, schedule_id, tank_id, user_id, care_type, notes, photo_url, done_at
	`,
		scheduleID, schedule.TankID, userID, schedule.ScheduleType,
		req.Notes, req.PhotoURL, now,
	).Scan(
		&log.ID, &log.ScheduleID, &log.TankID, &log.UserID,
		&log.CareType, &log.Notes, &log.PhotoURL, &log.DoneAt,
	)
	if err != nil {
		return nil, err
	}

	// 일정 next_due_at + last_done_at 업데이트
	_, err = tx.ExecContext(ctx, `
		UPDATE care_schedules
		SET last_done_at = $1, next_due_at = $2
		WHERE id = $3
	`, now, nextDue, scheduleID)
	if err != nil {
		return nil, err
	}

	// 스트릭 갱신
	if err := s.updateStreakTx(ctx, tx, userID, now); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &log, nil
}

// calcNextDue 빈도에 따른 다음 due_at 계산
func (s *CareHubService) calcNextDue(frequency string, intervalDays *int, from time.Time) time.Time {
	switch frequency {
	case "daily":
		return from.Add(24 * time.Hour)
	case "weekly":
		return from.Add(7 * 24 * time.Hour)
	case "biweekly":
		return from.Add(14 * 24 * time.Hour)
	case "monthly":
		return from.AddDate(0, 1, 0)
	case "custom":
		days := 1
		if intervalDays != nil && *intervalDays > 0 {
			days = *intervalDays
		}
		return from.Add(time.Duration(days) * 24 * time.Hour)
	default:
		return from.Add(24 * time.Hour)
	}
}

// updateStreakTx 트랜잭션 내에서 스트릭 갱신
func (s *CareHubService) updateStreakTx(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID, now time.Time) error {
	today := now.Format("2006-01-02")

	var streak domain.CareStreak
	err := tx.QueryRowContext(ctx, `
		SELECT user_id, current_streak, longest_streak, last_care_date, updated_at
		FROM care_streaks WHERE user_id = $1
	`, userID).Scan(
		&streak.UserID, &streak.CurrentStreak, &streak.LongestStreak,
		&streak.LastCareDate, &streak.UpdatedAt,
	)

	newCurrent := 1
	if err == nil && streak.LastCareDate != nil {
		last := *streak.LastCareDate
		yesterday := now.Add(-24 * time.Hour).Format("2006-01-02")
		if last == today {
			// 오늘 이미 케어 완료 - 스트릭 유지
			return nil
		} else if last == yesterday {
			// 연속 - 스트릭 증가
			newCurrent = streak.CurrentStreak + 1
		}
		// 그 외: 연속 끊김 - 1로 초기화
	}

	newLongest := newCurrent
	if err == nil && streak.LongestStreak > newLongest {
		newLongest = streak.LongestStreak
	}

	_, upsertErr := tx.ExecContext(ctx, `
		INSERT INTO care_streaks (user_id, current_streak, longest_streak, last_care_date, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (user_id) DO UPDATE
		SET current_streak = $2,
		    longest_streak = $3,
		    last_care_date = $4,
		    updated_at     = NOW()
	`, userID, newCurrent, newLongest, today)
	return upsertErr
}

// GetTodayTasks 오늘 기한인 케어 일정 목록 조회
func (s *CareHubService) GetTodayTasks(ctx context.Context, userID uuid.UUID) ([]domain.CareSchedule, error) {
	var schedules []domain.CareSchedule
	// 오늘 자정부터 내일 자정까지 due인 활성 일정
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	err := s.db.SelectContext(ctx, &schedules, `
		SELECT id, tank_id, user_id, schedule_type, title, description, frequency,
		       interval_days, next_due_at, last_done_at, is_active, created_at
		FROM care_schedules
		WHERE user_id = $1
		  AND is_active = TRUE
		  AND next_due_at >= $2
		  AND next_due_at < $3
		ORDER BY next_due_at ASC
	`, userID, startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}
	return schedules, nil
}

// GetStreak 사용자 스트릭 조회
func (s *CareHubService) GetStreak(ctx context.Context, userID uuid.UUID) (*domain.CareStreak, error) {
	var streak domain.CareStreak
	err := s.db.QueryRowContext(ctx, `
		SELECT user_id, current_streak, longest_streak, last_care_date, updated_at
		FROM care_streaks WHERE user_id = $1
	`, userID).Scan(
		&streak.UserID, &streak.CurrentStreak, &streak.LongestStreak,
		&streak.LastCareDate, &streak.UpdatedAt,
	)
	if err != nil {
		// 케어 기록이 없는 신규 사용자 - 빈 스트릭 반환
		return &domain.CareStreak{
			UserID:        userID,
			CurrentStreak: 0,
			LongestStreak: 0,
		}, nil
	}
	return &streak, nil
}
