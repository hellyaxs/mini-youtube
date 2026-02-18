package ginapi

import (
	gin "github.com/gin-gonic/gin"
	"github.com/hellyaxs/miniyoutube/internal/application/usecase"
	"github.com/hellyaxs/miniyoutube/internal/infra/http/gin/handlers"
)

// Router configura e retorna o engine Gin com as rotas da API v1.
func Router(uploadUC *usecase.UploadVideoUseCase, listUC *usecase.ListVideosUseCase, maxUploadBytes int64) *gin.Engine {
	r := gin.Default()
	r.MaxMultipartMemory = maxUploadBytes
	v1 := r.Group("/api/v1")
	v1.POST("/upload", handlers.UploadVideoHandler(uploadUC, maxUploadBytes))
	v1.GET("/videos", handlers.ListVideosHandler(listUC))
	return r
}
