package file

import (
	"regexp"
	"strconv"
	"time"

	fb "github.com/alvidir/filebrowser"
)

const (
	Directory Flag = 0x01

	Read  Permission = 0x01
	Write Permission = 0x02
	Owner Permission = 0x04

	FilenameRegex string = "^[^/]+$"

	MetadataCreatedAtKey = "created_at"
	MetadataUpdatedAtKey = "updated_at"
	MetadataDeletedAtKey = "deleted_at"
	MetadataSizeKey      = "size"
	MetadataAppKey       = "app"
	MetadataRefKey       = "ref"

	TimestampBase = 16
)

var (
	r, _ = regexp.Compile(FilenameRegex)
)

type Permission uint8
type Metadata map[string]string
type Flag uint8
type Ctrl uint8

type File struct {
	id          string
	name        string
	metadata    Metadata
	directory   string
	permissions map[int32]Permission
	protected   bool // true avoids the file from saving
	flags       Flag
	data        []byte
}

func NewFile(id string, filename string) (*File, error) {
	if !r.MatchString(filename) {
		return nil, fb.ErrRegexNotMatch
	}

	meta := make(Metadata)

	meta[MetadataCreatedAtKey] = strconv.FormatInt(time.Now().Unix(), TimestampBase)
	meta[MetadataUpdatedAtKey] = meta[MetadataCreatedAtKey]

	return &File{
		id:          id,
		name:        filename,
		metadata:    meta,
		permissions: make(map[int32]Permission),
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

func (file *File) Flags() Flag {
	return file.flags
}

func (file *File) Metadata() Metadata {
	return file.metadata
}

func (file *File) Directory() string {
	return file.directory
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

func (file *File) SetID(id string) {
	file.id = id
}

func (file *File) SetName(name string) {
	file.name = name
}

func (file *File) Permission(uid int32) (perm Permission) {
	if file.permissions != nil {
		perm = file.permissions[uid]
	}

	return
}

func (file *File) AddPermission(uid int32, perm Permission) {
	if file.permissions == nil {
		file.permissions = make(map[int32]Permission)
	}

	file.permissions[uid] |= perm
}

func (file *File) RevokePermission(uid int32, perm Permission) {
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

func (file *File) AddMetadata(key string, value string) (old string, exists bool) {
	if file.metadata == nil {
		file.metadata = make(Metadata)
	}

	old, exists = file.metadata[key]
	file.metadata[key] = value
	return
}

func (file *File) SetDirectory(dir string) {
	file.directory = dir
}

func (file *File) Data() []byte {
	return file.data
}

// ProtectFields clear from the file all those fields the given user has no permissions
// to read.
func (file *File) ProtectFields(uid int32) {
	if file.IsContributor(uid) {
		// if the user itself is contributor it has the right to know
		// who can read and write the file
		return
	}

	file.MarkAsProtected()
	for id, p := range file.permissions {
		// if the user has read-only permissions it has the right to know
		// who are the contributors of the file
		if id != uid && p&(Owner|Write) == 0 {
			delete(file.permissions, id)
		}
	}
}

func (file *File) MarkAsProtected() {
	file.protected = true
}

func (file *File) IsContributor(uid int32) bool {
	// is contributor if, and only if, the user is owner or has write permissions
	return file.permissions[uid]&(Owner|Write) != 0
}

func (file *File) SetFlag(flag Flag) {
	file.flags |= flag
}
