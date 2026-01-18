-- name: CreateFile :one
INSERT INTO files (
  url,
  name,
  content_type,
  size,
  checksum,
  uploader
) VALUES (
  sqlc.arg(url),
  sqlc.arg(name),
  sqlc.arg(content_type),
  sqlc.arg(size),
  sqlc.arg(checksum),
  sqlc.arg(uploader)
) RETURNING
  id,
  url,
  name,
  content_type,
  size,
  checksum,
  uploader,
  status,
  created_at;

-- name: GetFileByURL :one
SELECT
  id,
  url,
  name,
  content_type,
  size,
  checksum,
  uploader,
  status,
  created_at
FROM files
WHERE url = sqlc.arg(url)
  AND status = 'NORMAL'
LIMIT 1;
