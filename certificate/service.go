package certificate

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"

	fb "github.com/alvidir/filebrowser"
)

const (
	TokenHeaderType = "JWT"
	userIdBase      = 16
	userIdBitSize   = 32
)

var (
	TokenAlgorithm = jwt.SigningMethodES256
	tokenHeader    = map[string]interface{}{
		"typ": TokenHeaderType,
		"alg": TokenAlgorithm.Alg(),
	}
)

type JWTCertificateService struct {
	signKey *ecdsa.PrivateKey
	ttl     *time.Duration
	issuer  string
	logger  *zap.Logger
}

func NewCertificateService(sign *ecdsa.PrivateKey, issuer string, ttl *time.Duration, logger *zap.Logger) *JWTCertificateService {
	return &JWTCertificateService{
		signKey: sign,
		ttl:     ttl,
		issuer:  issuer,
		logger:  logger,
	}
}

func (service *JWTCertificateService) SignFileAccessCertificate(cert *FileAccessCertificate) error {
	claims, err := cert.Claims(service.ttl, service.issuer)
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

	cert.token = signed
	return nil
}

func (service *JWTCertificateService) ParseFileAccessCertificate(tokenStr string) (*FileAccessCertificate, error) {
	claims := new(FileAccessClaims)
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return &service.signKey.PublicKey, nil
	})

	if err != nil || !token.Valid {
		service.logger.Error("parsing jwt",
			zap.String("token", tokenStr),
			zap.Bool("is_valid", token.Valid),
			zap.Error(err))

		return nil, fb.ErrInvalidToken
	}

	userId, err := strconv.ParseInt(claims.Subject, userIdBase, userIdBitSize)
	if err != nil {
		service.logger.Error("parsing string to int32",
			zap.Error(err))

		return nil, fb.ErrUnknown
	}

	return &FileAccessCertificate{
		id:         claims.ID,
		fileId:     claims.FileId,
		userId:     int32(userId),
		permission: claims.Permission(),
		token:      tokenStr,
	}, nil
}

func ParsePKCS8PrivateKey(b64 string) (*ecdsa.PrivateKey, error) {
	raw, err := base64.RawStdEncoding.DecodeString(b64)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode([]byte(raw))
	if block == nil {
		return nil, fb.ErrNotFound
	}

	value, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	privateKey, ok := value.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fb.ErrUnknown
	}

	return privateKey, nil
}
