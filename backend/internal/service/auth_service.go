package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/2hot4you/aiaos/backend/internal/domain"
	"github.com/2hot4you/aiaos/backend/internal/repository/postgres"
	"github.com/2hot4you/aiaos/backend/pkg/config"
	"github.com/2hot4you/aiaos/backend/pkg/crypto"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

type AuthService struct {
	userRepo *postgres.UserRepo
	rdb      *redis.Client
	jwtCfg   config.JWTConfig
}

func NewAuthService(userRepo *postgres.UserRepo, rdb *redis.Client, jwtCfg config.JWTConfig) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		rdb:      rdb,
		jwtCfg:   jwtCfg,
	}
}

type LoginResult struct {
	Token string       `json:"token"`
	User  *domain.User `json:"user"`
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	// Check lock
	lockKey := fmt.Sprintf("lock:%s", username)
	if s.rdb.Exists(ctx, lockKey).Val() > 0 {
		return nil, domain.ErrAccountLocked
	}

	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		s.incrementFailCount(ctx, username)
		return nil, domain.ErrInvalidCredential
	}

	if !user.Enabled {
		return nil, domain.ErrAccountDisabled
	}

	if !crypto.CheckPassword(password, user.PasswordHash) {
		s.incrementFailCount(ctx, username)
		return nil, domain.ErrInvalidCredential
	}

	// Clear fail count on success
	failKey := fmt.Sprintf("login_fail:%s", username)
	s.rdb.Del(ctx, failKey)

	// Update last login
	_ = s.userRepo.UpdateLastLogin(user.ID)

	// Generate JWT
	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		Token: token,
		User:  user,
	}, nil
}

func (s *AuthService) GetCurrentUser(userID int64) (*domain.User, error) {
	return s.userRepo.FindByID(userID)
}

func (s *AuthService) ChangePassword(userID int64, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return domain.ErrUserNotFound
	}

	if !crypto.CheckPassword(oldPassword, user.PasswordHash) {
		return domain.ErrOldPasswordWrong
	}

	hash, err := crypto.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = hash
	return s.userRepo.Update(user)
}

func (s *AuthService) generateToken(user *domain.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":      strconv.FormatInt(user.ID, 10),
		"username": user.Username,
		"role":     user.Role,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().AddDate(0, 0, s.jwtCfg.ExpireDays).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtCfg.Secret))
}

func (s *AuthService) incrementFailCount(ctx context.Context, username string) {
	failKey := fmt.Sprintf("login_fail:%s", username)
	count := s.rdb.Incr(ctx, failKey).Val()
	s.rdb.Expire(ctx, failKey, 15*time.Minute)

	if count >= 5 {
		lockKey := fmt.Sprintf("lock:%s", username)
		s.rdb.Set(ctx, lockKey, "1", 15*time.Minute)
	}
}

func (s *AuthService) ValidateToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.jwtCfg.Secret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}
