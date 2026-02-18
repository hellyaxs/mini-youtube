package handlers

import (
	"net/http"

	gin "github.com/gin-gonic/gin"
	"github.com/hellyaxs/miniyoutube/internal/application/usecase"
)

const defaultMaxMultipartMemory = 100 << 20 // 100 MB

// UploadVideoHandler retorna um handler Gin para POST de upload de vídeo.
func UploadVideoHandler(uc *usecase.UploadVideoUseCase, maxBytes int64) gin.HandlerFunc {
	if maxBytes <= 0 {
		maxBytes = defaultMaxMultipartMemory
	}
	return func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "arquivo é obrigatório", "detail": err.Error()})
			return
		}
		title := c.PostForm("title")
		if title == "" {
			title = file.Filename
		}
		f, err := file.Open()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "falha ao abrir arquivo", "detail": err.Error()})
			return
		}
		defer f.Close()
		out, err := uc.Execute(c.Request.Context(), usecase.UploadVideoInput{
			File:     f,
			Name:     file.Filename,
			Title:    title,
			MaxBytes: maxBytes,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{
			"id":        out.VideoID,
			"status":    out.Status,
			"file_path": out.FilePath,
		})
	}
}
