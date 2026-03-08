package crawler

import (
	"context"

	"github.com/cashyalla/aquaverse/internal/domain"
)

// SourceAdapter 데이터 수집 소스 인터페이스.
// 각 외부 데이터 소스(FishBase, GBIF, Wikidata 등)는 이 인터페이스를 구현한다.
type SourceAdapter interface {
	// Name 소스 식별자 (raw_crawl_data.source_name에 저장됨)
	Name() string

	// Crawl 데이터를 수집하고 save 콜백을 통해 저장한다.
	// save가 호출될 때마다 하나의 RawCrawlData 레코드가 저장된다.
	// 반환값 fetched: 실제로 저장 시도한 레코드 수
	// API 연결 실패 시 (0, nil) 을 반환해 graceful degradation을 지원한다.
	Crawl(ctx context.Context, jobID int64, save func(*domain.RawCrawlData) error) (fetched int, err error)
}
