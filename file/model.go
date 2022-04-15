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

func NewFile(filename string, data []byte) *File {
	return &File{
		id:          "",
		name:        filename,
		metadata:    make(Metadata),
		permissions: make(Permissions),
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

func (file *File) Value(key string) (value string, exists bool) {
	if file.metadata != nil {
		value, exists = file.metadata[key]
	}

	return
}

func (file *File) Permissions(uid int32) (perm uint8) {
	if file.permissions == nil {
		perm = file.permissions[uid]
	}

	return
}

func (file *File) AddPermissions(uid int32, perm uint8) {
	file.permissions[uid] |= perm
}

func (file *File) AddValue(key string, value string) (old string, exists bool) {
	old, exists = file.metadata[key]
	file.metadata[key] = value
	return
}

func (file *File) SetId(id string) {
	file.id = id
}
