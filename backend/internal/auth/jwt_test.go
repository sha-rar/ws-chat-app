package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAndParseToken(t *testing.T) {
	SetSecret("test-secret-test-secret-test-secret")
	expiresAt := time.Now().Add(time.Hour)

	token, err := GenerateToken(42, "alice", "session-1", expiresAt)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	claims, err := ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}
	if claims.UserID != 42 || claims.Username != "alice" || claims.ID != "session-1" {
		t.Errorf("claims = %+v, want user 42 alice with session-1", claims)
	}
	if !claims.ExpiresAt.Time.Equal(expiresAt.Truncate(time.Second)) {
		t.Errorf("ExpiresAt = %v, want %v", claims.ExpiresAt.Time, expiresAt.Truncate(time.Second))
	}
}

func TestParseTokenRejectsInvalidTokens(t *testing.T) {
	SetSecret("test-secret-test-secret-test-secret")

	sign := func(method jwt.SigningMethod, key any, claims Claims) string {
		t.Helper()
		s, err := jwt.NewWithClaims(method, claims).SignedString(key)
		if err != nil {
			t.Fatalf("sign: %v", err)
		}
		return s
	}
	valid := Claims{
		UserID:   1,
		Username: "alice",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        "session-1",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	expired := valid
	expired.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Minute))

	tests := map[string]string{
		"garbage":             "not-a-token",
		"wrong secret":        sign(jwt.SigningMethodHS256, []byte("some-other-secret"), valid),
		"expired":             sign(jwt.SigningMethodHS256, []byte("test-secret-test-secret-test-secret"), expired),
		"unexpected alg":      sign(jwt.SigningMethodHS512, []byte("test-secret-test-secret-test-secret"), valid),
		"unsigned (alg none)": sign(jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, valid),
	}

	for name, token := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseToken(token); err == nil {
				t.Error("ParseToken accepted an invalid token")
			}
		})
	}
}
