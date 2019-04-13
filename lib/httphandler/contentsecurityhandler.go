package httphandler

import (
	"net/http"

	"github.com/vsaien/cuter/lib/codec"
	"github.com/vsaien/cuter/lib/logx"
)

const contentSecurity = "X-Content-Security"

func ContentSecurityHandler(decrypters map[string]codec.RsaDecrypter, strict bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodDelete, http.MethodGet, http.MethodPost, http.MethodPut:
				header, err := ParseContentSecurity(decrypters, r)
				if err != nil {
					// TODO: temporarily disable ErrInvalidHeader, because it's flooding, enable it when ready.
					if err != ErrInvalidHeader {
						logx.Infof("Signature verification failed, X-Content-Security: %s, error: %s",
							r.Header.Get(contentSecurity), err.Error())
					}
					handleVerificationFailure(w, r, next, strict)
				} else if !VerifySignature(r, header) {
					logx.Infof("Signature verification failed, X-Content-Security: %s",
						r.Header.Get(contentSecurity))
					handleVerificationFailure(w, r, next, strict)
				} else if r.ContentLength > 0 && header.Encrypted() {
					CryptionHandler(header.Key)(next).ServeHTTP(w, r)
				} else {
					next.ServeHTTP(w, r)
				}
			default:
				next.ServeHTTP(w, r)
			}
		})
	}
}

func handleVerificationFailure(w http.ResponseWriter, r *http.Request, next http.Handler, strict bool) {
	if strict {
		w.WriteHeader(http.StatusUnauthorized)
	} else {
		next.ServeHTTP(w, r)
	}
}
