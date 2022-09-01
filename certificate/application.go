package certificate

import (
	"crypto/ecdsa"

	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

type CertificateRepository interface {
}

type CertificateApplication struct {
	certRepo   CertificateRepository
	privateKey ecdsa.PrivateKey
	logger     *zap.Logger
}

func NewCertificateApplication(certRepo CertificateRepository, logger *zap.Logger) *CertificateApplication {
	return &CertificateApplication{
		certRepo: certRepo,
		logger:   logger,
	}
}

func (app *CertificateApplication) Create(permissions uint8) (Certificate, error) {
	token := &jwt.Token{
		Raw:    "",
		Method: jwt.SigningMethodES256,
		Header: map[string]interface{}{
			"typ": "JWT",
			"alg": jwt.SigningMethodES256.Alg(),
		},
		Claims:    jwt.RegisteredClaims{},
		Signature: "",
		Valid:     false,
	}

	token.SignedString(app.privateKey)
	return nil, nil
}
