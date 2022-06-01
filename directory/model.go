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
		dest = path.Join(directory, fmt.Sprintf("%s (%v)", filename, counter))
		counter++
	}

	return dest
}

func (dir *Directory) Files() map[string]*file.File {
	return dir.files
}

func (dir *Directory) AddFile(file *file.File, path string) string {
	path = dir.getAvailablePath(path)
	dir.files[path] = file
	return path
}

func (dir *Directory) RemoveFile(file *file.File) {
	for path, subject := range dir.files {
		if subject.Id() == file.Id() {
			delete(dir.files, path)
		}
	}
}

func (dir *Directory) List() map[string]string {
	list := make(map[string]string)
	if dir.files == nil {
		return list
	}

	for path, file := range dir.files {
		list[path] = file.Id()
	}

	return list
}
