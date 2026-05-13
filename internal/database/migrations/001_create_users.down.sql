DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at();
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS user_role;
DROP TYPE IF EXISTS kyc_status;
