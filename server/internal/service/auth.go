package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"flowtask-server/internal/config"
	"flowtask-server/internal/model"
	"flowtask-server/internal/repository"
)

const refreshTokenKeyPrefix = "refresh_token:"

type AuthService struct {
	userRepo *repository.UserRepository
	rdb      *redis.Client
	jwtCfg   config.JWTConfig
}

func NewAuthService(userRepo *repository.UserRepository, rdb *redis.Client, jwtCfg config.JWTConfig) *AuthService {
	return &AuthService{userRepo: userRepo, rdb: rdb, jwtCfg: jwtCfg}
}

type RegisterInput struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"display_name" binding:"required,min=2,max=100"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	User         *model.User `json:"user"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
}

func (s *AuthService) Register(input RegisterInput) (*AuthResponse, error) {
	existing, _ := s.userRepo.FindByEmail(input.Email)
	if existing != nil {
		return nil, errors.New("email already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &model.User{
		Email:        input.Email,
		PasswordHash: string(hash),
		DisplayName:  input.DisplayName,
		Preferences:  model.JSONB{},
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return s.generateTokens(user)
}

func (s *AuthService) Login(input LoginInput) (*AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(input.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, fmt.Errorf("find user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	return s.generateTokens(user)
}

func (s *AuthService) Logout(userID, refreshToken string) error {
	ctx := context.Background()
	key := s.refreshTokenKey(userID, refreshToken)
	return s.rdb.Del(ctx, key).Err()
}

func (s *AuthService) RefreshToken(refreshToken string) (string, string, error) {
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtCfg.RefreshSecret), nil
	})
	if err != nil || !token.Valid {
		return "", "", errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", errors.New("invalid token claims")
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return "", "", errors.New("invalid user ID in token")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return "", "", errors.New("invalid user ID format")
	}

	// Verify the refresh token exists in Redis (not revoked)
	ctx := context.Background()
	key := s.refreshTokenKey(userID, refreshToken)
	exists, err := s.rdb.Exists(ctx, key).Result()
	if err != nil {
		return "", "", fmt.Errorf("check refresh token in redis: %w", err)
	}
	if exists == 0 {
		return "", "", errors.New("refresh token has been revoked or expired")
	}

	user, err := s.userRepo.FindByID(uid)
	if err != nil {
		return "", "", errors.New("user not found")
	}

	newAccess, err := s.generateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	newRefresh, err := s.generateRefreshToken(user)
	if err != nil {
		return "", "", err
	}

	// Revoke old refresh token and store the new one
	s.rdb.Del(ctx, key)
	if err := s.storeRefreshToken(user.ID.String(), newRefresh); err != nil {
		return "", "", fmt.Errorf("store new refresh token: %w", err)
	}

	return newAccess, newRefresh, nil
}

func (s *AuthService) generateTokens(user *model.User) (*AuthResponse, error) {
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, err
	}

	// Store refresh token in Redis for active revocation
	if err := s.storeRefreshToken(user.ID.String(), refreshToken); err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	return &AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) generateAccessToken(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":   user.ID.String(),
		"email": user.Email,
		"exp":   time.Now().Add(s.jwtCfg.AccessTTL).Unix(),
		"iat":   time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtCfg.AccessSecret))
}

func (s *AuthService) generateRefreshToken(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"sub": user.ID.String(),
		"exp": time.Now().Add(s.jwtCfg.RefreshTTL).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtCfg.RefreshSecret))
}

// refreshTokenKey builds a Redis key for storing a refresh token.
// Format: refresh_token:{user_id}:{sha256(token)}
func (s *AuthService) refreshTokenKey(userID, refreshToken string) string {
	hash := sha256.Sum256([]byte(refreshToken))
	tokenHash := hex.EncodeToString(hash[:])
	return fmt.Sprintf("%s%s:%s", refreshTokenKeyPrefix, userID, tokenHash)
}

// storeRefreshToken persists a refresh token in Redis with TTL matching the refresh expiry.
func (s *AuthService) storeRefreshToken(userID, refreshToken string) error {
	ctx := context.Background()
	key := s.refreshTokenKey(userID, refreshToken)
	return s.rdb.Set(ctx, key, "1", s.jwtCfg.RefreshTTL).Err()
}
