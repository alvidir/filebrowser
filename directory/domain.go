package directory

import (
	"fmt"
	"path"

	fb "github.com/alvidir/filebrowser"
	"github.com/alvidir/filebrowser/file"
)

type FilterFileFn func(string, *file.File) (string, *file.File)

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
	components := fb.PathComponents(dest)
	for index := 0; index < len(components); index++ {
		candidate := components[index]
		counter := 1
		for {
			subject := fb.NormalizePath(path.Join(components[0 : index+1]...))
			if _, exists := dir.files[subject]; !exists {
				break
			}

			components[index] = fmt.Sprintf("%s (%v)", candidate, counter)
			counter++
		}
	}

	return fb.NormalizePath(path.Join(components...))
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

func (dir *Directory) FilterFiles(filters []FilterFileFn) (map[string]*file.File, error) {
	filtered := make(map[string]*file.File)
	for p, file := range dir.files {
		selected := file
		key := p

		for _, filter := range filters {
			if filter == nil {
				continue
			}

			if key, selected = filter(p, file); selected == nil {
				break
			}
		}

		if selected != nil {
			filtered[key] = selected
		}
	}

	return filtered, nil
}
