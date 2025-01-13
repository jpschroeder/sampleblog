package main

import (
	"database/sql"
	"errors"
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
	const query = "SELECT id, title, content FROM posts"
	const view = `<!DOCTYPE html>
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
</html>`

	var (
		once    sync.Once
		initerr error
		stmt    *sql.Stmt
		tmpl    *template.Template
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Initialization: Prepare database query and parse html template only once
		once.Do(func() {
			var stmterr, tmplerr error
			stmt, stmterr = getDB().PrepareContext(ctx, query)
			tmpl, tmplerr = getTemplate().New("index").Parse(view)
			initerr = errors.Join(stmterr, tmplerr)
		})
		if initerr != nil {
			http.Error(w, "Failed to initialize handler: "+initerr.Error(), http.StatusInternalServerError)
			return
		}

		// Run database query
		rows, err := stmt.QueryContext(ctx)
		if err != nil {
			http.Error(w, "Failed to fetch posts: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Parse database results
		var posts []Post
		for rows.Next() {
			var post Post
			if err := rows.Scan(&post.ID, &post.Title, &post.Content); err != nil {
				http.Error(w, "Failed to parse posts", http.StatusInternalServerError)
				return
			}
			posts = append(posts, post)
		}

		// Render the html view template
		tmpl.Execute(w, posts)
	})
}

func postsShowHandler() http.Handler {
	const query = "SELECT id, title, content FROM posts WHERE id = ?"
	const view = `<!DOCTYPE html>
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
</html>`

	var (
		once    sync.Once
		initerr error
		stmt    *sql.Stmt
		tmpl    *template.Template
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Initialization: Prepare database query and parse html template only once
		once.Do(func() {
			var stmterr, tmplerr error
			stmt, stmterr = getDB().PrepareContext(ctx, query)
			tmpl, tmplerr = getTemplate().New("show").Parse(view)
			initerr = errors.Join(stmterr, tmplerr)
		})
		if initerr != nil {
			http.Error(w, "Failed to initialize handler: "+initerr.Error(), http.StatusInternalServerError)
			return
		}

		// Get the id from the path value
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		// Run database query
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

		// Render the html view template
		tmpl.Execute(w, post)
	})
}

func postsNewHandler() http.Handler {
	const view = `<!DOCTYPE html>
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
</html>`

	var (
		once    sync.Once
		initerr error
		tmpl    *template.Template
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		once.Do(func() {
			tmpl, initerr = getTemplate().New("new").Parse(view)
		})
		if initerr != nil {
			http.Error(w, "Failed to initialize handler: "+initerr.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.Execute(w, nil)
	})
}

func postsEditHandler() http.Handler {
	const query = "SELECT id, title, content FROM posts WHERE id = ?"
	const view = `<!DOCTYPE html>
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
</html>`

	var (
		once    sync.Once
		initerr error
		stmt    *sql.Stmt
		tmpl    *template.Template
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		once.Do(func() {
			var stmterr, tmplerr error
			stmt, stmterr = getDB().PrepareContext(ctx, query)
			tmpl, tmplerr = getTemplate().New("edit").Parse(view)
			initerr = errors.Join(stmterr, tmplerr)
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
	const query = "INSERT INTO posts (title, content) VALUES (?, ?)"
	var (
		once    sync.Once
		initerr error
		stmt    *sql.Stmt
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		once.Do(func() {
			stmt, initerr = getDB().PrepareContext(ctx, query)
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
	const query = "UPDATE posts SET title = ?, content = ? WHERE id = ?"
	var (
		once    sync.Once
		initerr error
		stmt    *sql.Stmt
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		once.Do(func() {
			stmt, initerr = getDB().PrepareContext(ctx, query)
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
	const query = "DELETE FROM posts WHERE id = ?"
	var (
		once    sync.Once
		initerr error
		stmt    *sql.Stmt
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		once.Do(func() {
			stmt, initerr = getDB().PrepareContext(ctx, query)
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
