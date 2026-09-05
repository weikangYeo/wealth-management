ALTER TABLE fund_info RENAME COLUMN url TO fsm_url;
ALTER TABLE fund_info ADD COLUMN scrape_param_value VARCHAR(255);
UPDATE fund_info SET scrape_param_value = fund_code;
