-- The sector breakdown of an asset that has more than one.
--
-- 000034 gave every asset a single sector, which is the right shape for a share
-- and for a sector fund — XLK is technology and nothing else — and the wrong
-- shape for the instrument most people actually hold. A whole-market ETF like
-- VOO is every sector at once, in known proportions the fund publishes, so the
-- single column left it with two bad answers: pick one industry and lie about
-- the other ten, or leave it NULL and have the breakdown report the biggest
-- position in the portfolio as "nobody has classified this yet" — work to do
-- that nobody can ever do.
--
-- So the column stays and this table sits beside it, rather than replacing it.
-- The two are exclusive by construction and the application is what holds that
-- line (market.AssetSpec/AssetUpdate, validated in asset_service.go): an asset
-- either carries one sector in assets.sector or N rows here, never both. A
-- single sector is the overwhelmingly common case and a table row per share
-- would have cost every read a join to learn what one column already says.
--
-- weight is a percentage, not a fraction: it is transcribed from a fund fact
-- sheet, where it reads "Information Technology 33.1%", and a column that made
-- the operator divide by a hundred first would be a column entered wrong. Four
-- decimals is far past what any provider publishes and leaves room for a
-- breakdown that was computed rather than copied.
--
-- The weights are NOT required to add up to 100. A fact sheet's do not — a
-- couple of tenths sit in cash and futures — and the ones somebody transcribes
-- by hand are worse. The allocation normalises over the total it finds instead
-- (see GetSectorAllocationByUserID), so a breakdown that covers 97.3 % still
-- accounts for 100 % of the position and the portfolio total stays whole. What is
-- rejected, in Go, is a total above 100: that is not an incomplete transcription
-- but a wrong one.
CREATE TABLE IF NOT EXISTS asset_sector_weights (
  asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
  sector   VARCHAR(32) NOT NULL,
  weight   NUMERIC(7, 4) NOT NULL CHECK (weight > 0 AND weight <= 100),
  -- One row per (asset, sector): a breakdown that named technology twice would
  -- be a transcription error, and the primary key says so instead of silently
  -- adding the two. Its index is also the only one this table needs — every
  -- read here starts from an asset_id, and the pair's leading column serves
  -- them.
  PRIMARY KEY (asset_id, sector)
);
