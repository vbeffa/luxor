
--- change / add / modify this schema as necessary
CREATE TABLE public.sample_table (
    id bigint NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    sample VARCHAR(255)
);

CREATE TABLE auth_requests (
  id int NOT NULL,
  username VARCHAR(255),
  pass_hash VARCHAR(255),
  req_ts TIMESTAMP
);
