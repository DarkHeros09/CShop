ALTER TABLE product
ADD COLUMN "search" tsvector
GENERATED ALWAYS AS (
setweight(to_tsvector('english', name), 'A') ||
setweight(to_tsvector('english', description), 'B')
) STORED;

CREATE INDEX search_idx
ON product USING GIN ("search");