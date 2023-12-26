CREATE INDEX IF NOT EXISTS accounts_owner_idx ON "accounts" ("owner");

CREATE UNIQUE INDEX IF NOT EXISTS accounts_owner_currency_idx ON "accounts" ("owner", "currency");

CREATE INDEX IF NOT EXISTS entries_account_id_idx ON "entries" ("account_id");

CREATE INDEX IF NOT EXISTS transfers_from_account_id_idx ON "transfers" ("from_account_id");

CREATE INDEX IF NOT EXISTS transfers_to_account_id_idx ON "transfers" ("to_account_id");

CREATE INDEX IF NOT EXISTS transfers_from_to_account_idx ON "transfers" ("from_account_id", "to_account_id");
