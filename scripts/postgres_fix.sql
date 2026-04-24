-- Gogal PostgreSQL 15/16 public schema privilege fix
-- Replace placeholders before running:
--   :db_name -> target database (example: gogaldb)
--   :db_user -> target role/user (example: gogaluser)

-- Example:
-- \set db_name gogaldb
-- \set db_user gogaluser

GRANT CONNECT ON DATABASE :db_name TO :db_user;
GRANT ALL PRIVILEGES ON DATABASE :db_name TO :db_user;
GRANT USAGE, CREATE ON SCHEMA public TO :db_user;
ALTER SCHEMA public OWNER TO :db_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO :db_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO :db_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO :db_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO :db_user;
