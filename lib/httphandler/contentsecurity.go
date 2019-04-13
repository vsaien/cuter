package httphandler

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/vsaien/cuter/lib/codec"
	"github.com/vsaien/cuter/lib/httpx"
	"github.com/vsaien/cuter/lib/iox"
	"github.com/vsaien/cuter/lib/logx"
)

const (
	requestUriHeader = "X-Request-Uri"
	signatureField   = "signature"
	timeField        = "time"
	timeTolerance    = 300 // seconds
)

var (
	ErrInvalidContentType = errors.New("invalid content type")
	ErrInvalidHeader      = errors.New("invalid X-Content-Security header")
	ErrInvalidKey         = errors.New("invalid key")
	ErrInvalidPublicKey   = errors.New("invalid public key")
	ErrInvalidSecret      = errors.New("invalid secret")
)

type ContentSecurityHeader struct {
	Key         []byte
	Timestamp   string
	ContentType int
	Signature   string
}

func (h *ContentSecurityHeader) Encrypted() bool {
	return h.ContentType == CryptionType
}

func ParseContentSecurity(decrypters map[string]codec.RsaDecrypter, r *http.Request) (*ContentSecurityHeader, error) {
	contentSecurity := r.Header.Get(ContentSecurity)
	attrs := httpx.ParseHeader(contentSecurity)
	fingerprint := attrs[KeyField]
	secret := attrs[SecretField]
	signature := attrs[signatureField]

	if len(fingerprint) == 0 || len(secret) == 0 || len(signature) == 0 {
		return nil, ErrInvalidHeader
	}

	decrypter, ok := decrypters[fingerprint]
	if !ok {
		return nil, ErrInvalidPublicKey
	}

	decryptedSecret, err := decrypter.DecryptBase64(secret)
	if err != nil {
		return nil, ErrInvalidSecret
	}

	attrs = httpx.ParseHeader(string(decryptedSecret))
	base64Key := attrs[KeyField]
	timestamp := attrs[timeField]
	contentType := attrs[TypeField]

	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, ErrInvalidKey
	}

	cType, err := strconv.Atoi(contentType)
	if err != nil {
		return nil, ErrInvalidContentType
	}

	return &ContentSecurityHeader{
		Key:         key,
		Timestamp:   timestamp,
		ContentType: cType,
		Signature:   signature,
	}, nil
}

func VerifySignature(r *http.Request, securityHeader *ContentSecurityHeader) bool {
	seconds, err := strconv.ParseInt(securityHeader.Timestamp, 10, 64)
	if err != nil {
		return false
	}

	now := time.Now().Unix()
	if seconds+timeTolerance < now || now+timeTolerance < seconds {
		return false
	}

	reqPath, reqQuery := getPathQuery(r)
	signContent := strings.Join([]string{
		securityHeader.Timestamp,
		r.Method,
		reqPath,
		reqQuery,
		computeBodySignature(r),
	}, "\n")
	actualSignature := codec.HmacBase64(securityHeader.Key, signContent)

	passed := securityHeader.Signature == actualSignature
	if !passed {
		logx.Infof("signature different, expect: %s, actual: %s", securityHeader.Signature, actualSignature)
	}

	return passed
}

func computeBodySignature(r *http.Request) string {
	var dup io.ReadCloser
	r.Body, dup = iox.DupReadCloser(r.Body)
	sha := sha256.New()
	io.Copy(sha, r.Body)
	r.Body = dup
	return fmt.Sprintf("%x", sha.Sum(nil))
}

func getPathQuery(r *http.Request) (string, string) {
	requestUri := r.Header.Get(requestUriHeader)
	if len(requestUri) == 0 {
		return r.URL.Path, r.URL.RawQuery
	}

	uri, err := url.Parse(requestUri)
	if err != nil {
		return r.URL.Path, r.URL.RawQuery
	}

	return uri.Path, uri.RawQuery
}
