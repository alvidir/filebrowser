package file

const (
	Public  = 0x01
	Private = 0x02
	Read    = 0x04
	Write   = 0x08
	Share   = 0x16
	Owner   = 0x32
)

type Metadata map[string]string
type Permissions map[int32]uint8

type File struct {
	id          string
	name        string
	metadata    Metadata
	permissions Permissions
	flags       uint8
	data        []byte
}

func NewFile(id string, filename string, data []byte, perm Permissions, meta Metadata) *File {
	return &File{
		id:          id,
		name:        filename,
		metadata:    meta,
		permissions: perm,
		flags:       Private,
		data:        data,
	}
}

func (file *File) Id() string {
	return file.id
}

func (file *File) Name() string {
	return file.name
}
