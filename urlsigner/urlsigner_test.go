package urlsigner

import (
	"crypto/rand"
	"encoding/base64"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func randomKey(tb testing.TB, size ...int) []byte {
	tb.Helper()

	defaultSize := 32

	if len(size) > 0 {
		defaultSize = size[0]
	}

	bytes := make([]byte, defaultSize)
	n, err := rand.Read(bytes)
	if err != nil {
		tb.Fatal(err)
	}

	if n != len(bytes) {
		tb.Fatal("not enough bytes read")
	}

	return bytes
}

func Test_NewBlake2B_KeyLengthPanicsAndBounds(t *testing.T) {
	t.Parallel()
	// too short
	require.Panics(t, func() { New("sha256", randomKey(t, 31)) })
	// too long
	require.Panics(t, func() { New("sha256", randomKey(t, 65)) })
	// boundary 32 OK
	_ = New("sha256", randomKey(t, 32))
	// boundary 64 OK
	_ = New("sha256", randomKey(t, 64))
}

func Test_Sign_And_Verify_Success_NoExpiration(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 8, 29, 20, 0, 0, 0, time.UTC)
	s := New("sha256", randomKey(t, 32), func() time.Time { return now })

	signed, err := s.Sign("https://example.com/path?hello=world", 0)
	require.NoError(t, err)

	u, err := url.Parse(signed)
	require.NoError(t, err)
	require.Empty(t, u.Query().Get(Expiration))
	require.NotEmpty(t, u.Query().Get(Signature))

	require.NoError(t, s.Verify(u))
}

func Test_Sign_And_Verify_Success_WithFutureExpiration(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 8, 29, 20, 0, 0, 123000000, time.UTC)
	dur := 2 * time.Hour
	s := New("sha256", randomKey(t, 32), func() time.Time { return now })

	// Sign should place expires = now + dur
	signed, err := s.Sign("https://example.com/resource", dur)
	require.NoError(t, err)

	u, err := url.Parse(signed)
	require.NoError(t, err)
	expiresStr := u.Query().Get(Expiration)
	require.NotEmpty(t, expiresStr)

	expires, err := time.Parse(time.RFC3339Nano, expiresStr)
	require.NoError(t, err)
	require.Equal(t, now.Add(dur), expires)

	require.NoError(t, s.Verify(u))
}

func Test_Verify_Expired(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 8, 29, 20, 0, 0, 0, time.UTC)
	s := New("sha256", randomKey(t, 32), func() time.Time { return now })

	// Create a signed URL with future expiration, then verify after time advanced beyond expiry
	signed, err := s.Sign("https://example.com/asset", time.Minute)
	require.NoError(t, err)

	u, _ := url.Parse(signed)
	// Move time forward 2 minutes to force expiration
	s.now = func() time.Time { return now.Add(2 * time.Minute) }
	require.Equal(t, ErrExpired, s.Verify(u))
}

func Test_Verify_MissingSignature(t *testing.T) {
	t.Parallel()
	u, _ := url.Parse("https://example.com/x?y=z")
	s := New("sha256", randomKey(t, 32))
	require.Equal(t, ErrMissingSignature, s.Verify(u))
}

func Test_Verify_InvalidSignature_Base64(t *testing.T) {
	t.Parallel()
	u, _ := url.Parse("https://example.com/x?signature=***not_base64***")
	s := New("sha256", randomKey(t, 32))
	require.Equal(t, ErrInvalidSignature, s.Verify(u))
}

func Test_Verify_InvalidSignature_Tampered(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 8, 29, 20, 0, 0, 0, time.UTC)
	s := New("sha256", randomKey(t, 32), func() time.Time { return now })

	signed, err := s.Sign("https://example.com/path?a=1", 0)
	require.NoError(t, err)

	u, _ := url.Parse(signed)
	q := u.Query()
	q.Set("a", "2") // tamper with query param
	u.RawQuery = q.Encode()
	require.Equal(t, ErrInvalidSignature, s.Verify(u))
}

func Test_Sign_FormatAndBase64(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 8, 29, 20, 0, 0, 456000000, time.UTC)
	dur := 90 * time.Minute
	s := New("sha256", randomKey(t, 32), func() time.Time { return now })

	signed, err := s.Sign("https://example.com/path", dur)
	require.NoError(t, err)

	u, err := url.Parse(signed)
	require.NoError(t, err)

	sig := u.Query().Get(Signature)
	require.NotEmpty(t, sig)
	_, err = base64.RawURLEncoding.DecodeString(sig)
	require.NoError(t, err)
}

func Test_Verify_InvalidExpires_ParseErrorBranch(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 8, 29, 20, 0, 0, 0, time.UTC)
	key := randomKey(t, 32)
	s := New("sha256", key, func() time.Time { return now })

	// Manually craft URL with invalid expires, then sign that exact URL string using hasher
	u, _ := url.Parse("https://example.com/path?hello=world&" + Expiration + "=not-a-time")
	h := s.hasher()
	_, _ = h.Write([]byte(u.String()))
	sig := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	q := u.Query()
	q.Set(Signature, sig)
	u.RawQuery = q.Encode()
	// Verify should attempt to parse expires and return a parsing error (not our custom errors)
	err := s.Verify(u)
	require.Error(t, err)
}

func Test_Sign_InvalidURL_ReturnsError(t *testing.T) {
	t.Parallel()
	s := New("sha256", randomKey(t, 32))
	_, err := s.Sign("http://%zz", 0)
	require.Error(t, err)
}

func Test_Sign_NoQuery_NoDuration(t *testing.T) {
	t.Parallel()
	s := New("sha256", randomKey(t, 32))
	signed, err := s.Sign("https://example.com/plain", 0)
	require.NoError(t, err)
	u, err := url.Parse(signed)
	require.NoError(t, err)
	require.Empty(t, u.Query().Get(Expiration))
	require.NotEmpty(t, u.Query().Get(Signature))
	require.NoError(t, s.Verify(u))
}

func Test_Sign_Uses_DefaultNow(t *testing.T) {
	t.Parallel()
	dur := 5 * time.Second
	s := New("sha256", randomKey(t, 32))
	signed, err := s.Sign("https://example.com/time", dur)
	require.NoError(t, err)
	u, err := url.Parse(signed)
	require.NoError(t, err)
	expiresStr := u.Query().Get(Expiration)
	require.NotEmpty(t, expiresStr)
	_, err = time.Parse(time.RFC3339Nano, expiresStr)
	require.NoError(t, err)
}

func Test_Sign_WithQuery_WithDuration(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 8, 29, 21, 0, 0, 0, time.UTC)
	dur := 10 * time.Minute
	s := New("sha256", randomKey(t, 32), func() time.Time { return now })
	signed, err := s.Sign("https://example.com/a?b=c", dur)
	require.NoError(t, err)
	u, err := url.Parse(signed)
	require.NoError(t, err)
	q := u.Query()
	require.NotEmpty(t, q.Get(Expiration))
	require.NotEmpty(t, q.Get(Signature))
}

func Test_Verify_MissingSignature_EmptyValue(t *testing.T) {
	t.Parallel()
	u, _ := url.Parse("https://example.com/x?signature=")
	s := New("sha256", randomKey(t, 32))
	require.Equal(t, ErrMissingSignature, s.Verify(u))
}

func Test_Verify_EmptyExpires_Value(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 8, 29, 22, 0, 0, 0, time.UTC)
	key := randomKey(t, 32)
	s := New("sha256", key, func() time.Time { return now })
	// create URL with empty expires but valid signature for that exact URL
	u, _ := url.Parse("https://example.com/z?" + Expiration + "=")
	h := s.hasher()
	_, _ = h.Write([]byte(u.String()))
	sig := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	q := u.Query()
	q.Set(Signature, sig)
	u.RawQuery = q.Encode()
	require.NoError(t, s.Verify(u))
}

func Test_Verify_Expiration_EqualNow(t *testing.T) {
	t.Parallel()
	now := time.Date(2025, 8, 29, 23, 0, 0, 0, time.UTC)
	key := randomKey(t, 32)
	s := New("sha256", key, func() time.Time { return now })
	// Create URL with expires exactly at now
	expires := url.QueryEscape(now.Format(time.RFC3339Nano))
	u, _ := url.Parse("https://example.com/e?" + Expiration + "=" + expires)
	h := s.hasher()
	_, _ = h.Write([]byte(u.String()))
	sig := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	q := u.Query()
	q.Set(Signature, sig)
	u.RawQuery = q.Encode()
	// Since expires == now, it should NOT be considered expired
	require.NoError(t, s.Verify(u))
}
