-- Xiangqi Lab MySQL 8.x baseline.
-- The current runnable development adapter is in-memory; this schema defines
-- the production persistence contract and its optimistic-lock constraints.

CREATE TABLE users (
    id BINARY(16) PRIMARY KEY,
    display_name VARCHAR(120) NOT NULL,
    created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE difficulty_profiles (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    level TINYINT UNSIGNED NOT NULL,
    profile_version INT UNSIGNED NOT NULL,
    name VARCHAR(80) NOT NULL,
    config JSON NOT NULL,
    active BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    UNIQUE KEY uq_difficulty_level_version (level, profile_version),
    KEY idx_difficulty_active (active, level)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE matches (
    id BINARY(16) PRIMARY KEY,
    user_id BINARY(16) NULL,
    version BIGINT UNSIGNED NOT NULL,
    status VARCHAR(40) NOT NULL,
    player_color VARCHAR(8) NOT NULL,
    side_to_move VARCHAR(8) NOT NULL,
    difficulty_level TINYINT UNSIGNED NOT NULL,
    engine_name VARCHAR(120) NOT NULL,
    difficulty_snapshot JSON NOT NULL,
    initial_fen VARCHAR(255) NOT NULL,
    current_fen VARCHAR(255) NOT NULL,
    current_hash BINARY(8) NOT NULL,
    outcome VARCHAR(24) NOT NULL,
    termination VARCHAR(40) NULL,
    allow_undo BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    finished_at TIMESTAMP(6) NULL,
    CONSTRAINT fk_matches_user FOREIGN KEY (user_id) REFERENCES users(id),
    KEY idx_matches_user_created (user_id, created_at),
    KEY idx_matches_status_updated (status, updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE match_moves (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    match_id BINARY(16) NOT NULL,
    ply INT UNSIGNED NOT NULL,
    match_version BIGINT UNSIGNED NOT NULL,
    move_iccs CHAR(4) NOT NULL,
    side VARCHAR(8) NOT NULL,
    actor VARCHAR(16) NOT NULL,
    captured_piece VARCHAR(40) NULL,
    fen_before VARCHAR(255) NOT NULL,
    fen_after VARCHAR(255) NOT NULL,
    hash_after BINARY(8) NOT NULL,
    think_time_ms INT UNSIGNED NULL,
    played_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    CONSTRAINT fk_match_moves_match FOREIGN KEY (match_id) REFERENCES matches(id) ON DELETE CASCADE,
    UNIQUE KEY uq_match_ply (match_id, ply),
    UNIQUE KEY uq_match_version (match_id, match_version),
    KEY idx_match_moves_hash (hash_after)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE match_events (
    id BINARY(16) PRIMARY KEY,
    match_id BINARY(16) NOT NULL,
    match_version BIGINT UNSIGNED NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    payload JSON NOT NULL,
    created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    CONSTRAINT fk_match_events_match FOREIGN KEY (match_id) REFERENCES matches(id) ON DELETE CASCADE,
    KEY idx_match_events_cursor (match_id, created_at, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE record_import_batches (
    id BINARY(16) PRIMARY KEY,
    user_id BINARY(16) NULL,
    status VARCHAR(24) NOT NULL,
    source_name VARCHAR(255) NOT NULL,
    source_hash BINARY(32) NULL,
    report JSON NULL,
    created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    completed_at TIMESTAMP(6) NULL,
    CONSTRAINT fk_import_user FOREIGN KEY (user_id) REFERENCES users(id),
    KEY idx_import_user_created (user_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE game_records (
    id BINARY(16) PRIMARY KEY,
    user_id BINARY(16) NULL,
    import_batch_id BINARY(16) NULL,
    name VARCHAR(255) NOT NULL,
    format VARCHAR(24) NOT NULL,
    source_hash BINARY(32) NOT NULL,
    content_hash BINARY(32) NOT NULL,
    initial_fen VARCHAR(255) NOT NULL,
    final_fen VARCHAR(255) NOT NULL,
    result VARCHAR(24) NULL,
    move_count INT UNSIGNED NOT NULL,
    metadata JSON NULL,
    created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    deleted_at TIMESTAMP(6) NULL,
    CONSTRAINT fk_records_user FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_records_import FOREIGN KEY (import_batch_id) REFERENCES record_import_batches(id),
    UNIQUE KEY uq_record_owner_content (user_id, content_hash),
    KEY idx_records_user_created (user_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE record_moves (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    record_id BINARY(16) NOT NULL,
    ply INT UNSIGNED NOT NULL,
    move_iccs CHAR(4) NOT NULL,
    side VARCHAR(8) NOT NULL,
    fen_before VARCHAR(255) NOT NULL,
    fen_after VARCHAR(255) NOT NULL,
    hash_after BINARY(8) NOT NULL,
    CONSTRAINT fk_record_moves_record FOREIGN KEY (record_id) REFERENCES game_records(id) ON DELETE CASCADE,
    UNIQUE KEY uq_record_ply (record_id, ply),
    KEY idx_record_moves_hash (hash_after)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE learning_jobs (
    id BINARY(16) PRIMARY KEY,
    user_id BINARY(16) NULL,
    status VARCHAR(24) NOT NULL,
    progress TINYINT UNSIGNED NOT NULL DEFAULT 0,
    input_spec JSON NOT NULL,
    checkpoint JSON NULL,
    error_code VARCHAR(64) NULL,
    error_message VARCHAR(500) NULL,
    created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    completed_at TIMESTAMP(6) NULL,
    CONSTRAINT fk_learning_job_user FOREIGN KEY (user_id) REFERENCES users(id),
    KEY idx_learning_jobs_status_created (status, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE learning_versions (
    id BINARY(16) PRIMARY KEY,
    job_id BINARY(16) NOT NULL,
    user_id BINARY(16) NULL,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(24) NOT NULL,
    algorithm_version VARCHAR(80) NOT NULL,
    parameters JSON NOT NULL,
    quality_report JSON NOT NULL,
    created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    activated_at TIMESTAMP(6) NULL,
    CONSTRAINT fk_learning_version_job FOREIGN KEY (job_id) REFERENCES learning_jobs(id),
    CONSTRAINT fk_learning_version_user FOREIGN KEY (user_id) REFERENCES users(id),
    KEY idx_learning_versions_user_status (user_id, status, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE opening_book_entries (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    learning_version_id BINARY(16) NOT NULL,
    position_hash BINARY(8) NOT NULL,
    fen_signature VARCHAR(255) NOT NULL,
    side_to_move VARCHAR(8) NOT NULL,
    move_iccs CHAR(4) NOT NULL,
    samples INT UNSIGNED NOT NULL,
    red_wins INT UNSIGNED NOT NULL DEFAULT 0,
    black_wins INT UNSIGNED NOT NULL DEFAULT 0,
    draws INT UNSIGNED NOT NULL DEFAULT 0,
    score DECIMAL(8,6) NULL,
    CONSTRAINT fk_book_version FOREIGN KEY (learning_version_id) REFERENCES learning_versions(id) ON DELETE CASCADE,
    UNIQUE KEY uq_book_version_position_move (learning_version_id, position_hash, move_iccs),
    KEY idx_book_lookup (learning_version_id, position_hash, side_to_move)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE analysis_jobs (
    id BINARY(16) PRIMARY KEY,
    match_id BINARY(16) NOT NULL,
    user_id BINARY(16) NULL,
    status VARCHAR(24) NOT NULL,
    progress TINYINT UNSIGNED NOT NULL DEFAULT 0,
    engine_snapshot JSON NOT NULL,
    error_code VARCHAR(64) NULL,
    error_message VARCHAR(500) NULL,
    created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    completed_at TIMESTAMP(6) NULL,
    CONSTRAINT fk_analysis_match FOREIGN KEY (match_id) REFERENCES matches(id) ON DELETE CASCADE,
    CONSTRAINT fk_analysis_user FOREIGN KEY (user_id) REFERENCES users(id),
    KEY idx_analysis_match_created (match_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE move_analyses (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    analysis_job_id BINARY(16) NOT NULL,
    ply INT UNSIGNED NOT NULL,
    actual_move CHAR(4) NOT NULL,
    best_move CHAR(4) NOT NULL,
    score_loss_cp INT NULL,
    classification VARCHAR(32) NOT NULL,
    depth SMALLINT UNSIGNED NOT NULL,
    nodes BIGINT UNSIGNED NOT NULL,
    candidates JSON NULL,
    CONSTRAINT fk_move_analysis_job FOREIGN KEY (analysis_job_id) REFERENCES analysis_jobs(id) ON DELETE CASCADE,
    UNIQUE KEY uq_analysis_ply (analysis_job_id, ply)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE idempotency_keys (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_scope VARCHAR(64) NOT NULL,
    route VARCHAR(120) NOT NULL,
    idempotency_key VARCHAR(160) NOT NULL,
    request_digest BINARY(32) NOT NULL,
    response_status SMALLINT UNSIGNED NOT NULL,
    response_body JSON NOT NULL,
    expires_at TIMESTAMP(6) NOT NULL,
    created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    UNIQUE KEY uq_idempotency_scope_route_key (user_scope, route, idempotency_key),
    KEY idx_idempotency_expiry (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE outbox_events (
    id BINARY(16) PRIMARY KEY,
    aggregate_type VARCHAR(40) NOT NULL,
    aggregate_id BINARY(16) NOT NULL,
    event_type VARCHAR(80) NOT NULL,
    payload JSON NOT NULL,
    available_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    published_at TIMESTAMP(6) NULL,
    attempts TINYINT UNSIGNED NOT NULL DEFAULT 0,
    KEY idx_outbox_pending (published_at, available_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

