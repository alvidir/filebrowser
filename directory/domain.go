package directory

import (
	"fmt"
	"path"
	"strings"

	"github.com/alvidir/filebrowser/file"
)

const (
	PathSeparator = "/"
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

func (dir *Directory) AddFile(file *file.File, path string) string {
	path = dir.getAvailablePath(path)
	dir.files[path] = file
	return path
}

func (dir *Directory) RemoveFile(file *file.File) {
	for path, f := range dir.files {
		if f.Id() == file.Id() {
			delete(dir.files, path)
		}
	}
}

func (dir *Directory) Files() map[string]*file.File {
	return dir.files
}

func (dir *Directory) List(target string) map[string]*file.File {
	if len(target) == 0 {
		return dir.files
	}

	filtered := make(map[string]*file.File)
	depth := len(strings.Split(target, PathSeparator))
	for path, f := range dir.files {
		if !strings.HasPrefix(path, target) {
			continue
		}

		name := strings.Split(path, PathSeparator)[depth]
		if _, exists := filtered[name]; exists {
			// if the same filename appears more than once, then it is a directory name
			filtered[name], _ = file.NewFile("", name)
			filtered[name].SetFlag(file.Directory)
		} else {
			filtered[name] = f
		}
	}

	return filtered
}
