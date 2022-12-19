package certificate

import (
	"strconv"
	"time"

	fb "github.com/alvidir/filebrowser"
	"github.com/golang-jwt/jwt/v4"
)

type File interface {
	Permission(uid int32) fb.Permission
	Id() string
}

type FileAccessClaims struct {
	jwt.RegisteredClaims `json:",inline"`
	FileId               string `json:"file_id"`
	Read                 bool   `json:"can_read"`
	Write                bool   `json:"can_write"`
	Owner                bool   `json:"is_owner"`
}

func (claims *FileAccessClaims) Permission() (perms fb.Permission) {
	if claims.Read {
		perms |= fb.Read
	}

	if claims.Write {
		perms |= fb.Write
	}

	if claims.Owner {
		perms |= fb.Owner
	}

	return
}

type FileAccessCertificate struct {
	id         string
	fileId     string
	userId     int32
	permission fb.Permission
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
		Read:   cert.permission&fb.Read != 0,
		Write:  cert.permission&fb.Write != 0,
		Owner:  cert.permission&fb.Owner != 0,
	}

	if ttl != nil {
		claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(*ttl))
	}

	return claims, nil
}

func (cert *FileAccessCertificate) Token() string {
	return cert.token
}
