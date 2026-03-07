-- 012_fts.up.sql
-- PostgreSQL Full Text Search for fish_data

-- 확장 설치
CREATE EXTENSION IF NOT EXISTS unaccent;

-- pg_bigm은 Docker PostgreSQL에 pre-installed되지 않을 수 있음 — 안전하게 SKIP
-- CREATE EXTENSION IF NOT EXISTS pg_bigm;

-- fish_data에 tsvector 컬럼 추가
ALTER TABLE fish_data ADD COLUMN IF NOT EXISTS search_vector tsvector;

-- 기존 데이터로 tsvector 업데이트
UPDATE fish_data SET search_vector =
  to_tsvector('simple',
    coalesce(unaccent(primary_common_name), '') || ' ' ||
    coalesce(unaccent(scientific_name), '') || ' ' ||
    coalesce(unaccent(family), '') || ' ' ||
    coalesce(unaccent(care_notes), '')
  );

-- GIN 인덱스 생성
CREATE INDEX IF NOT EXISTS idx_fish_search_vector ON fish_data USING GIN(search_vector);

-- 트리거 함수: INSERT/UPDATE 시 자동 갱신
CREATE OR REPLACE FUNCTION update_fish_search_vector()
RETURNS TRIGGER AS $$
BEGIN
  NEW.search_vector := to_tsvector('simple',
    coalesce(unaccent(NEW.primary_common_name), '') || ' ' ||
    coalesce(unaccent(NEW.scientific_name), '') || ' ' ||
    coalesce(unaccent(NEW.family), '') || ' ' ||
    coalesce(unaccent(NEW.care_notes), '')
  );
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_fish_search_vector ON fish_data;
CREATE TRIGGER trg_fish_search_vector
  BEFORE INSERT OR UPDATE ON fish_data
  FOR EACH ROW EXECUTE FUNCTION update_fish_search_vector();

-- 어종명 synonym 딕셔너리는 파일 기반이라 Docker 환경에서 생략 (unaccent으로 대체)
