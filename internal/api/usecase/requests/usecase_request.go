package requests

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"

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

func (u *Usecase) ScanRequest(ctx context.Context, param string, request *models.Request) (string, error) {
	paramVal, _ := generateRandomString(generateRandomNumber())

	ri, err := reqUtils.MakeRequest(request)
	if err != nil {
		return "", err
	}

	newURL := new(url.URL)
	*newURL = *ri.URL
	query := newURL.Query()
	query.Add(param, paramVal)
	newURL.RawQuery = query.Encode()

	ri.URL = newURL

	resp, err := http.DefaultClient.Do(ri)
	if err != nil {
		u.log.Errorf("[usecase] error on scan: %v\n", err)
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		u.log.Errorf("[usecase] error reading response body: %v\n", err)
		return "", err
	}

	if strings.Contains(string(body), paramVal) {
		return param, nil
	}

	return "", nil
}

func generateRandomNumber() int {
	max := big.NewInt(20)
	randomNumber, _ := rand.Int(rand.Reader, max)
	return int(randomNumber.Int64()) + 1
}

func generateRandomString(length int) (string, error) {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	randomString := base64.URLEncoding.EncodeToString(randomBytes)
	return randomString[:length], nil
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
