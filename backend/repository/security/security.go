package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"

	"github.com/yukihito-jokyu/topic2html/backend/domain/auth"
)

// ServiceはServer限定の乱数・ハッシュ・暗号化を提供します。
type Service struct {
	key          [32]byte
	randomSource io.Reader
	newAES       func([]byte) (cipher.Block, error)
}

// Newは保護鍵からAES-256-GCMの保護サービスを作成します。
func New(protectionKey string) (*Service, error) {
	return newService(protectionKey, rand.Reader, aes.NewCipher)
}

func newService(protectionKey string, randomSource io.Reader, newAES func([]byte) (cipher.Block, error)) (*Service, error) {
	if protectionKey == "" {
		return nil, errors.New("protection key is empty")
	}

	return &Service{
		key:          sha256.Sum256([]byte(protectionKey)),
		randomSource: randomSource,
		newAES:       newAES,
	}, nil
}

// RandomBytesは暗号学的に安全な乱数を返します。
func (s *Service) RandomBytes(size int) ([]byte, error) {
	if s == nil || size <= 0 {
		return nil, errors.New("random size is invalid")
	}
	value := make([]byte, size)
	if _, err := io.ReadFull(s.randomSource, value); err != nil {
		return nil, errors.New("random generation failed")
	}

	return value, nil
}

// Hashは保護値を保存用にハッシュします。
func (s *Service) Hash(value []byte) auth.Hash {
	digest := sha256.Sum256(value)

	return append(auth.Hash(nil), digest[:]...)
}

// Sealは保護値を暗号化します。nonceはciphertextの先頭に含めます。
func (s *Service) Seal(value []byte) (auth.Ciphertext, error) {
	if s == nil {
		return nil, errors.New("security service is nil")
	}
	gcm, err := s.gcm()
	if err != nil {
		return nil, err
	}
	nonce, err := s.RandomBytes(gcm.NonceSize())
	if err != nil {
		return nil, err
	}

	return append(nonce, gcm.Seal(nil, nonce, value, nil)...), nil
}

// OpenはSealで暗号化された値を復号します。
func (s *Service) Open(value auth.Ciphertext) ([]byte, error) {
	if s == nil {
		return nil, errors.New("security service is nil")
	}
	gcm, err := s.gcm()
	if err != nil || len(value) < gcm.NonceSize() {
		return nil, errors.New("ciphertext is invalid")
	}
	nonce, ciphertext := value[:gcm.NonceSize()], value[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("ciphertext is invalid")
	}

	return plaintext, nil
}

func (s *Service) gcm() (cipher.AEAD, error) {
	block, err := s.newAES(s.key[:])
	if err != nil {
		return nil, errors.New("cipher initialization failed")
	}

	return cipher.NewGCM(block)
}
