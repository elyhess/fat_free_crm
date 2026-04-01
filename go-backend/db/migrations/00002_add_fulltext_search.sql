-- +goose Up

-- Add tsvector columns for full-text search
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS tsv tsvector;
ALTER TABLE contacts ADD COLUMN IF NOT EXISTS tsv tsvector;
ALTER TABLE leads ADD COLUMN IF NOT EXISTS tsv tsvector;
ALTER TABLE opportunities ADD COLUMN IF NOT EXISTS tsv tsvector;
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS tsv tsvector;

-- Create GIN indexes
CREATE INDEX IF NOT EXISTS idx_accounts_tsv ON accounts USING GIN (tsv);
CREATE INDEX IF NOT EXISTS idx_contacts_tsv ON contacts USING GIN (tsv);
CREATE INDEX IF NOT EXISTS idx_leads_tsv ON leads USING GIN (tsv);
CREATE INDEX IF NOT EXISTS idx_opportunities_tsv ON opportunities USING GIN (tsv);
CREATE INDEX IF NOT EXISTS idx_campaigns_tsv ON campaigns USING GIN (tsv);

-- Populate tsvector columns from existing data
UPDATE accounts SET tsv = to_tsvector('english', coalesce(name, '') || ' ' || coalesce(email, ''));
UPDATE contacts SET tsv = to_tsvector('english', coalesce(first_name, '') || ' ' || coalesce(last_name, '') || ' ' || coalesce(email, '') || ' ' || coalesce(phone, '') || ' ' || coalesce(mobile, ''));
UPDATE leads SET tsv = to_tsvector('english', coalesce(first_name, '') || ' ' || coalesce(last_name, '') || ' ' || coalesce(company, '') || ' ' || coalesce(email, ''));
UPDATE opportunities SET tsv = to_tsvector('english', coalesce(name, ''));
UPDATE campaigns SET tsv = to_tsvector('english', coalesce(name, ''));

-- Create triggers to auto-update tsvector on insert/update
CREATE OR REPLACE FUNCTION accounts_tsv_trigger() RETURNS trigger AS $$
BEGIN
  NEW.tsv := to_tsvector('english', coalesce(NEW.name, '') || ' ' || coalesce(NEW.email, ''));
  RETURN NEW;
END
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION contacts_tsv_trigger() RETURNS trigger AS $$
BEGIN
  NEW.tsv := to_tsvector('english', coalesce(NEW.first_name, '') || ' ' || coalesce(NEW.last_name, '') || ' ' || coalesce(NEW.email, '') || ' ' || coalesce(NEW.phone, '') || ' ' || coalesce(NEW.mobile, ''));
  RETURN NEW;
END
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION leads_tsv_trigger() RETURNS trigger AS $$
BEGIN
  NEW.tsv := to_tsvector('english', coalesce(NEW.first_name, '') || ' ' || coalesce(NEW.last_name, '') || ' ' || coalesce(NEW.company, '') || ' ' || coalesce(NEW.email, ''));
  RETURN NEW;
END
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION opportunities_tsv_trigger() RETURNS trigger AS $$
BEGIN
  NEW.tsv := to_tsvector('english', coalesce(NEW.name, ''));
  RETURN NEW;
END
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION campaigns_tsv_trigger() RETURNS trigger AS $$
BEGIN
  NEW.tsv := to_tsvector('english', coalesce(NEW.name, ''));
  RETURN NEW;
END
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS tsvector_update_accounts ON accounts;
CREATE TRIGGER tsvector_update_accounts BEFORE INSERT OR UPDATE ON accounts
  FOR EACH ROW EXECUTE FUNCTION accounts_tsv_trigger();

DROP TRIGGER IF EXISTS tsvector_update_contacts ON contacts;
CREATE TRIGGER tsvector_update_contacts BEFORE INSERT OR UPDATE ON contacts
  FOR EACH ROW EXECUTE FUNCTION contacts_tsv_trigger();

DROP TRIGGER IF EXISTS tsvector_update_leads ON leads;
CREATE TRIGGER tsvector_update_leads BEFORE INSERT OR UPDATE ON leads
  FOR EACH ROW EXECUTE FUNCTION leads_tsv_trigger();

DROP TRIGGER IF EXISTS tsvector_update_opportunities ON opportunities;
CREATE TRIGGER tsvector_update_opportunities BEFORE INSERT OR UPDATE ON opportunities
  FOR EACH ROW EXECUTE FUNCTION opportunities_tsv_trigger();

DROP TRIGGER IF EXISTS tsvector_update_campaigns ON campaigns;
CREATE TRIGGER tsvector_update_campaigns BEFORE INSERT OR UPDATE ON campaigns
  FOR EACH ROW EXECUTE FUNCTION campaigns_tsv_trigger();

-- +goose Down

DROP TRIGGER IF EXISTS tsvector_update_campaigns ON campaigns;
DROP TRIGGER IF EXISTS tsvector_update_opportunities ON opportunities;
DROP TRIGGER IF EXISTS tsvector_update_leads ON leads;
DROP TRIGGER IF EXISTS tsvector_update_contacts ON contacts;
DROP TRIGGER IF EXISTS tsvector_update_accounts ON accounts;

DROP FUNCTION IF EXISTS campaigns_tsv_trigger();
DROP FUNCTION IF EXISTS opportunities_tsv_trigger();
DROP FUNCTION IF EXISTS leads_tsv_trigger();
DROP FUNCTION IF EXISTS contacts_tsv_trigger();
DROP FUNCTION IF EXISTS accounts_tsv_trigger();

DROP INDEX IF EXISTS idx_campaigns_tsv;
DROP INDEX IF EXISTS idx_opportunities_tsv;
DROP INDEX IF EXISTS idx_leads_tsv;
DROP INDEX IF EXISTS idx_contacts_tsv;
DROP INDEX IF EXISTS idx_accounts_tsv;

ALTER TABLE campaigns DROP COLUMN IF EXISTS tsv;
ALTER TABLE opportunities DROP COLUMN IF EXISTS tsv;
ALTER TABLE leads DROP COLUMN IF EXISTS tsv;
ALTER TABLE contacts DROP COLUMN IF EXISTS tsv;
ALTER TABLE accounts DROP COLUMN IF EXISTS tsv;
