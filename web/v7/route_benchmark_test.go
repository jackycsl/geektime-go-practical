package web

import (
	"net/http"
	"testing"
)

func BenchmarkFindRoute_Static(b *testing.B) {
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
	}
	mockHandler := func(ctx *Context) {}

	testCases := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "root",
			method: http.MethodGet,
			path:   "/",
		},
		{
			name:   "user",
			method: http.MethodGet,
			path:   "/user",
		},
	}

	r := newRouter()
	for _, tr := range testRoutes {
		r.addRoute(tr.method, tr.path, mockHandler)
	}

	for n := 0; n < b.N; n++ {
		for _, tc := range testCases {
			r.findRoute(tc.method, tc.path)
		}
	}
}

func BenchmarkFindRoute_Wildcard(b *testing.B) {
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodPost,
			path:   "/order/*",
		},
	}
	mockHandler := func(ctx *Context) {}

	testCases := []struct {
		name   string
		method string
		path   string
	}{
		{
			// 命中/order/*
			name:   "star match",
			method: http.MethodPost,
			path:   "/order/delete",
		},
	}

	r := newRouter()
	for _, tr := range testRoutes {
		r.addRoute(tr.method, tr.path, mockHandler)
	}

	for n := 0; n < b.N; n++ {
		for _, tc := range testCases {
			r.findRoute(tc.method, tc.path)
		}
	}
}

func BenchmarkFindRoute_Param(b *testing.B) {
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/param/:id",
		},
	}
	mockHandler := func(ctx *Context) {}

	testCases := []struct {
		name   string
		method string
		path   string
	}{
		{
			// 命中 /param/:id
			name:   ":id",
			method: http.MethodGet,
			path:   "/param/123",
		},
	}

	r := newRouter()
	for _, tr := range testRoutes {
		r.addRoute(tr.method, tr.path, mockHandler)
	}

	for n := 0; n < b.N; n++ {
		for _, tc := range testCases {
			r.findRoute(tc.method, tc.path)
		}
	}
}

func BenchmarkFindRoute_Regex(b *testing.B) {
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodDelete,
			path:   "/reg/:id(.*)",
		},
	}
	mockHandler := func(ctx *Context) {}

	testCases := []struct {
		name   string
		method string
		path   string
	}{
		{
			// 命中 /reg/:id(.*)
			name:   ":id(.*)",
			method: http.MethodDelete,
			path:   "/reg/123",
		},
	}

	r := newRouter()
	for _, tr := range testRoutes {
		r.addRoute(tr.method, tr.path, mockHandler)
	}

	for n := 0; n < b.N; n++ {
		for _, tc := range testCases {
			r.findRoute(tc.method, tc.path)
		}
	}
}
