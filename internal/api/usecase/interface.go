package usecase

import (
	"context"
	"net/http"
	"proxy/internal/models"
)

type Usecase interface {
	GetAllRequests(ctx context.Context) ([]models.Request, error)
	GetRequestById(ctx context.Context, id uint64) (*models.Request, error)
	RepeatRequest(ctx context.Context, id uint64) (*http.Request, error)
	ScanRequest(ctx context.Context, param string, request *models.Request) (string, error)

	SaveRequest(ctx context.Context, request models.Request) (uint64, error)
	SaveResponse(ctx context.Context, response models.Response) error
}
