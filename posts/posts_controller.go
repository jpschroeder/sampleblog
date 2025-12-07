package posts

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/jpschroeder/sampleblog/queries"
	"github.com/jpschroeder/sampleblog/util"
)

type PostsController struct {
	queries        *queries.Queries
	index_template util.Template
	show_template  util.Template
	new_template   util.Template
	edit_template  util.Template
}

func NewPostsController(queries *queries.Queries, templates util.TemplateParser) *PostsController {
	return &PostsController{
		queries:        queries,
		index_template: templates.ParseFiles("posts/post_index.html"),
		show_template:  templates.ParseFiles("posts/post_show.html"),
		new_template:   templates.ParseFiles("posts/post_new.html"),
		edit_template:  templates.ParseFiles("posts/post_edit.html"),
	}
}

func (c *PostsController) Index(w http.ResponseWriter, r *http.Request) {
	posts, err := c.queries.ListPosts(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch posts: "+err.Error(), http.StatusInternalServerError)
		return
	}
	c.index_template.Execute(w, posts)
}

func (c *PostsController) Show(w http.ResponseWriter, r *http.Request, id int64) {
	post, err := c.queries.GetPost(r.Context(), id)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, "Failed to fetch post", http.StatusInternalServerError)
		return
	}

	c.show_template.Execute(w, post)
}

func (c *PostsController) New(w http.ResponseWriter, r *http.Request) {
	c.new_template.Execute(w, nil)
}

func (c *PostsController) Edit(w http.ResponseWriter, r *http.Request, id int64) {
	post, err := c.queries.GetPost(r.Context(), id)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, "Failed to fetch post", http.StatusInternalServerError)
		return
	}

	c.edit_template.Execute(w, post)
}

func (c *PostsController) Create(w http.ResponseWriter, r *http.Request) {
	post, err := c.queries.CreatePost(r.Context(), queries.CreatePostParams{
		Title:   r.FormValue("title"),
		Content: r.FormValue("content"),
	})
	if err != nil {
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/posts/%d", post.ID), http.StatusSeeOther)
}

func (c *PostsController) Update(w http.ResponseWriter, r *http.Request, id int64) {
	err := c.queries.UpdatePost(r.Context(), queries.UpdatePostParams{
		Title:   r.FormValue("title"),
		Content: r.FormValue("content"),
		ID:      id,
	})
	if err != nil {
		http.Error(w, "Failed to update post", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/posts/%d", id), http.StatusSeeOther)
}

func (c *PostsController) Destroy(w http.ResponseWriter, r *http.Request, id int64) {
	err := c.queries.DeletePost(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to delete post", http.StatusInternalServerError)
		return
	}

	// Redirect to the list of posts after deletion
	http.Redirect(w, r, "/posts", http.StatusSeeOther)
}
