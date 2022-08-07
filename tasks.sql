BEGIN;

CREATE TABLE task (
  id          serial PRIMARY KEY,
  description text,
  tags        text[],
  due         timestamptz,
  created     timestamptz NOT NULL
);

COMMIT;
