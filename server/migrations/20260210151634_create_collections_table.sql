-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS collections(
  id UUID NOT NULL PRIMARY KEY,
  user_id UUID NOT NULL,
  org_id UUID NOT NULL,
  data TEXT,
  org_view BOOL DEFAULT FALSE,
  org_edit BOOL DEFAULT FALSE,
  org_share BOOL DEFAULT FALSE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS collections;
-- +goose StatementEnd
