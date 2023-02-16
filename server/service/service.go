package service

import (
	"ach/internal/serializer"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Service interface {
	Handle(c *gin.Context) (any, error)
}

func Handler(s Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := c.ShouldBindJSON(s)
		if err != nil && err != io.EOF {
			c.JSON(http.StatusBadRequest, serializer.ErrorResponse(err))
			return
		}

		res, err := s.Handle(c)
		if err != nil {
			log.Println(err.Error())
			c.JSON(http.StatusBadRequest, serializer.ErrorResponse(err))
		} else {
			log.Println("StatusOK")
			c.JSON(http.StatusOK, serializer.Response(res))
		}
	}
}