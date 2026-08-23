-- Shared trigger function: every table with an updated_at column attaches this
-- so the application never has to set the timestamp by hand.
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
