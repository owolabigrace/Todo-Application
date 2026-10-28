CREATE TABLE categories (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name       VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX categories_name_lower_idx ON categories (lower(name));

ALTER TABLE todos
    ADD COLUMN category_id BIGINT REFERENCES categories(id) ON DELETE SET NULL;

CREATE INDEX todos_category_id_idx ON todos (category_id);
