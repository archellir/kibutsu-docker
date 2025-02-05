package main

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed frontend/build/*
var staticFiles embed.FS

func GetFileSystem() http.FileSystem {
	fsys, err := fs.Sub(staticFiles, "frontend/build")
	if err != nil {
		panic(err)
	}
	return http.FS(fsys)
}
