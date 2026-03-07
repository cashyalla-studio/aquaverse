package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

const imageJobQueue = "aquaverse:image:jobs"

type ImageJob struct {
	JobID     int64  `json:"job_id"`
	ObjectKey string `json:"object_key"`
	Bucket    string `json:"bucket"`
}

type ImageWorker struct {
	db  *sqlx.DB
	rdb *redis.Client
}

func NewImageWorker(db *sqlx.DB, rdb *redis.Client) *ImageWorker {
	return &ImageWorker{db: db, rdb: rdb}
}

// EnqueueImageJob: 이미지 업로드 후 처리 작업 큐에 추가
func (w *ImageWorker) EnqueueImageJob(ctx context.Context, objectKey, bucket string) error {
	// DB에 작업 레코드 생성
	var jobID int64
	err := w.db.QueryRowxContext(ctx,
		`INSERT INTO image_jobs (object_key, bucket, status) VALUES ($1, $2, 'PENDING') RETURNING id`,
		objectKey, bucket,
	).Scan(&jobID)
	if err != nil {
		return fmt.Errorf("create image job: %w", err)
	}

	// Redis 큐에 발행
	job := ImageJob{JobID: jobID, ObjectKey: objectKey, Bucket: bucket}
	data, _ := json.Marshal(job)
	return w.rdb.LPush(ctx, imageJobQueue, data).Err()
}

// StartWorker: 이미지 처리 워커 goroutine (메인 서버 내 실행)
// 실제 이미지 변환은 별도 image-worker 서비스에서 처리.
// 이 워커는 큐 모니터링 + 상태 업데이트 담당.
func (w *ImageWorker) StartWorker(ctx context.Context) {
	slog.Info("image worker started")
	for {
		select {
		case <-ctx.Done():
			slog.Info("image worker stopped")
			return
		default:
			// BRPOP으로 블로킹 대기 (5초 타임아웃)
			result, err := w.rdb.BRPop(ctx, 5*time.Second, imageJobQueue).Result()
			if err != nil {
				continue // 타임아웃 또는 오류 시 재시도
			}
			if len(result) < 2 {
				continue
			}

			var job ImageJob
			if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
				slog.Error("image job unmarshal failed", "err", err)
				continue
			}

			w.processJob(ctx, job)
		}
	}
}

func (w *ImageWorker) processJob(ctx context.Context, job ImageJob) {
	slog.Info("processing image job", "job_id", job.JobID, "key", job.ObjectKey)

	// 상태를 PROCESSING으로
	w.db.ExecContext(ctx,
		`UPDATE image_jobs SET status='PROCESSING' WHERE id=$1`, job.JobID,
	)

	// TODO: 실제 이미지 처리 (WebP 변환, 썸네일 생성)
	// P1에서는 악성 파일 검증만 수행
	if err := w.validateFile(ctx, job); err != nil {
		slog.Error("image validation failed", "job_id", job.JobID, "err", err)
		w.db.ExecContext(ctx,
			`UPDATE image_jobs SET status='FAILED', error_msg=$2, processed_at=NOW() WHERE id=$1`,
			job.JobID, err.Error(),
		)
		return
	}

	w.db.ExecContext(ctx,
		`UPDATE image_jobs SET status='DONE', processed_at=NOW() WHERE id=$1`, job.JobID,
	)
	slog.Info("image job done", "job_id", job.JobID)
}

// validateFile: 파일 타입 검증 (MIME 스니핑)
func (w *ImageWorker) validateFile(ctx context.Context, job ImageJob) error {
	// P1: 기본 검증 — 추후 MinIO에서 파일 읽어서 Magic Bytes 확인
	// 허용 bucket/확장자 정책 확인
	allowedBuckets := map[string]bool{"listings": true, "fish": true, "avatars": true}
	if !allowedBuckets[job.Bucket] {
		return fmt.Errorf("unknown bucket: %s", job.Bucket)
	}
	// TODO: MinIO GetObject → Read first 512 bytes → http.DetectContentType
	return nil
}
