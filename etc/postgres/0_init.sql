CREATE TABLE auth_requests (
  id bigint NOT NULL,
  username VARCHAR(255),
  pass_hash VARCHAR(255),
  created_at timestamp WITHOUT time zone DEFAULT now() NOT NULL
);

CREATE TABLE subscriptions (
  id bigint NOT NULL,
  subscription_id_1 VARCHAR(32),
  subscription_id_2 VARCHAR(32),
  extra_nonce_1 VARCHAR(32)
);
