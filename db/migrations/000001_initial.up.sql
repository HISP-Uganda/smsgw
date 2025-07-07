CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS user_roles
(
    id          BIGSERIAL NOT NULL PRIMARY KEY,
    name        TEXT      NOT NULL UNIQUE,
    description text        DEFAULT '',
    created     timestamptz DEFAULT CURRENT_TIMESTAMP,
    updated     timestamptz DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_role_permissions
(
    id         bigserial   NOT NULL PRIMARY KEY,
    user_role  BIGINT      NOT NULL REFERENCES user_roles ON DELETE CASCADE ON UPDATE CASCADE,
    sys_module TEXT        NOT NULL, -- the name of the module - defined above this level
    sys_perms  VARCHAR(16) NOT NULL,
    created    timestamptz DEFAULT CURRENT_TIMESTAMP,
    updated    timestamptz DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (sys_module, user_role)
);

CREATE TABLE IF NOT EXISTS users
(
    id                bigserial NOT NULL PRIMARY KEY,
    uid               TEXT      NOT NULL DEFAULT '',
    user_role         BIGINT    NOT NULL REFERENCES user_roles ON DELETE RESTRICT ON UPDATE CASCADE,
    username          TEXT      NOT NULL UNIQUE,
    password          TEXT      NOT NULL, -- blowfish hash of password
    onetime_password  TEXT,
    firstname         TEXT      NOT NULL,
    lastname          TEXT      NOT NULL,
    telephone         TEXT      NOT NULL DEFAULT '',
    email             TEXT,
    is_active         BOOLEAN   NOT NULL DEFAULT 't',
    is_admin_user    BOOLEAN   NOT NULL DEFAULT 'f',
    failed_attempts   TEXT               DEFAULT '0/' || to_char(NOW(), 'YYYYmmdd'),
    transaction_limit TEXT               DEFAULT '0/' || to_char(NOW(), 'YYYYmmdd'),
    last_login        timestamptz,
    created           timestamptz        DEFAULT CURRENT_TIMESTAMP,
    updated           timestamptz        DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX users_username_idx ON users (username);

CREATE TABLE IF NOT EXISTS user_apitoken
(
    id        bigserial NOT NULL PRIMARY KEY,
    user_id   BIGINT    NOT NULL REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE,
    token     TEXT      NOT NULL DEFAULT '',
    is_active BOOLEAN   NOT NULL DEFAULT TRUE,
    expires_at timestamptz DEFAULT NULL, -- if NULL, token does not expire
    created   timestamptz        DEFAULT CURRENT_TIMESTAMP,
    updated   timestamptz        DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE blacklist
(
    id      bigserial PRIMARY KEY,
    msisdn  text        NOT NULL,
    created timestamptz NOT NULL DEFAULT current_timestamp,
    updated timestamptz          DEFAULT current_timestamp
);

CREATE INDEX blacklist_msisdn ON blacklist (msisdn);
CREATE INDEX blacklist_created ON blacklist (created);
CREATE INDEX blacklist_updated ON blacklist (updated);

CREATE TABLE audit_log
(
    id         BIGSERIAL   NOT NULL PRIMARY KEY,
    logtype    VARCHAR(32) NOT NULL DEFAULT '',
    actor      TEXT        NOT NULL,
    action     text        NOT NULL DEFAULT '',
    remote_ip  INET,
    detail     TEXT        NOT NULL,
    created_by INTEGER REFERENCES users (id), -- like actor id
    created    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX audit_log_created ON audit_log (created);
CREATE INDEX audit_log_logtype ON audit_log (logtype);
CREATE INDEX audit_log_action ON audit_log (action);

INSERT INTO user_roles(name, description)
VALUES ('Administrator', 'For the Administrators'),
       ('SMS User', 'For SMS third party apps');

INSERT INTO user_role_permissions(user_role, sys_module, sys_perms)
VALUES ((SELECT id FROM user_roles WHERE name = 'Administrator'), 'Users', 'rmad');

INSERT INTO users(firstname, lastname, username, password, email, user_role, is_admin_user)
VALUES ('Samuel', 'Sekiwere', 'admin', crypt('@dm1n', gen_salt('bf')), 'sekiskylink@gmail.com',
        (SELECT id FROM user_roles WHERE name = 'Administrator'), 't');