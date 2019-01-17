package core

import (
	"net/http"
	"os"
)

type spaFileSystem struct {
	base http.FileSystem
}

func (fs *spaFileSystem) Open(name string) (http.File, error) {
	file, err := fs.base.Open(name)
	if err == nil {
		return file, nil
	}
	if _, ok := err.(*os.PathError); ok {
		return fs.base.Open("/index.html")
	}
	return nil, err
}

// SpaFileSystem wraps a filesystem with a single page app style index redirect.
func SpaFileSystem(base http.FileSystem) http.FileSystem {
	return &spaFileSystem{base}
}
