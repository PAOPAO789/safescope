package security

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/safescope/safescope/apps/api/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type BcryptHasher struct {
	Cost int
}

func (h BcryptHasher) Hash(password string) (string, error) {
	cost := h.Cost
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(hash), err
}

func (h BcryptHasher) Compare(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

type Claims struct {
	Email string      `json:"email"`
	Name  string      `json:"name"`
	Role  domain.Role `json:"role"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	Secret []byte
	TTL    time.Duration
}

func (m JWTManager) Issue(user *domain.User) (string, time.Time, error) {
	now := time.Now().UTC()
	expires := now.Add(m.TTL)
	claims := Claims{
		Email: user.Email, Name: user.Name, Role: user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: user.ID, IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(expires),
			Issuer: "safescope-api",
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.Secret)
	return token, expires, err
}

func (m JWTManager) Parse(token string) (*domain.User, error) {
	claims := &Claims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return m.Secret, nil
	}, jwt.WithIssuer("safescope-api"))
	if err != nil || !parsed.Valid || !claims.Role.Valid() {
		return nil, domain.ErrUnauthorized
	}
	return &domain.User{
		ID: claims.Subject, Email: claims.Email, Name: claims.Name, Role: claims.Role,
	}, nil
}
