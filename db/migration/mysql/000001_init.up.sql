CREATE TABLE IF NOT EXISTS projects (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name       VARCHAR(255)    NOT NULL,
    owner_id   BIGINT UNSIGNED NULL,
    created_at DATETIME(3)     NOT NULL,
    updated_at DATETIME(3)     NOT NULL,
    PRIMARY KEY (id),
    KEY idx_projects_owner (owner_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS services (
    id             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    project_id     BIGINT UNSIGNED NOT NULL,
    name           VARCHAR(255)    NOT NULL,
    public_id      VARCHAR(64)     NOT NULL,
    api_key_hash   CHAR(64)        NOT NULL,
    api_key_prefix VARCHAR(32)     NOT NULL,
    created_at     DATETIME(3)     NOT NULL,
    updated_at     DATETIME(3)     NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_services_public_id (public_id),
    UNIQUE KEY uq_services_api_key_hash (api_key_hash),
    KEY idx_services_project (project_id),
    CONSTRAINT fk_services_project FOREIGN KEY (project_id)
        REFERENCES projects (id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS issues (
    id                         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    service_id                 BIGINT UNSIGNED NOT NULL,
    fingerprint                CHAR(64)        NOT NULL,
    title                      VARCHAR(512)    NOT NULL DEFAULT '',
    culprit                    VARCHAR(512)    NOT NULL DEFAULT '',
    level                      VARCHAR(16)     NOT NULL DEFAULT 'error',
    status                     VARCHAR(16)     NOT NULL DEFAULT 'unresolved',
    first_seen                 DATETIME(3)     NOT NULL,
    last_seen                  DATETIME(3)     NOT NULL,
    event_count                BIGINT UNSIGNED NOT NULL DEFAULT 0,
    affected_users_estimate    BIGINT UNSIGNED NOT NULL DEFAULT 0,
    affected_sessions_estimate BIGINT UNSIGNED NOT NULL DEFAULT 0,
    regressed_at               DATETIME(3)     NULL,
    metadata                   JSON            NULL,
    created_at                 DATETIME(3)     NOT NULL,
    updated_at                 DATETIME(3)     NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_issues_service_fingerprint (service_id, fingerprint),
    KEY idx_issues_service_status_lastseen (service_id, status, last_seen),
    CONSTRAINT fk_issues_service FOREIGN KEY (service_id)
        REFERENCES services (id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
