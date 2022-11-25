package web

import (
	"bytes"
	"context"
	"text/template"
)

type TemplateEngine interface {
	// Render 渲染页面
	// data 是渲染页面所需要的数据
	Render(ctx context.Context, tplName string, data any) ([]byte, error)
}

type GoTemplateEngine struct {
	T *template.Template
}

func (g *GoTemplateEngine) Render(ctx context.Context, tplName string, data any) ([]byte, error) {
	bs := &bytes.Buffer{}
	err := g.T.ExecuteTemplate(bs, tplName, data)
	return bs.Bytes(), err
}
