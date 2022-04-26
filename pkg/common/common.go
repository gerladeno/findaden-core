package common

import (
	"errors"
	"io"

	"github.com/google/uuid"
)

var (
	ErrUnauthenticated      = errors.New("err user failed to authenticate")
	ErrInvalidSigningMethod = errors.New("err invalid signing method")
	ErrInvalidAccessToken   = errors.New("err invalid access token")
	ErrInvalidPhoneNumber   = errors.New("err invalid phone number")
	ErrPhoneNotFound        = errors.New("err phone not found")
)

func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

type CountingReader struct {
	Reader    io.Reader
	BytesRead int
}

func (r *CountingReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	r.BytesRead += n
	return n, err
}
