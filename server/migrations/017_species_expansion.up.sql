-- Add creature_category to fish_data table (extending existing structure)
ALTER TABLE fish_data ADD COLUMN IF NOT EXISTS creature_category VARCHAR(30) NOT NULL DEFAULT 'fish';

-- Index for filtering by category
CREATE INDEX IF NOT EXISTS idx_fish_data_creature_category ON fish_data(creature_category);

-- Category-specific extra attributes table
CREATE TABLE IF NOT EXISTS species_extra_attributes (
    fish_data_id     BIGINT PRIMARY KEY REFERENCES fish_data(id) ON DELETE CASCADE,
    creature_category VARCHAR(30) NOT NULL DEFAULT 'fish',

    -- Reptile/Amphibian fields
    humidity_min      DECIMAL(5,2),
    humidity_max      DECIMAL(5,2),
    uv_requirement    VARCHAR(20),    -- 'none', 'low', 'medium', 'high', 'very_high'
    basking_temp_c    DECIMAL(5,2),   -- basking spot temperature
    substrate_type    VARCHAR(100),   -- 'soil', 'sand', 'coconut_fiber', 'paper_towel', etc.
    enclosure_type    VARCHAR(50),    -- 'terrarium', 'aquaterrarium', 'paludarium', 'vivarium'

    -- Insect/Arachnid fields
    colony_size_min   INT,
    colony_size_max   INT,
    queen_required    BOOLEAN DEFAULT FALSE,
    venom_level       VARCHAR(20),    -- 'none', 'mild', 'moderate', 'dangerous'

    -- General exotic pet fields
    lifespan_years_min INT,
    lifespan_years_max INT,
    adult_size_cm      DECIMAL(6,2),

    -- Legal status (Korea)
    legal_status_kr   VARCHAR(30) DEFAULT 'legal', -- 'legal', 'cites_appendix1', 'cites_appendix2', 'restricted', 'banned'
    cites_appendix    VARCHAR(5),                   -- 'I', 'II', 'III'

    notes             TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Creature category reference table
CREATE TABLE IF NOT EXISTS creature_categories (
    id          SERIAL PRIMARY KEY,
    code        VARCHAR(30) UNIQUE NOT NULL,  -- 'fish', 'reptile', 'amphibian', 'insect', 'arachnid', 'bird', 'mammal'
    name_ko     VARCHAR(50) NOT NULL,
    name_en     VARCHAR(50) NOT NULL,
    icon_emoji  VARCHAR(10),
    sort_order  INT DEFAULT 0,
    is_active   BOOLEAN DEFAULT TRUE
);

INSERT INTO creature_categories (code, name_ko, name_en, icon_emoji, sort_order) VALUES
    ('fish',       '열대어',   'Tropical Fish',  '🐠', 1),
    ('reptile',    '파충류',   'Reptiles',       '🦎', 2),
    ('amphibian',  '양서류',   'Amphibians',     '🐸', 3),
    ('insect',     '곤충',     'Insects',        '🐜', 4),
    ('arachnid',   '거미류',   'Arachnids',      '🕷️', 5),
    ('bird',       '조류',     'Birds',          '🦜', 6),
    ('mammal',     '소동물',   'Small Mammals',  '🐹', 7)
ON CONFLICT (code) DO NOTHING;
