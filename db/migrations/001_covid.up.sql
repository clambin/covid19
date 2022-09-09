CREATE TABLE IF NOT EXISTS covid19 (
   time TIMESTAMP WITHOUT TIME ZONE,
   country_code TEXT,
   country_name TEXT,
   confirmed BIGINT,
   death BIGINT,
   recovered BIGINT
);
CREATE INDEX IF NOT EXISTS idx_covid_country_name ON covid19(country_name);
CREATE INDEX IF NOT EXISTS idx_covid_country_code ON covid19(country_code);
CREATE INDEX IF NOT EXISTS idx_covid_time ON covid19(time);
