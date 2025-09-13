package urlsigner

import (
	"crypto/hmac"
	"encoding/base64"
	"errors"
	"hash"
	"net/url"
	"time"

	"github.com/CodeLieutenant/utils"
)

type (
	Signer interface {
		Sign(string, time.Duration) (string, error)
		Verify(*url.URL) error
	}

	HMACSigner struct {
		now    func() time.Time
		hasher func() hash.Hash
	}
)

const (
	Expiration = "expires"
	Signature  = "signature"
)

var (
	ErrMissingSignature = errors.New("missing signature")
	ErrInvalidSignature = errors.New("invalid signature")
	ErrExpired          = errors.New("url expired")
)

func New(algo string, keyBytes []byte, now ...func() time.Time) *HMACSigner {
	if len(keyBytes) > 64 || len(keyBytes) < 32 {
		panic("key must be greater then 32 and less than 64 bytes")
	}

	n := func() time.Time {
		return time.Now().UTC()
	}

	if len(now) > 0 {
		n = now[0]
	}

	return &HMACSigner{
		now: n,
		hasher: func() hash.Hash {
			return hmac.New(utils.ParseHasher(algo), keyBytes)
		},
	}
}

func (s *HMACSigner) Sign(urlString string, duration time.Duration) (string, error) {
	u, err := url.Parse(urlString)
	if err != nil {
		return "", err
	}

	query := u.Query()

	if duration != 0 {
		query.Set("expires", s.now().Add(duration).Format(time.RFC3339Nano))
	}

	u.RawQuery = query.Encode()
	sigBytes := s.sumURLString(u.String())
	signature := base64.RawURLEncoding.EncodeToString(sigBytes)

	if len(u.RawQuery) > 0 {
		u.RawQuery += "&" + Signature + "=" + signature
	} else {
		u.RawQuery = Signature + "=" + signature
	}

	return u.String(), nil
}

func (s *HMACSigner) sumURLString(str string) []byte {
	h := s.hasher()
	_, _ = h.Write(utils.UnsafeBytes(str))

	return h.Sum(nil)
}

func (s *HMACSigner) extractAndDecodeSignature(query url.Values) ([]byte, url.Values, error) {
	signature := query.Get(Signature)
	if signature == "" {
		return nil, query, ErrMissingSignature
	}

	sigBytes, err := base64.RawURLEncoding.DecodeString(signature)
	if err != nil {
		return nil, query, ErrInvalidSignature
	}

	query.Del(Signature)

	return sigBytes, query, nil
}

func (s *HMACSigner) verifyMAC(original string, sigBytes []byte) bool {
	computed := s.sumURLString(original)

	return hmac.Equal(sigBytes, computed)
}

func (s *HMACSigner) verifyExpiration(query url.Values) error {
	if len(query.Get(Expiration)) == 0 {
		return nil
	}
	expires, err := time.Parse(time.RFC3339Nano, query.Get(Expiration))
	if err != nil {
		panic("urlsigner: invalid expiration format, this should never happen since we validated the signature, CHECK THE FORMATTING FOR THE TIME")
	}
	if expires.Before(s.now()) {
		return ErrExpired
	}

	return nil
}

func (s *HMACSigner) Verify(u *url.URL) error {
	query := u.Query()
	sigBytes, query, err := s.extractAndDecodeSignature(query)
	if err != nil {
		return err
	}

	u.RawQuery = query.Encode()
	if !s.verifyMAC(u.String(), sigBytes) {
		return ErrInvalidSignature
	}

	return s.verifyExpiration(query)
}
