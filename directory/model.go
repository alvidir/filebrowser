package directory

import (
	"fmt"
	"path"
)

type Directory struct {
	id     string
	userId int32
	shared map[string]string
	hosted map[string]string
}

func NewDirectory(userId int32) *Directory {
	return &Directory{
		id:     "",
		userId: userId,
		shared: make(map[string]string),
		hosted: make(map[string]string),
	}
}

func (dir *Directory) getAvailablePath(fpath string, shared bool) string {
	base := path.Base(fpath)
	dpath := path.Dir(fpath)

	dict := dir.hosted
	if shared {
		dict = dir.shared
	}

	counter := 1
	cpath := fpath
	for _, exists := dict[fpath]; exists; _, exists = dict[fpath] {
		cpath = path.Join(dpath, fmt.Sprintf("%s_%v", base, counter))
	}

	return cpath

}

func (dir *Directory) addFile(fileId, path string, shared bool) {
	if fpath := dir.getAvailablePath(path, shared); shared {
		if dir.shared == nil {
			dir.shared = make(map[string]string)
		}

		dir.shared[fpath] = fileId
	} else {
		if dir.hosted == nil {
			dir.hosted = make(map[string]string)
		}

		dir.hosted[fpath] = fileId
	}
}
