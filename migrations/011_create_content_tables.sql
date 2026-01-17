-- +goose Up
CREATE TABLE content_items (
    id         BIGSERIAL PRIMARY KEY,
    module_id  BIGINT NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    title      TEXT NOT NULL DEFAULT '',
    type       TEXT NOT NULL CHECK (type IN ('lecture')),
    status     TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published')),
    position   INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX content_items_module_id_idx ON content_items(module_id);
CREATE INDEX content_items_module_position_idx ON content_items(module_id, position);

CREATE TABLE lectures (
    content_item_id BIGINT PRIMARY KEY REFERENCES content_items(id) ON DELETE CASCADE,
    content         TEXT NOT NULL DEFAULT ''
);

-- +goose Down
DROP TABLE IF EXISTS lectures;
DROP TABLE IF EXISTS content_items;
