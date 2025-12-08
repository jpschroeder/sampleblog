package posts

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/jpschroeder/sampleblog/queries"
	"github.com/jpschroeder/sampleblog/util"
)

type Controller struct {
	querier        queries.Querier
	index_template util.Template
	show_template  util.Template
	new_template   util.Template
	edit_template  util.Template
}

func NewController(querier queries.Querier, templates util.TemplateParser) *Controller {
	return &Controller{
		querier:        querier,
		index_template: templates.ParseFiles("posts/post_index.html"),
		show_template:  templates.ParseFiles("posts/post_show.html"),
		new_template:   templates.ParseFiles("posts/post_new.html"),
		edit_template:  templates.ParseFiles("posts/post_edit.html"),
	}
}

func (c *Controller) Index(w http.ResponseWriter, r *http.Request) {
	posts, err := c.querier.ListPosts(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch posts: "+err.Error(), http.StatusInternalServerError)
		return
	}
	c.index_template.Execute(w, posts)
}

func (c *Controller) Show(w http.ResponseWriter, r *http.Request, id int64) {
	post, err := c.querier.GetPost(r.Context(), id)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, "Failed to fetch post", http.StatusInternalServerError)
		return
	}

	c.show_template.Execute(w, post)
}

func (c *Controller) New(w http.ResponseWriter, r *http.Request) {
	c.new_template.Execute(w, nil)
}

func (c *Controller) Edit(w http.ResponseWriter, r *http.Request, id int64) {
	post, err := c.querier.GetPost(r.Context(), id)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, "Failed to fetch post", http.StatusInternalServerError)
		return
	}

	c.edit_template.Execute(w, post)
}

func (c *Controller) Create(w http.ResponseWriter, r *http.Request) {
	post, err := c.querier.CreatePost(r.Context(), queries.CreatePostParams{
		Title:   r.FormValue("title"),
		Content: r.FormValue("content"),
	})
	if err != nil {
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/posts/%d", post.ID), http.StatusSeeOther)
}

func (c *Controller) Update(w http.ResponseWriter, r *http.Request, id int64) {
	err := c.querier.UpdatePost(r.Context(), queries.UpdatePostParams{
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

func (c *Controller) Destroy(w http.ResponseWriter, r *http.Request, id int64) {
	err := c.querier.DeletePost(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to delete post", http.StatusInternalServerError)
		return
	}

	// Redirect to the list of posts after deletion
	http.Redirect(w, r, "/posts", http.StatusSeeOther)
}
