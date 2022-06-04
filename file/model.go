package file

import (
	"regexp"

	fb "github.com/alvidir/filebrowser"
)

const (
	Public        uint8  = 0x01
	Read          uint8  = 0x02
	Write         uint8  = 0x04
	Grant         uint8  = 0x08
	Owner         uint8  = 0x10
	FilenameRegex string = "^[^/]+$"
)

var (
	r, _ = regexp.Compile(FilenameRegex)
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

func NewFile(id string, filename string) (*File, error) {
	if !r.MatchString(filename) {
		return nil, fb.ErrInvalidFormat
	}

	return &File{
		id:          id,
		name:        filename,
		metadata:    make(Metadata),
		permissions: make(Permissions),
		flags:       0,
		data:        make([]byte, 0),
	}, nil
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

func (file *File) Owners() []int32 {
	owners := make([]int32, 1) // a file has, for sure, at least one owner
	for uid, perm := range file.permissions {
		if perm&Owner == 0 {
			continue
		}

		if owners[0] != 0 {
			owners = append(owners, uid)
		} else {
			owners[0] = uid
		}
	}

	return owners
}

func (file *File) SharedWith() []int32 {
	shared := make([]int32, len(file.permissions))

	index := 0
	for uid := range file.permissions {
		shared[index] = uid
		index++
	}

	return shared
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

func (file *File) RevokePermissions(uid int32, perm uint8) {
	if file.permissions == nil {
		return
	}

	if p, exists := file.permissions[uid]; !exists {
		return
	} else if perm = p & ^perm; perm == 0 {
		delete(file.permissions, uid)
	} else {
		file.permissions[uid] = perm
	}
}

func (file *File) RevokeAccess(uid int32) bool {
	if file.permissions == nil {
		return false
	}

	if _, exists := file.permissions[uid]; !exists {
		return false
	}

	delete(file.permissions, uid)
	return true
}

func (file *File) AddValue(key string, value string) (old string, exists bool) {
	if file.metadata == nil {
		file.metadata = make(Metadata)
	}

	old, exists = file.metadata[key]
	file.metadata[key] = value
	return
}
