ALTER TABLE "product"
ADD COLUMN "search" tsvector
GENERATED ALWAYS AS (
setweight(to_tsvector('english', COALESCE(name, '')), 'A') ||
setweight(to_tsvector('english', COALESCE(description, '')), 'B')
) STORED;

CREATE INDEX search_idx
ON "product" USING GIN ("search");