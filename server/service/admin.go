package service

import (
	"ach/internal/models"

	"github.com/gin-gonic/gin"
)

type GetUsersService struct{}

func (s *GetUsersService) Handle(c *gin.Context) (any, error) {
	return models.GetUserList()
}