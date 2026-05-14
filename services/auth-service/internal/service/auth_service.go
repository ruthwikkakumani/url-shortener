package service

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/ruthwikkakumani/redirection-engine/services/auth-service/internal/model"
	"github.com/ruthwikkakumani/redirection-engine/services/auth-service/internal/repository"
	"go.uber.org/zap"
)

type AuthService struct {
	logger   *zap.Logger
	userRepo *repository.UserRepo
}

func NewAuthService(logger *zap.Logger, userRepo *repository.UserRepo) *AuthService {
	return &AuthService{
		logger:   logger,
		userRepo: userRepo,
	}
}

func (s *AuthService) RegisterService(name string, email string, password string) (error){
	
	// Validate if user already exists
	user, err := s.userRepo.GetUserByEmail(email)
	if err == nil && user != nil {
		s.logger.Warn("email already exists",
			zap.String("email", email),
		)
		return fmt.Errorf("user already exists")
	}
	
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			s.logger.Error("failed to check user existence",
				zap.Error(err),
			)
			return err
		}
	
	hashedPassword, err := HashPassword(password)
	if err != nil {
		s.logger.Error("failed to hash password",
			zap.Error(err),
		)
		return err
	}
	
	return s.userRepo.CreateUser(&model.User{
			Name:     name,
			Email:    email,
			Password: hashedPassword,
	})
}

func (s *AuthService) LoginService(email string, password string) (string, error){
	
	user, err := s.userRepo.GetUserByEmail(email)
	
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.logger.Warn("user not found",
				zap.String("email", email),
			)
			return "", fmt.Errorf("invalid credentials")
		}
		
		s.logger.Error("failed to fetch db",
			zap.Error(err),
		)
		
		return "", err
	}
	
	if !VerifyPassword(user.Password, password) {
		s.logger.Warn("invalid credentials",
			zap.String("email", email),
		)
		
		return "", fmt.Errorf("invalid credentials")
	}
	
	jwtToken, err := generateJWT(user.ID)
	if err != nil {
		s.logger.Error("failed to generate JWT token",
			zap.Error(err),
		)
		
		return "", fmt.Errorf("failed to generate JWT token")
	}
	
	s.logger.Info("login successful",
		zap.String("user_id", user.ID),
	)
	
	return jwtToken, nil
}

func (s *AuthService) RequestPasswordReset(email string) error {
	_, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.logger.Warn("password reset requested for non-existent email", zap.String("email", email))
			return nil // Don't leak user existence
		}
		return err
	}

	token, err := generateResetToken(email)
	if err != nil {
		return err
	}

	// In a real app, send an email. For now, log it.
	s.logger.Info("password reset token generated",
		zap.String("email", email),
		zap.String("token", token),
	)

	return nil
}

func (s *AuthService) ResetPassword(email, token, newPassword string) error {
	tokenEmail, err := verifyResetToken(token)
	if err != nil {
		return errors.New("invalid or expired token")
	}

	if tokenEmail != email {
		return errors.New("token does not match email")
	}

	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePassword(email, hashedPassword)
}