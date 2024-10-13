package auth

import "github.com/golang-jwt/jwt/v5"

type jwtAuthenticator struct {
	secret []byte
	aud    string // Audience
	iss    string // Issuer
}

func NewJWTAuthenticator(secret, audience, issuer string) Authenticator {
	return &jwtAuthenticator{
		secret: []byte(secret),
		aud:    audience,
		iss:    issuer,
	}
}

func (j *jwtAuthenticator) GenerateToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *jwtAuthenticator) ValidateToken(token string) (*jwt.Token, error) {
	return nil, nil
}
