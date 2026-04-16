package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandler struct {
	logger *zap.Logger
}

func NewAuthHandler(logger *zap.Logger) (*AuthHandler) {
	return &AuthHandler{
	logger: logger,
	}
}

func (h *AuthHandler) RegisterHandler(c *gin.Context) {
	
}

func (h *AuthHandler) LoginHandler(c *gin.Context) {
	
}