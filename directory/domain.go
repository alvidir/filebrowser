package directory

import (
	"fmt"
	"path"
	"strings"

	fb "github.com/alvidir/filebrowser"
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

func (dir *Directory) FilesByPath(target string) (map[string]*file.File, error) {
	if !path.IsAbs(target) {
		target = path.Join(PathSeparator, target)
	}

	if target == PathSeparator {
		target = ""
	}

	filtered := make(map[string]*file.File)
	depth := len(strings.Split(target, PathSeparator))
	for p, f := range dir.files {
		if !path.IsAbs(p) {
			p = path.Join(PathSeparator, p)
		}

		if strings.Compare(p, target) == 0 {
			return nil, fb.ErrNotFound
		} else if !strings.HasPrefix(p, target) {
			continue
		}

		items := strings.Split(p, PathSeparator)
		name := items[depth]
		if _, exists := filtered[name]; !exists && len(items) > depth+1 {
			filtered[name], _ = file.NewFile("", name)
			filtered[name].SetFlag(file.Directory)
		} else {
			filtered[name] = f
		}
	}

	return filtered, nil
}
