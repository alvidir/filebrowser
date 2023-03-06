package certificate

import (
	"strconv"
	"time"

	fb "github.com/alvidir/filebrowser"
	"github.com/golang-jwt/jwt/v4"
)

const (
	Read  Permission = 0x01
	Write Permission = 0x02
	Owner Permission = 0x04
)

type Permission uint8

type File interface {
	Permission(uid int32) Permission
	Id() string
}

type FileAccessClaims struct {
	jwt.RegisteredClaims `json:",inline"`
	FileId               string `json:"file_id"`
	Read                 bool   `json:"can_read"`
	Write                bool   `json:"can_write"`
	Owner                bool   `json:"is_owner"`
}

func (claims *FileAccessClaims) Permission() (perms Permission) {
	if claims.Read {
		perms |= Read
	}

	if claims.Write {
		perms |= Write
	}

	if claims.Owner {
		perms |= Owner
	}

	return
}

type FileAccessCertificate struct {
	id         string
	fileId     string
	userId     int32
	permission Permission
	token      string
}

func NewFileAccessCertificate(uid int32, file File) *FileAccessCertificate {
	perm := file.Permission(uid)
	return &FileAccessCertificate{
		id:         "",
		fileId:     file.Id(),
		userId:     uid,
		permission: perm,
		token:      "",
	}
}

func (cert *FileAccessCertificate) Claims(ttl *time.Duration, issuer string) (*FileAccessClaims, error) {
	if len(cert.id) == 0 {
		return nil, fb.ErrUnidentified
	}

	claims := &FileAccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   strconv.FormatInt(int64(cert.userId), userIdBase),
			Audience:  []string{},
			NotBefore: &jwt.NumericDate{},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        cert.id,
		},
		FileId: cert.fileId,
		Read:   cert.permission&Read != 0,
		Write:  cert.permission&Write != 0,
		Owner:  cert.permission&Owner != 0,
	}

	if ttl != nil {
		claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(*ttl))
	}

	return claims, nil
}

func (cert *FileAccessCertificate) Token() string {
	return cert.token
}
