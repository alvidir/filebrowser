package certificate

import (
	fb "github.com/alvidir/filebrowser"
)

type File interface {
	Permissions(uid int32) fb.Permissions
	Id() string
}

type FileAccessCertificate struct {
	id          string
	fileId      string
	userId      int32
	permissions fb.Permissions
	token       []byte
}

func NewFileAccessCertificate(uid int32, file File) *FileAccessCertificate {
	perm := file.Permissions(uid)
	return &FileAccessCertificate{
		id:          "",
		fileId:      file.Id(),
		userId:      uid,
		permissions: perm,
		token:       []byte{},
	}
}
