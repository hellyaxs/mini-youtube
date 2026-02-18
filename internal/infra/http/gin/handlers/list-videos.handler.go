package handlers

import (
	"net/http"
	"strconv"

	gin "github.com/gin-gonic/gin"
	"github.com/hellyaxs/miniyoutube/internal/application/usecase"
)

const defaultPageSize = 20

// ListVideosHandler retorna um handler Gin para GET /api/v1/videos (paginado).
func ListVideosHandler(uc *usecase.ListVideosUseCase) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", strconv.Itoa(defaultPageSize)))
		list, err := uc.Execute(c.Request.Context(), page, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"videos":    list,
			"page":      page,
			"page_size": pageSize,
		})
	}
}
