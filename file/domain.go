package file

import (
	"regexp"
	"strconv"
	"time"

	fb "github.com/alvidir/filebrowser"
)

const (
	Blurred Flags = 0x01
	Remote  Flags = 0x02

	FilenameRegex string = "^[^/]+$"

	MetadataCreatedAtKey = "created_at"
	MetadataUpdatedAtKey = "updated_at"
	MetadataDeletedAtKey = "deleted_at"
	MetadataUrlKey       = "url"
	MetadataAppKey       = "app"

	TimestampBase = 16
)

var (
	r, _ = regexp.Compile(FilenameRegex)
)

type Metadata map[string]string
type Flags uint8

type File struct {
	id          string
	name        string
	metadata    Metadata
	permissions map[int32]fb.Permissions
	flags       Flags
	data        []byte
}

func NewFile(id string, filename string) (*File, error) {
	if !r.MatchString(filename) {
		return nil, fb.ErrInvalidFormat
	}

	meta := make(Metadata)

	meta[MetadataCreatedAtKey] = strconv.FormatInt(time.Now().Unix(), TimestampBase)
	meta[MetadataUpdatedAtKey] = meta[MetadataCreatedAtKey]

	return &File{
		id:          id,
		name:        filename,
		metadata:    meta,
		permissions: make(map[int32]fb.Permissions),
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

func (file *File) Metadata() Metadata {
	return file.metadata
}

func (file *File) Owners() []int32 {
	owners := make([]int32, 1) // a file has, for sure, at least one owner
	for uid, perm := range file.permissions {
		if perm&fb.Owner == 0 {
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

func (file *File) Permissions(uid int32) (perm fb.Permissions) {
	if file.permissions != nil {
		perm = file.permissions[uid]
	}

	return
}

func (file *File) AddPermissions(uid int32, perm fb.Permissions) {
	if file.permissions == nil {
		file.permissions = make(map[int32]fb.Permissions)
	}

	file.permissions[uid] |= perm
}

func (file *File) RevokePermissions(uid int32, perm fb.Permissions) {
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

func (file *File) Data() []byte {
	return file.data
}

func (file *File) HideProtectedFields(uid int32) {
	if file.IsContributor(uid) {
		// if the user itself is contributor it has the right to know
		// who can read and write the file
		return
	}

	file.flags |= Blurred
	for id, p := range file.permissions {
		// if the user has read-only permissions it has the right to know
		// who are the contributors of the file
		if id != uid && p&(fb.Owner|fb.Write) == 0 {
			delete(file.permissions, id)
		}
	}
}

func (file *File) IsRemote() bool {
	_, exists := file.metadata[MetadataAppKey]
	return exists
}

func (file *File) IsContributor(uid int32) bool {
	// is contributor if, and only if, the user is owner or has write permissions
	return file.permissions[uid]&(fb.Owner|fb.Write) != 0
}
