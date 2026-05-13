package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"backend/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

const (
	AccessTokenHeaderName  = "X-Access-Token"
	RefreshTokenCookieName = "refresh_token"
)

var (
	ErrMissingSecret  = errors.New("jwt secret is not configured")
	ErrInvalidToken   = errors.New("invalid token")
	ErrInvalidTokenTy = errors.New("invalid token type")
)

type TokenIdentity struct {
	UserID uint
	Email  string
	Role   string
	RoleID uint
}

type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}

type AccessTokenDetails struct {
	Token     string
	ExpiresAt time.Time
	JTI       string
}

type RefreshTokenDetails struct {
	Token     string
	ExpiresAt time.Time
	JTI       string
}

type TokenService struct {
	secret      string
	store       *InMemoryTokenStore
	accessTTL   time.Duration
	refreshTTL  time.Duration
	logoutGrace time.Duration
}

func NewTokenService(secret string, store *InMemoryTokenStore) *TokenService {
	if store == nil {
		store = NewInMemoryTokenStore()
	}

	return &TokenService{
		secret:      secret,
		store:       store,
		accessTTL:   20 * time.Minute,
		refreshTTL:  7 * 24 * time.Hour,
		logoutGrace: 5 * time.Minute,
	}
}

func NewTokenIdentityFromUser(user *models.User) TokenIdentity {
	return TokenIdentity{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role.Name,
		RoleID: user.RoleID,
	}
}

func (s *TokenService) HasSecret() bool {
	return s.secret != ""
}

func (s *TokenService) Secret() string {
	return s.secret
}

func (s *TokenService) AccessTTL() time.Duration {
	return s.accessTTL
}

func (s *TokenService) RefreshTTL() time.Duration {
	return s.refreshTTL
}

func (s *TokenService) LogoutGrace() time.Duration {
	return s.logoutGrace
}

func (s *TokenService) IssueTokenPair(identity TokenIdentity) (TokenPair, error) {
	access, err := s.IssueAccessToken(identity)
	if err != nil {
		return TokenPair{}, err
	}

	refresh, err := s.IssueRefreshToken(identity)
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:      access.Token,
		RefreshToken:     refresh.Token,
		AccessExpiresAt:  access.ExpiresAt,
		RefreshExpiresAt: refresh.ExpiresAt,
	}, nil
}

func (s *TokenService) IssueAccessToken(identity TokenIdentity) (AccessTokenDetails, error) {
	if s.secret == "" {
		return AccessTokenDetails{}, ErrMissingSecret
	}

	token, exp, jti, err := s.issueToken(identity, "access", s.accessTTL)
	if err != nil {
		return AccessTokenDetails{}, err
	}

	return AccessTokenDetails{Token: token, ExpiresAt: exp, JTI: jti}, nil
}

func (s *TokenService) IssueRefreshToken(identity TokenIdentity) (RefreshTokenDetails, error) {
	if s.secret == "" {
		return RefreshTokenDetails{}, ErrMissingSecret
	}

	token, exp, jti, err := s.issueToken(identity, "refresh", s.refreshTTL)
	if err != nil {
		return RefreshTokenDetails{}, err
	}

	s.store.StoreRefreshToken(token, RefreshTokenRecord{UserID: identity.UserID, ExpiresAt: exp})
	return RefreshTokenDetails{Token: token, ExpiresAt: exp, JTI: jti}, nil
}

func (s *TokenService) RefreshAccessToken(refreshToken string) (AccessTokenDetails, TokenIdentity, error) {
	identity, err := s.ValidateRefreshToken(refreshToken)
	if err != nil {
		return AccessTokenDetails{}, TokenIdentity{}, err
	}

	access, err := s.IssueAccessToken(identity)
	if err != nil {
		return AccessTokenDetails{}, TokenIdentity{}, err
	}

	return access, identity, nil
}

func (s *TokenService) ValidateRefreshToken(tokenStr string) (TokenIdentity, error) {
	token, err := s.parseToken(tokenStr)
	if err != nil || !token.Valid {
		return TokenIdentity{}, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return TokenIdentity{}, ErrInvalidToken
	}

	typ, _ := claims["typ"].(string)
	if typ != "refresh" {
		return TokenIdentity{}, ErrInvalidTokenTy
	}

	identity, err := identityFromClaims(claims)
	if err != nil {
		return TokenIdentity{}, ErrInvalidToken
	}

	record, ok := s.store.GetRefreshToken(tokenStr, time.Now())
	if !ok || record.UserID != identity.UserID {
		return TokenIdentity{}, ErrInvalidToken
	}

	return identity, nil
}

func (s *TokenService) RevokeRefreshToken(tokenStr string) {
	s.store.DeleteRefreshToken(tokenStr)
}

func (s *TokenService) BlacklistAccessToken(jti string) {
	if jti == "" {
		return
	}

	s.store.SetAccessTokenLogout(jti, time.Now().Add(s.logoutGrace))
}

func (s *TokenService) CheckAccessTokenLogout(jti string) (deny bool, inGrace bool) {
	return s.store.CheckAccessTokenLogout(jti, time.Now())
}

func (s *TokenService) issueToken(identity TokenIdentity, tokenType string, ttl time.Duration) (string, time.Time, string, error) {
	now := time.Now()
	exp := now.Add(ttl)
	jti, err := newTokenID()
	if err != nil {
		return "", time.Time{}, "", err
	}

	claims := jwt.MapClaims{
		"user_id": identity.UserID,
		"email":   identity.Email,
		"role":    identity.Role,
		"role_id": identity.RoleID,
		"typ":     tokenType,
		"jti":     jti,
		"iat":     now.Unix(),
		"exp":     exp.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.secret))
	if err != nil {
		return "", time.Time{}, "", err
	}

	return signed, exp, jti, nil
}

func (s *TokenService) parseToken(tokenStr string) (*jwt.Token, error) {
	parser := jwt.NewParser(jwt.WithLeeway(2 * time.Minute))
	return parser.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenSignatureInvalid
		}

		return []byte(s.secret), nil
	})
}

func identityFromClaims(claims jwt.MapClaims) (TokenIdentity, error) {
	userID, ok := getUintClaim(claims, "user_id")
	if !ok {
		return TokenIdentity{}, ErrInvalidToken
	}

	email, ok := claims["email"].(string)
	if !ok {
		email = ""
	}

	role, ok := claims["role"].(string)
	if !ok {
		role = ""
	}

	roleID, ok := getUintClaim(claims, "role_id")
	if !ok {
		roleID = 0
	}

	return TokenIdentity{
		UserID: userID,
		Email:  email,
		Role:   role,
		RoleID: roleID,
	}, nil
}

func newTokenID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func getUintClaim(claims jwt.MapClaims, key string) (uint, bool) {
	value, ok := claims[key]
	if !ok {
		return 0, false
	}

	switch typed := value.(type) {
	case float64:
		return uint(typed), true
	case float32:
		return uint(typed), true
	case int:
		return uint(typed), true
	case int64:
		return uint(typed), true
	case uint:
		return typed, true
	case uint64:
		return uint(typed), true
	default:
		return 0, false
	}
}
