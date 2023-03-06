package directory

import (
	"fmt"
	"path"
	"path/filepath"
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
	path   string
}

func NewDirectory(userId int32) *Directory {
	return &Directory{
		id:     "",
		userId: userId,
		files:  make(map[string]*file.File),
	}
}

func pathComponents(p string) []string {
	paths := strings.Split(p, PathSeparator)

	components := make([]string, 0, len(paths))
	components = append(components, PathSeparator)

	for _, p := range paths {
		if len(p) > 0 {
			components = append(components, p)
		}
	}

	return components
}

func (dir *Directory) getAvailablePath(dest string) string {
	components := pathComponents(dest)
	for index := 0; index < len(components); index++ {
		candidate := components[index]
		counter := 1

		for {
			subject := filepath.Join(components[0 : index+1]...)
			if _, exists := dir.files[subject]; !exists {
				break
			}

			components[index] = fmt.Sprintf("%s_%d", candidate, counter)
			counter++
		}
	}

	return filepath.Join(components...)
}

func (dir *Directory) AddFile(file *file.File, fp string) string {
	fp = dir.getAvailablePath(fp)
	file.SetDirectory(path.Dir(fp))
	dir.files[fp] = file
	return fp
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
