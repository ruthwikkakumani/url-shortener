package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ruthwikkakumani/url-shortener/services/auth-service/internal/config"
	"github.com/ruthwikkakumani/url-shortener/services/auth-service/internal/service"
	"go.uber.org/zap"
)

type AuthHandler struct {
	logger *zap.Logger
	authService *service.AuthService
}

type registerReq struct {
	Name string `json:"name" binding:"required,min=3"`
	Email string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func NewAuthHandler(logger *zap.Logger, authService *service.AuthService) (*AuthHandler) {
	return &AuthHandler{
		logger: logger,
		authService: authService,
	}
}

// Register handler
func (h *AuthHandler) RegisterHandler(c *gin.Context) {
	
	var req registerReq
	
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid register request", 
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	
	// Forward req to AuthService
	if err := h.authService.RegisterService(req.Name, req.Email, req.Password); err != nil {
		h.logger.Error("register failed",
			zap.String("email", req.Email),
			zap.Error(err),
		)
		
		c.JSON(http.StatusConflict, gin.H{
			"error": "unable to register user",
		})
		
		return 
	}
	
	h.logger.Info("user registered successfully",
		zap.String("user", req.Name),
		zap.String("email", req.Email),
	)
	
	// Response
	c.JSON(http.StatusCreated, gin.H{
    	"message": "user registered successfully",
	})
}

// Login Handler
func (h *AuthHandler) LoginHandler(c *gin.Context) {
	
	var req loginReq
	
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid request", 
			zap.Error(err),
		)
		
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request payload",
		})
		return
	}
	
	token, err := h.authService.LoginService(req.Email, req.Password)
	if err != nil {
		h.logger.Warn("login failed",
			zap.String("email", req.Email),
			zap.Error(err),
		)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error" : "invalid credentials",
		})
		
		return
	}
	
	// Set cookie
	isProd := config.GetEnv("ENV", "development") == "production"

	c.SetCookie(
	    "token",
	    token,
	    3600*24,
	    "/",
	    "",
	    isProd,
	    true,
	)
	
	// Temporary message
	c.JSON(http.StatusOK, gin.H{
		"message": "login successful",
	})
	
}