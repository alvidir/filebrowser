package certificate

type Permissions interface {
	Read() bool
	Write() bool
	Owner() bool
}

type File interface {
	Permissions(uid int32) Permissions
	Id() string
}

type FileAccessCertificate struct {
	id     string
	fileId string
	userId int32
	read   bool
	write  bool
	owner  bool
	token  []byte
}

func NewFileAccessCertificate(uid int32, file File) *FileAccessCertificate {
	perm := file.Permissions(uid)
	return &FileAccessCertificate{
		id:     "",
		fileId: file.Id(),
		userId: uid,
		read:   perm.Read(),
		write:  perm.Write(),
		owner:  perm.Owner(),
		token:  []byte{},
	}
}
