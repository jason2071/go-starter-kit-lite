package platform

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/jason2071/go-starter-kit-lite/internal/usecase"
)

type Security struct {
	secret    []byte
	issuer    string
	audience  string
	accessTTL time.Duration
}

type accessTokenClaims struct {
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

func NewSecurity(
	secret string,
	issuer string,
	audience string,
	accessTTL time.Duration,
) *Security {
	return &Security{
		secret:    []byte(secret),
		issuer:    issuer,
		audience:  audience,
		accessTTL: accessTTL,
	}
}

func (s *Security) GenerateAccessToken(
	userID string,
	roles []string,
) (string, error) {
	now := time.Now().UTC()

	claims := accessTokenClaims{
		Roles: roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    s.issuer,
			Audience:  jwt.ClaimStrings{s.audience},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
		},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString(s.secret)
}

func (s *Security) ParseAccessToken(raw string) (*usecase.AccessClaims, error) {
	claims := &accessTokenClaims{}

	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(s.issuer),
		jwt.WithAudience(s.audience),
	)

	token, err := parser.ParseWithClaims(
		raw,
		claims,
		func(*jwt.Token) (any, error) {
			return s.secret, nil
		},
	)
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid access token")
	}

	return &usecase.AccessClaims{
		Subject: claims.Subject,
		Roles:   claims.Roles,
	}, nil
}

func (s *Security) AccessTokenTTL() time.Duration {
	return s.accessTTL
}

func (s *Security) GenerateRefreshToken() (string, string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", "", err
	}

	rawToken := base64.RawURLEncoding.EncodeToString(randomBytes)
	return rawToken, s.HashRefreshToken(rawToken), nil
}

func (s *Security) HashRefreshToken(raw string) string {
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}

func (s *Security) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	return string(hash), err
}

func (s *Security) Compare(hash, password string) error {
	return bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(password),
	)
}
