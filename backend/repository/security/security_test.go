package security

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"io"
	"testing"

	domainauth "github.com/yukihito-jokyu/topic2html/backend/domain/auth"
)

func TestService(t *testing.T) {
	service, err := New("test-protection-key")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	sealed, err := service.Seal([]byte("secret"))
	if err != nil {
		t.Fatalf("Seal() error = %v", err)
	}
	if bytes.Equal(sealed, []byte("secret")) {
		t.Fatal("sealed value is plaintext")
	}
	opened, err := service.Open(sealed)
	if err != nil || !bytes.Equal(opened, []byte("secret")) {
		t.Fatalf("Open() = %q, %v", opened, err)
	}
	hash := service.Hash([]byte("secret"))
	if len(hash) != 32 || bytes.Equal(hash, []byte("secret")) {
		t.Fatalf("Hash() = %x", hash)
	}
	random, err := service.RandomBytes(16)
	if err != nil || len(random) != 16 {
		t.Fatalf("RandomBytes() = %d, %v", len(random), err)
	}
	modified := append(domainauth.Ciphertext(nil), sealed...)
	modified[len(modified)-1] ^= 1
	var nilService *Service
	randomFailureService, err := newService("test-protection-key", errorReader{}, aes.NewCipher)
	if err != nil {
		t.Fatalf("newService() error = %v", err)
	}
	cipherFailureService, err := newService("test-protection-key", nil, func([]byte) (cipher.Block, error) {
		return nil, errors.New("cipher failure")
	})
	if err != nil {
		t.Fatalf("newService() error = %v", err)
	}
	tests := []struct {
		name string
		call func() error
	}{
		{
			name: "empty key",
			call: func() error {
				_, err := New("")

				return err
			},
		},
		{
			name: "zero random size",
			call: func() error {
				_, err := service.RandomBytes(0)

				return err
			},
		},
		{
			name: "negative random size",
			call: func() error {
				_, err := service.RandomBytes(-1)

				return err
			},
		},
		{
			name: "short ciphertext",
			call: func() error {
				_, err := service.Open(domainauth.Ciphertext("short"))

				return err
			},
		},
		{
			name: "tampered ciphertext",
			call: func() error {
				_, err := service.Open(modified)

				return err
			},
		},
		{
			name: "nil Seal",
			call: func() error {
				_, err := nilService.Seal(nil)

				return err
			},
		},
		{
			name: "nil Open",
			call: func() error {
				_, err := nilService.Open(nil)

				return err
			},
		},
		{
			name: "nil RandomBytes",
			call: func() error {
				_, err := nilService.RandomBytes(1)

				return err
			},
		},
		{
			name: "random reader failure",
			call: func() error {
				_, err := randomFailureService.RandomBytes(1)

				return err
			},
		},
		{
			name: "random reader failure during Seal",
			call: func() error {
				_, err := randomFailureService.Seal([]byte("secret"))

				return err
			},
		},
		{
			name: "cipher initialization failure during Seal",
			call: func() error {
				_, err := cipherFailureService.Seal([]byte("secret"))

				return err
			},
		},
		{
			name: "cipher initialization failure during Open",
			call: func() error {
				_, err := cipherFailureService.Open(sealed)

				return err
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); err == nil {
				t.Fatal("operation succeeded")
			}
		})
	}
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }
