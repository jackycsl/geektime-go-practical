package web

import (
	"html/template"
	"log"
	"mime/multipart"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpload(t *testing.T) {
	tpl, err := template.ParseGlob("testdata/tpls/*.gohtml")
	require.NoError(t, err)
	engine := &GoTemplateEngine{
		T: tpl,
	}

	h := NewHTTPServer(ServerWithTemplateEngine(engine))
	h.Get("/upload", func(ctx *Context) {
		err := ctx.Render("upload.gohtml", nil)
		if err != nil {
			log.Println(err)
		}
	})

	fu := FileUploader{
		// <input type="file" name="myfile" />
		FileField: "myfile",
		DstPathFunc: func(header *multipart.FileHeader) string {
			return filepath.Join("testdata", "upload", header.Filename)
		},
	}
	h.Post("/upload", fu.Handle())
	h.Start(":8081")
}

func TestFileDownloader_Handle(t *testing.T) {
	s := NewHTTPServer()
	s.Get("/download", (&FileDownloader{
		// 下载的文件所在目录
		Dir: "./testdata/download",
	}).Handle())
	// 在浏览器里面输入 localhost:8081/download?file=test.txt
	s.Start(":8081")
}

func TestStaticResourceHandler_Handle(t *testing.T) {
	h := NewHTTPServer()

	s, err := NewStaticResourceHandler(filepath.Join("testdata", "static"))
	require.NoError(t, err)

	// /static/js/:file

	// localhost:8081/static/xxx.jpg
	h.Get("/static/:file", s.Handle)
	h.Start(":8081")
}
