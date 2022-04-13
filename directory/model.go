package directory

import (
	"fmt"
	"path"

	"github.com/alvidir/filebrowser/file"
)

type Directory struct {
	id     string
	userId int32
	files  map[string]*file.File
}

func NewDirectory(userId int32) *Directory {
	return &Directory{
		id:     "",
		userId: userId,
		files:  make(map[string]*file.File),
	}
}

func (dir *Directory) getAvailablePath(dest string) string {
	filename := path.Base(dest)
	directory := path.Dir(dest)

	counter := 1
	for _, exists := dir.files[dest]; exists; _, exists = dir.files[dest] {
		dest = path.Join(directory, fmt.Sprintf("%s_%v", filename, counter))
		counter++
	}

	return dest
}

func (dir *Directory) AddFile(file *file.File, path string) string {
	path = dir.getAvailablePath(path)
	dir.files[path] = file
	return path
}
