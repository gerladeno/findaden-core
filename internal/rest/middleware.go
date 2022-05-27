package rest

import (
	"context"
	"crypto/rsa"
	"errors"
	"net/http"
	"strings"

	"github.com/gerladeno/homie-core/pkg/common"
	"github.com/golang-jwt/jwt"
)

type Claims struct {
	jwt.StandardClaims
	UUID string `json:"uuid"`
}

type idType string

const uuidKey idType = `UUID`

func (h *handler) jwtAuth(next http.Handler) http.Handler {
	var fn http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeErrResponse(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 {
			writeErrResponse(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if headerParts[0] != "Bearer" {
			writeErrResponse(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		id, err := parseToken(headerParts[1], h.key)
		switch {
		case err == nil:
		case errors.Is(err, common.ErrInvalidAccessToken):
			writeErrResponse(w, "Unauthorized", http.StatusUnauthorized)
			return
		default:
			h.log.Warnf("err parsing token: %v", err)
			writeErrResponse(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), uuidKey, id))
		next.ServeHTTP(w, r)
	}
	return fn
}

func parseToken(accessToken string, key *rsa.PublicKey) (string, error) {
	token, err := jwt.ParseWithClaims(accessToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, common.ErrInvalidSigningMethod
		}
		return key, nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.UUID, nil
	}
	return "", common.ErrInvalidAccessToken
}
