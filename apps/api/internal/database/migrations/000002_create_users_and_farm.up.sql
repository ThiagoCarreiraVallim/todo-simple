CREATE TABLE users (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    username   text NOT NULL UNIQUE,   -- login sem senha; guardado em minúsculo
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE farms (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,  -- 1 fazenda por usuário
    name       text NOT NULL DEFAULT '',
    coins      integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE animals (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    farm_id      uuid NOT NULL REFERENCES farms(id) ON DELETE CASCADE,
    type         text NOT NULL,
    comfort_base real NOT NULL DEFAULT 50,
    last_fed_at  timestamptz NOT NULL DEFAULT now(),
    last_egg_at  timestamptz NOT NULL DEFAULT now(),
    created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX animals_farm_id_idx ON animals (farm_id);

CREATE TABLE crops (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    farm_id    uuid NOT NULL REFERENCES farms(id) ON DELETE CASCADE,
    plot_index integer NOT NULL,
    type       text NOT NULL,
    planted_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (farm_id, plot_index)
);

CREATE TABLE farm_items (
    farm_id uuid NOT NULL REFERENCES farms(id) ON DELETE CASCADE,
    item    text NOT NULL,
    qty     integer NOT NULL DEFAULT 0,
    PRIMARY KEY (farm_id, item)
);
