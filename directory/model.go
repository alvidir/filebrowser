package directory

import (
	"github.com/alvidir/filebrowser/file"
)

type Path string

type Directory struct {
	Id     string
	UserId int32
	Shared map[Path]*file.File
	Hosted map[Path]*file.File
}
