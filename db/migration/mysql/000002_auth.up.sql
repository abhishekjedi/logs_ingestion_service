CREATE TABLE IF NOT EXISTS users (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    email      VARCHAR(255)    NOT NULL,
    name       VARCHAR(255)    NOT NULL DEFAULT '',
    avatar_url VARCHAR(1024)   NOT NULL DEFAULT '',
    google_sub VARCHAR(255)    NULL,
    created_at DATETIME(3)     NOT NULL,
    updated_at DATETIME(3)     NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_users_email (email),
    UNIQUE KEY uq_users_google_sub (google_sub)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS organizations (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name       VARCHAR(255)    NOT NULL,
    slug       VARCHAR(255)    NOT NULL,
    created_by BIGINT UNSIGNED NOT NULL,
    created_at DATETIME(3)     NOT NULL,
    updated_at DATETIME(3)     NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_org_slug (slug),
    KEY idx_org_creator (created_by),
    CONSTRAINT fk_org_creator FOREIGN KEY (created_by) REFERENCES users (id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS organization_members (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    org_id     BIGINT UNSIGNED NOT NULL,
    user_id    BIGINT UNSIGNED NULL,
    email      VARCHAR(255)    NOT NULL,
    role       VARCHAR(16)     NOT NULL DEFAULT 'member',
    status     VARCHAR(16)     NOT NULL DEFAULT 'active',
    created_at DATETIME(3)     NOT NULL,
    updated_at DATETIME(3)     NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_member_org_email (org_id, email),
    KEY idx_member_user (user_id),
    CONSTRAINT fk_member_org FOREIGN KEY (org_id)
        REFERENCES organizations (id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

ALTER TABLE projects
    ADD COLUMN org_id BIGINT UNSIGNED NULL AFTER id,
    ADD KEY idx_projects_org (org_id);
