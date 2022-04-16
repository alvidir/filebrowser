package file

const (
	Public = 0x01
	Read   = 0x02
	Write  = 0x04
	Grant  = 0x08
	Owner  = 0x10
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

func NewFile(id string, filename string, data []byte) *File {
	return &File{
		id:          id,
		name:        filename,
		metadata:    make(Metadata),
		permissions: make(Permissions),
		flags:       0,
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
	if file.permissions != nil {
		perm = file.permissions[uid]
	}

	return
}

func (file *File) AddPermissions(uid int32, perm uint8) {
	if file.permissions == nil {
		file.permissions = make(Permissions)
	}

	file.permissions[uid] |= perm
}

func (file *File) AddValue(key string, value string) (old string, exists bool) {
	if file.metadata == nil {
		file.metadata = make(Metadata)
	}

	old, exists = file.metadata[key]
	file.metadata[key] = value
	return
}
