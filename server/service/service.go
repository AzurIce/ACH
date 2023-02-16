package service

import (
	"ach/internal/serializer"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Service interface {
	Handle(c *gin.Context) (any, error)
}

func Handler(s Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := c.BindJSON(s)
		if err != nil && err != io.EOF {
			c.JSON(http.StatusBadRequest, serializer.ErrorResponse(err))
			return
		}

		res, err := s.Handle(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, serializer.ErrorResponse(err))
		} else {
			c.JSON(http.StatusOK, serializer.Response(res))
		}
	}
}