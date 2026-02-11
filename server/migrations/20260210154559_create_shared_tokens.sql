-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS shared_tokens (
  token TEXT NOT NULL PRIMARY KEY,
  collection_id UUID NOT NULL REFERENCES collections ON DELETE CASCADE,
  user_id UUID NOT NULL,
  org_id UUID NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS shared_tokens;
-- +goose StatementEnd
