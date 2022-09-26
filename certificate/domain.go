package certificate

import (
	fb "github.com/alvidir/filebrowser"
)

type File interface {
	Permission(uid int32) fb.Permission
	Id() string
}

type FileAccessCertificate struct {
	id         string
	fileId     string
	userId     int32
	permission fb.Permission
	token      []byte
}

func NewFileAccessCertificate(uid int32, file File) *FileAccessCertificate {
	perm := file.Permission(uid)
	return &FileAccessCertificate{
		id:         "",
		fileId:     file.Id(),
		userId:     uid,
		permission: perm,
		token:      []byte{},
	}
}
