-- name: ListPosts :many
SELECT id, title, content FROM posts;

-- name: GetPost :one
SELECT id, title, content FROM posts WHERE id = ?;

-- name: CreatePost :one
INSERT INTO posts (title, content) VALUES (?, ?) RETURNING *;

-- name: UpdatePost :exec
UPDATE posts SET title = ?, content = ? WHERE id = ? RETURNING *;

-- name: DeletePost :exec
DELETE FROM posts WHERE id = ?;
