package auth

import "github.com/golang-jwt/jwt/v5"

type JWTAuthenticator struct {
	secret []byte
	aud    string // Audience
	iss    string // Issuer
}

func NewJWTAuthenticator(secret, audience, issuer string) *JWTAuthenticator {
	return &JWTAuthenticator{
		secret: []byte(secret),
		aud:    audience,
		iss:    issuer,
	}
}

func (j *JWTAuthenticator) GenerateToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *JWTAuthenticator) ValidateToken(token string) (*jwt.Token, error) {
	jwtKeyFn := func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return j.secret, nil
	}

	return jwt.Parse(token, jwtKeyFn,
		jwt.WithExpirationRequired(),
		jwt.WithAudience(j.aud),
		jwt.WithIssuer(j.iss),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)
}
