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

type fileAccessClaims struct {
	jwt.RegisteredClaims `json:",inline"`
	FileId               string `json:"file_id"`
	Read                 bool   `json:"can_read"`
	Write                bool   `json:"can_write"`
	Owner                bool   `json:"is_owner"`
}

func newFileAccessClaims(cert *FileAccessCertificate, ttl time.Duration) (*fileAccessClaims, error) {
	if len(cert.id) == 0 {
		return nil, fb.ErrUnidentified
	}

	return &fileAccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    TokenIssuer,
			Subject:   strconv.Itoa(int(cert.userId)),
			Audience:  []string{},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			NotBefore: &jwt.NumericDate{},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        cert.id,
		},
		FileId: cert.fileId,
		Read:   cert.read,
		Write:  cert.write,
		Owner:  cert.owner,
	}, nil
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

func (service *CertificateService) SignCertificate(cert *FileAccessCertificate) error {
	claims, err := newFileAccessClaims(cert, *service.ttl)
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
