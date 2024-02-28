package http

import (
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"proxy/internal/api/usecase"
	"proxy/internal/models"
	"proxy/pkg/logger"

	// "proxy/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Logger  logger.Logger
	Usecase usecase.Usecase
}

func NewHandler(logger logger.Logger, uc usecase.Usecase) *Handler {
	return &Handler{
		Logger:  logger,
		Usecase: uc,
	}
}

func (h *Handler) GetRequests(ctx *gin.Context) {
	requests, err := h.Usecase.GetAllRequests(ctx.Request.Context())
	if err != nil {
		var errNoRequests *models.ErrRequestNotFuound
		if errors.As(err, &errNoRequests) {
			ctx.JSON(http.StatusNoContent, gin.H{"error": err.Error()})
			return
		}
		h.Logger.Errorf("failed to get all requests %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"requests": requests})
}

func (h *Handler) GetRequestById(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id")[:], 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	request, err := h.Usecase.GetRequestById(ctx.Request.Context(), id)
	if err != nil {
		var errNoRequests *models.ErrRequestNotFuound
		if errors.As(err, &errNoRequests) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"request": request})
}

func (h *Handler) RepeatRequest(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id")[:], 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	requestHttp, err := h.Usecase.RepeatRequest(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	repeatedResp, err := http.DefaultClient.Do(requestHttp)
	if err != nil {
		log.Print(err)
		ctx.JSON(http.StatusInternalServerError, "failed to send request to repeat")
		return
	}

	var b []byte
	if b, err = httputil.DumpResponse(repeatedResp, true); err != nil {
		log.Print(err)
		ctx.JSON(http.StatusInternalServerError, "failed to dump response")
	}

	ctx.String(http.StatusOK, string(b))
}

func (h Handler) ScanRequest(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id")[:], 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	scanResult, err := h.Usecase.ScanRequest(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"request_scan": scanResult})
}
