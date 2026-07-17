CREATE TABLE lists (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug       text NOT NULL UNIQUE,
    name       text NOT NULL,
    color      text NOT NULL DEFAULT 'zinc',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE tasks (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    list_id      uuid NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
    title        text NOT NULL,
    done         boolean NOT NULL DEFAULT false,
    completed_at timestamptz,
    position     integer NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX tasks_list_id_position_idx ON tasks (list_id, position);
