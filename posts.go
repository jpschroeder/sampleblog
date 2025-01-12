package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"sync"
)

// Post represents a blog post type
type Post struct {
	ID      int
	Title   string
	Content string
}

func postsIndexHandler() http.Handler {
	var once sync.Once
	var initerr error
	var tmpl *template.Template
	var stmt *sql.Stmt

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		once.Do(func() {
			stmt, initerr = getDB().PrepareContext(ctx, `
SELECT id, title, content FROM posts`)
			if initerr != nil {
				return
			}
			tmpl, initerr = getTemplate().New("index").Parse(`<!DOCTYPE html>
<html>
<head>
	<title>Blog</title>
</head>
<body>
	<h1>Posts</h1>
	{{range .}}
	<p><a href="/posts/{{.ID}}">{{.Title}}</a></p>
	{{end}}
	<a href="/posts/new">New Post</a>
</body>
</html>`)
		})
		if initerr != nil {
			http.Error(w, "Failed to initialize handler: "+initerr.Error(), http.StatusInternalServerError)
			return
		}

		rows, err := stmt.QueryContext(ctx)
		if err != nil {
			http.Error(w, "Failed to fetch posts: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var posts []Post
		for rows.Next() {
			var post Post
			if err := rows.Scan(&post.ID, &post.Title, &post.Content); err != nil {
				http.Error(w, "Failed to parse posts", http.StatusInternalServerError)
				return
			}
			posts = append(posts, post)
		}
		tmpl.Execute(w, posts)
	})
}

func postsShowHandler() http.Handler {
	var once sync.Once
	var initerr error
	var tmpl *template.Template
	var stmt *sql.Stmt

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		once.Do(func() {
			stmt, initerr = getDB().PrepareContext(ctx, `
SELECT id, title, content FROM posts WHERE id = ?`)
			if initerr != nil {
				return
			}
			tmpl, initerr = template.New("show").Parse(`<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
</head>
<body>
    <h1>{{.Title}}</h1>
    <p>{{.Content}}</p>
    <p><a href="/posts/{{.ID}}/edit">Edit this Post</a></p>
    <p><a href="/posts">Back to Posts</a></p>
		<form action="/posts/{{.ID}}/destroy" method="post">
			<button type="submit">Destroy this Post</button>
		</form>
</body>
</html>`)
		})
		if initerr != nil {
			http.Error(w, "Failed to initialize handler: "+initerr.Error(), http.StatusInternalServerError)
			return
		}

		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		row := stmt.QueryRowContext(ctx, id)
		var post Post
		if err := row.Scan(&post.ID, &post.Title, &post.Content); err != nil {
			if err == sql.ErrNoRows {
				http.NotFound(w, r)
			} else {
				http.Error(w, "Failed to fetch post", http.StatusInternalServerError)
			}
			return
		}
		tmpl.Execute(w, post)
	})
}

func postsNewHandler() http.Handler {
	var once sync.Once
	var initerr error
	var tmpl *template.Template

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		once.Do(func() {
			tmpl, initerr = getTemplate().New("new").Parse(`<!DOCTYPE html>
<html>
<head>
    <title>New Post</title>
</head>
<body>
    <h1>Create a New Post</h1>
    <form action="/posts" method="post">
        <label for="title">Title:</label><br>
        <input type="text" id="title" name="title"><br>
        <label for="content">Content:</label><br>
        <textarea id="content" name="content"></textarea><br>
        <button type="submit">Create Post</button>
    </form>
    <a href="/posts">Back to Posts</a>
</body>
</html>`)
		})
		if initerr != nil {
			http.Error(w, "Failed to initialize handler: "+initerr.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.Execute(w, nil)
	})
}

func postsEditHandler() http.Handler {
	var once sync.Once
	var initerr error
	var tmpl *template.Template
	var stmt *sql.Stmt

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		once.Do(func() {
			stmt, initerr = getDB().PrepareContext(ctx, `
SELECT id, title, content FROM posts WHERE id = ?`)
			if initerr != nil {
				return
			}
			tmpl, initerr = getTemplate().New("edit").Parse(`<!DOCTYPE html>
<html>
<head>
    <title>Edit Post</title>
</head>
<body>
    <h1>Edit Post</h1>
    <form action="/posts/{{.ID}}" method="post">
        <label for="title">Title:</label><br>
        <input type="text" id="title" name="title" value="{{.Title}}"><br>
        <label for="content">Content:</label><br>
        <textarea id="content" name="content">{{.Content}}</textarea><br>
        <button type="submit">Update Post</button>
    </form>
    <p><a href="/posts/{{.ID}}">Show this Post</a></p>
    <p><a href="/posts">Back to Posts</a></p>
</body>
</html>`)
		})
		if initerr != nil {
			http.Error(w, "Failed to initialize handler: "+initerr.Error(), http.StatusInternalServerError)
			return
		}

		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		row := stmt.QueryRowContext(ctx, id)
		var post Post
		if err := row.Scan(&post.ID, &post.Title, &post.Content); err != nil {
			if err == sql.ErrNoRows {
				http.NotFound(w, r)
			} else {
				http.Error(w, "Failed to fetch post", http.StatusInternalServerError)
			}
			return
		}
		tmpl.Execute(w, post)
	})
}

func postsCreateHandler() http.Handler {
	var once sync.Once
	var stmt *sql.Stmt
	var initerr error

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		once.Do(func() {
			stmt, initerr = getDB().PrepareContext(ctx, `
INSERT INTO posts (title, content) VALUES (?, ?)`)
		})
		if initerr != nil {
			http.Error(w, "Failed to initialize handler: "+initerr.Error(), http.StatusInternalServerError)
			return
		}

		title := r.FormValue("title")
		content := r.FormValue("content")
		res, err := stmt.ExecContext(ctx, title, content)
		if err != nil {
			http.Error(w, "Failed to create post", http.StatusInternalServerError)
			return
		}

		id, err := res.LastInsertId()
		if err != nil {
			http.Error(w, "Failed to retrieve new post ID", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/posts/%d", id), http.StatusSeeOther)
	})
}

func postsUpdateHandler() http.Handler {
	var once sync.Once
	var stmt *sql.Stmt
	var initerr error

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		once.Do(func() {
			stmt, initerr = getDB().PrepareContext(ctx, `
UPDATE posts SET title = ?, content = ? WHERE id = ?`)
		})
		if initerr != nil {
			http.Error(w, "Failed to initialize handler: "+initerr.Error(), http.StatusInternalServerError)
			return
		}

		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		title := r.FormValue("title")
		content := r.FormValue("content")
		_, err = stmt.ExecContext(ctx, title, content, id)
		if err != nil {
			http.Error(w, "Failed to update post", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/posts/%d", id), http.StatusSeeOther)
	})
}

func postsDestroyHandler() http.Handler {
	var once sync.Once
	var stmt *sql.Stmt
	var initerr error

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		once.Do(func() {
			stmt, initerr = getDB().PrepareContext(ctx, `
DELETE FROM posts WHERE id = ?`)
		})
		if initerr != nil {
			http.Error(w, "Failed to initialize handler: "+initerr.Error(), http.StatusInternalServerError)
			return
		}

		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		_, err = stmt.ExecContext(ctx, id)
		if err != nil {
			http.Error(w, "Failed to delete post", http.StatusInternalServerError)
			return
		}

		// Redirect to the list of posts after deletion
		http.Redirect(w, r, "/posts", http.StatusSeeOther)
	})
}
