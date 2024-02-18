-- +goose Up
CREATE TABLE IF NOT EXISTS notes (
    id SERIAL PRIMARY KEY,
    title varchar(30),
    description varchar(255), 
    date_added timestamptz,
    date_notify timestamptz,
    delay interval
);

-- +goose Down
DROP TABLE IF EXISTS notes;