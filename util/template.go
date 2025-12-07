package util

import (
	"html/template"
	"io"
	"io/fs"
)

type Template interface {
	Execute(wr io.Writer, data any)
}

type TemplateParser interface {
	ParseFiles(patterns ...string) Template
}

type HTMLTemplate struct {
	tmpl *template.Template
}

func (g *HTMLTemplate) Execute(wr io.Writer, data any) {
	g.tmpl.Execute(wr, data)
}

type HTMLTemplateParser struct {
	fs fs.FS
}

func NewHTMLTemplateParser(fs fs.FS) *HTMLTemplateParser {
	return &HTMLTemplateParser{fs}
}

func (p *HTMLTemplateParser) ParseFiles(patterns ...string) Template {
	templ := template.Must(template.ParseFS(p.fs, patterns...))
	return &HTMLTemplate{templ}
}
