CREATE TABLE IF NOT EXISTS replay_integrations (
    id                   BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    project_id           BIGINT UNSIGNED NOT NULL,
    provider             VARCHAR(32)     NOT NULL,
    external_project_key VARCHAR(255)    NOT NULL,
    api_base_url         VARCHAR(1024)   NOT NULL,
    ingest_point         VARCHAR(1024)   NOT NULL,
    api_key_ciphertext   TEXT            NOT NULL,
    enabled              BOOLEAN         NOT NULL DEFAULT TRUE,
    last_validated_at    DATETIME(3)     NULL,
    last_error           VARCHAR(1024)   NOT NULL DEFAULT '',
    created_at           DATETIME(3)     NOT NULL,
    updated_at           DATETIME(3)     NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_replay_project_provider_key (project_id, provider, external_project_key),
    KEY idx_replay_integrations_project (project_id),
    CONSTRAINT fk_replay_integrations_project FOREIGN KEY (project_id)
        REFERENCES projects (id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
