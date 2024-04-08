package repository

import (
	"context"
	"proxy/internal/models"
)

type Repository interface {
	GetAllRequests(ctx context.Context) ([]models.Request, error)
	GetRequestById(ctx context.Context, id uint64) (*models.Request, error)

	SaveRequest(ctx context.Context, request models.Request) (uint64, error)
	SaveResponse(ctx context.Context, response models.Response) error
}
