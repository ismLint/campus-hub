BEGIN;

CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    telegram_id BIGINT UNIQUE,
    username TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE subjects (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT subjects_name_not_empty
        CHECK (length(trim(name)) > 0),

    CONSTRAINT subjects_user_name_unique
        UNIQUE (user_id, name),

    CONSTRAINT subjects_id_user_unique
        UNIQUE (id, user_id)
);

CREATE TABLE tasks (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subject_id BIGINT,
    title TEXT NOT NULL,
    subject TEXT,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'todo', --todo, in_progress, done
    deadline TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT tasks_title_not_empty
        CHECK (length(trim(title)) > 0),

    CONSTRAINT tasks_status_valid
        CHECK (status IN ('todo', 'in_progress', 'done')),

    CONSTRAINT tasks_subject_owner_fk
        FOREIGN KEY (subject_id, user_id)
        REFERENCES subjects(id, user_id)
);

CREATE INDEX idx_tasks_user_status
    ON tasks(user_id, status);

CREATE INDEX idx_tasks_user_deadline
    ON tasks(user_id, deadline)
    WHERE deadline IS NOT NULL;

COMMIT;