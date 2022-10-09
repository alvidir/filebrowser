package certificate

import (
	"errors"
	"testing"
	"time"

	fb "github.com/alvidir/filebrowser"
	"go.uber.org/zap"
)

const (
	pemBase64 = "LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1JR0hBZ0VBTUJNR0J5cUdTTTQ5QWdFR0NDcUdTTTQ5QXdFSEJHMHdhd0lCQVFRZy9JMGJTbVZxL1BBN2FhRHgKN1FFSGdoTGxCVS9NcWFWMUJab3ZhM2Y5aHJxaFJBTkNBQVJXZVcwd3MydmlnWi96SzRXcGk3Rm1mK0VPb3FybQpmUlIrZjF2azZ5dnBGd0gzZllkMlllNXl4b3ZsaTROK1ZNNlRXVFErTmVFc2ZmTWY2TkFBMloxbQotLS0tLUVORCBQUklWQVRFIEtFWS0tLS0tCg"
)

var (
	privateKey, _ = ParsePKCS8PrivateKey(pemBase64)
)

type FileMock struct {
	permissions fb.Permission
	id          string
}

func (file *FileMock) Permission(uid int32) fb.Permission {
	return file.permissions
}

func (file *FileMock) Id() string {
	return file.id
}

func TestParsePKCS8PrivateKey(t *testing.T) {
	if _, err := ParsePKCS8PrivateKey(pemBase64); err != nil {
		t.Fatal(err)
	}
}

func TestCertificateValidation(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	tests := []struct {
		name    string
		ttl     time.Duration
		isValid bool
	}{
		{name: "alive certificate should not fail", ttl: 1 * time.Hour, isValid: true},
		{name: "dead certificate should fail", ttl: -1 * time.Hour, isValid: false},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := NewCertificateService(privateKey, &test.ttl, logger)
			file := &FileMock{
				permissions: fb.Owner,
				id:          "123",
			}

			var userId int32 = 999
			cert := NewFileAccessCertificate(userId, file)

			var certId string = "123"
			cert.id = certId

			if err := service.SignFileAccessCertificate(cert); err != nil {
				t.Fatalf("%s, signing certificate: %s", test.name, err.Error())
			}

			token := string(cert.token)
			got, err := service.ParseFileAccessCertificate(token)
			if errors.Is(err, fb.ErrInvalidToken) && !test.isValid {
				return
			}

			if err != nil {
				t.Fatalf("%s, parsing certificate: %s", test.name, err.Error())
			}

			if got.fileId != file.id || got.permission != file.permissions {
				t.Errorf("got wrong claims %+v", got)
			}
		})
	}
}
