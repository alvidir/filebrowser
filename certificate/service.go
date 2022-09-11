package certificate

import (
	"crypto/ecdsa"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"

	fb "github.com/alvidir/filebrowser"
)

const (
	TokenIssuer     = "filebrowser.alvidir.com"
	TokenHeaderType = "JWT"
)

var (
	TokenAlgorithm = jwt.SigningMethodES256
	tokenHeader    = map[string]interface{}{
		"typ": TokenHeaderType,
		"alg": TokenAlgorithm.Alg(),
	}
)

type Permissions interface {
	Read() bool
	Write() bool
	Owner() bool
}

type claims struct {
	jwt.RegisteredClaims `json:",inline"`
	FileId               string `json:"file_id"`
	Read                 bool   `json:"can_read"`
	Write                bool   `json:"can_write"`
	Owner                bool   `json:"is_owner"`
}

type CertificateService struct {
	signKey *ecdsa.PrivateKey
	ttl     *time.Duration
	logger  *zap.Logger
}

func NewCertificateService(ttl *time.Duration, sign *ecdsa.PrivateKey, logger *zap.Logger) *CertificateService {
	return &CertificateService{
		signKey: sign,
		ttl:     ttl,
		logger:  logger,
	}
}

func (service *CertificateService) NewUserAuthorization(uid int32, fid string, perm Permissions) (string, error) {
	claims := claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    TokenIssuer,
			Subject:   strconv.Itoa(int(uid)),
			Audience:  []string{},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(*service.ttl)),
			NotBefore: &jwt.NumericDate{},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        "",
		},
		FileId: fid,
		Read:   perm.Read(),
		Write:  perm.Write(),
		Owner:  perm.Owner(),
	}

	token := &jwt.Token{
		Raw:       "",
		Method:    TokenAlgorithm,
		Header:    tokenHeader,
		Claims:    claims,
		Signature: "",
		Valid:     false,
	}

	certificate, err := token.SignedString(service.signKey)
	if err != nil {
		service.logger.Error("signing certificate",
			zap.Int32("user_id", uid),
			zap.String("file_id", fid),
			zap.Error(err))

		return "", fb.ErrUnknown
	}

	return certificate, nil
}
