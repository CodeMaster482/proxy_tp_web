package requests

import (
	"context"
	"fmt"
	"net/http"

	"proxy/internal/api/repository"
	"proxy/internal/models"
	"proxy/pkg/logger"

	reqUtils "proxy/pkg/http"
)

type Usecase struct {
	Repo repository.Repository
	log  logger.Logger
}

func NewUsecase(r repository.Repository, log logger.Logger) *Usecase {
	return &Usecase{
		Repo: r,
		log:  log,
	}
}

func (u *Usecase) GetAllRequests(ctx context.Context) ([]models.Request, error) {
	requests, err := u.Repo.GetAllRequests(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return requests, nil
}

func (u *Usecase) GetRequestById(ctx context.Context, id uint64) (*models.Request, error) {
	request, err := u.Repo.GetRequestById(ctx, id)
	if err != nil {
		return &models.Request{}, fmt.Errorf("%w", err)
	}
	return request, nil
}

func (u *Usecase) RepeatRequest(ctx context.Context, id uint64) (*http.Request, error) {
	req, err := u.Repo.GetRequestById(ctx, id)
	if err != nil {
		return &http.Request{}, err
	}
	ri, err := reqUtils.MakeRequest(req)
	if err != nil {
		return &http.Request{}, err
	}

	return ri, nil
}

func (u *Usecase) ScanRequest(ctx context.Context, id uint64) (string, error) {
	// TODO: Scan request
	return "", nil
}

func (u *Usecase) SaveRequest(ctx context.Context, request models.Request) (uint64, error) {
	id, err := u.Repo.SaveRequest(ctx, request)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (u *Usecase) SaveResponse(ctx context.Context, response models.Response) error {
	if err := u.Repo.SaveResponse(ctx, response); err != nil {
		return err
	}

	return nil
}
