package directory

import (
	"github.com/alvidir/filebrowser/file"
)

type Path string

type Directory struct {
	id     string
	userId int32
	shared map[Path]*file.File
	hosted map[Path]*file.File
}

func NewDirectory(userId int32) *Directory {
	return &Directory{
		id:     "",
		userId: userId,
		shared: make(map[Path]*file.File),
		hosted: make(map[Path]*file.File),
	}
}
