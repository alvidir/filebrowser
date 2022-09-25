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
	TokenHeaderType = "JWT"
)

var (
	TokenAlgorithm = jwt.SigningMethodES256
	tokenHeader    = map[string]interface{}{
		"typ": TokenHeaderType,
		"alg": TokenAlgorithm.Alg(),
	}
)

type fileAccessClaims struct {
	jwt.RegisteredClaims `json:",inline"`
	FileId               string `json:"file_id"`
	Read                 bool   `json:"can_read"`
	Write                bool   `json:"can_write"`
	Owner                bool   `json:"is_owner"`
}

func newFileAccessClaims(cert *FileAccessCertificate, ttl time.Duration, issuer string) (*fileAccessClaims, error) {
	if len(cert.id) == 0 {
		return nil, fb.ErrUnidentified
	}

	return &fileAccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   strconv.Itoa(int(cert.userId)),
			Audience:  []string{},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			NotBefore: &jwt.NumericDate{},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        cert.id,
		},
		FileId: cert.fileId,
		Read:   cert.permissions&fb.Read != 0,
		Write:  cert.permissions&fb.Write != 0,
		Owner:  cert.permissions&fb.Owner != 0,
	}, nil
}

type JWTCertificateService struct {
	signKey *ecdsa.PrivateKey
	ttl     *time.Duration
	issuer  string
	logger  *zap.Logger
}

func NewCertificateService(ttl *time.Duration, sign *ecdsa.PrivateKey, logger *zap.Logger) *JWTCertificateService {
	return &JWTCertificateService{
		signKey: sign,
		ttl:     ttl,
		logger:  logger,
	}
}

func (service *JWTCertificateService) SignFileAccessCertificate(cert *FileAccessCertificate) error {
	claims, err := newFileAccessClaims(cert, *service.ttl, service.issuer)
	if err != nil {
		return err
	}

	token := &jwt.Token{
		Raw:       "",
		Method:    TokenAlgorithm,
		Header:    tokenHeader,
		Claims:    claims,
		Signature: "",
		Valid:     false,
	}

	signed, err := token.SignedString(service.signKey)
	if err != nil {
		service.logger.Error("signing jwt",
			zap.Int32("user_id", cert.userId),
			zap.String("file_id", cert.fileId),
			zap.Error(err))

		return fb.ErrUnknown
	}

	cert.token = []byte(signed)
	return nil
}
