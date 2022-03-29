package filebrowser

const (
	Public  = 0x01
	Private = 0x02
	Read    = 0x04
	Write   = 0x08
	Share   = 0x16
	Owner   = 0x32
)

type User struct {
	Id     string
	UserID int32
	Shared map[string]Path
	Owned  map[string]Path
}

type File struct {
	Name     string
	Value    []byte
	Metadata map[string]string
}

type Path struct {
	Name        string
	Flags       uint8
	Permissions map[int32]uint8
	Paths       map[string]Path
	Files       map[string]File
}
