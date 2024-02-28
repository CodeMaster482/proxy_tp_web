package requests

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"proxy/internal/models"
	"proxy/pkg/logger"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	RequestsAll = `SELECT * FROM request ORDER BY created_at`
	RequestById = `SELECT * FROM request WHERE id=$1`
	AddRequest  = `INSERT INTO request (method, host, path, headers, query_params, post_params, cookies, body) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
	AddResponse = `INSERT INTO response (request_id, status_code, http_version, headers, body) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	/* INSERT INTO request (method, "url", body, headers)
	   VALUES ('GET', 'https://example.com', 'body content', '{"Content-Type": "application/json"}'); */
)

type Repository struct {
	db  *pgxpool.Pool
	log logger.Logger
}

func NewRepository(db *pgxpool.Pool, log logger.Logger) *Repository {
	return &Repository{
		db:  db,
		log: log,
	}
}

func (r *Repository) GetAllRequests(ctx context.Context) ([]models.Request, error) {
	var requests []models.Request

	rows, err := r.db.Query(ctx, RequestsAll)
	if err != nil {
		return nil, fmt.Errorf("[repo] Error no tags found: %w, %v", &models.ErrRequestNotFuound{}, err)
	}

	var rawHeaders json.RawMessage
	var rawCookies json.RawMessage
	var rawGetParams json.RawMessage
	var rawPostParams json.RawMessage

	for rows.Next() {
		var request models.Request
		if err := rows.Scan(
			&request.Id,
			&request.Method,
			&request.Host,
			&request.Path,
			&rawHeaders,
			&rawGetParams,
			&rawPostParams,
			&rawCookies,
			&request.Body,
			&request.CreatedAt,
		); err != nil {
			return nil, err
		}

		var headers map[string][]string
		if err := json.Unmarshal(rawHeaders, &headers); err != nil {
			return nil, err
		}

		var query map[string][]string
		if err := json.Unmarshal(rawGetParams, &query); err != nil {
			return nil, err
		}

		var formdata map[string][]string
		if err := json.Unmarshal(rawPostParams, &formdata); err != nil {
			return nil, err
		}

		var cookies map[string]string
		if err := json.Unmarshal(rawCookies, &cookies); err != nil {
			return nil, err
		}

		request.Headers = headers
		request.Cookies = cookies
		request.Get_Params = query
		request.Post_Params = formdata

		requests = append(requests, request)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("[repo] Error rows error: %v", err)
	}

	if len(requests) == 0 {
		return nil, fmt.Errorf("[repo] Error no requests found: %w:%v", &models.ErrRequestNotFuound{}, err)
	}
	return requests, nil
}

func (r *Repository) GetRequestById(ctx context.Context, id uint64) (*models.Request, error) {
	row := r.db.QueryRow(ctx, RequestById, id)
	var req models.Request

	var rawHeaders json.RawMessage
	var rawCookies json.RawMessage
	var rawGetParams json.RawMessage
	var rawPostParams json.RawMessage

	err := row.Scan(
		&req.Id,
		&req.Method,
		&req.Host,
		&req.Path,
		&rawHeaders,
		&rawGetParams,
		&rawPostParams,
		&rawCookies,
		&req.Body,
		&req.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("[repo] %w, %w", &models.ErrRequestNotFuound{}, err)
	} else if err != nil {
		return nil, fmt.Errorf("failed request db %w", err)
	}

	var headers map[string][]string
	if err := json.Unmarshal(rawHeaders, &headers); err != nil {
		return nil, err
	}

	var query map[string][]string
	if err := json.Unmarshal(rawGetParams, &query); err != nil {
		return nil, err
	}

	var formdata map[string][]string
	if err := json.Unmarshal(rawPostParams, &formdata); err != nil {
		return nil, err
	}

	var cookies map[string]string
	if err := json.Unmarshal(rawCookies, &cookies); err != nil {
		return nil, err
	}

	req.Headers = headers
	req.Cookies = cookies
	req.Get_Params = query
	req.Post_Params = formdata

	return &req, nil
}

func (r *Repository) SaveRequest(ctx context.Context, request models.Request) (uint64, error) {
	rawHeaders, err := json.Marshal(&request.Headers)
	if err != nil {
		return 0, err
	}

	rawCookies, err := json.Marshal(&request.Cookies)
	if err != nil {
		return 0, err
	}

	rawQuery, err := json.Marshal(&request.Get_Params)
	if err != nil {
		return 0, err
	}

	rawPostParams, err := json.Marshal(&request.Post_Params)
	if err != nil {
		return 0, err
	}

	row := r.db.QueryRow(ctx, AddRequest,
		request.Method,
		request.Host,
		request.Path,
		rawHeaders,
		rawQuery,
		rawPostParams,
		rawCookies,
		request.Body,
	)

	if err := row.Scan(&request.Id); err != nil {
		return 0, err
	}

	return request.Id, nil
}

func (r *Repository) SaveResponse(ctx context.Context, response models.Response) error {
	rawHeaders, err := json.Marshal(&response.Headers)
	if err != nil {
		return err
	}

	row := r.db.QueryRow(ctx, AddResponse,
		response.RequestId,
		response.Code,
		rawHeaders,
		response.Body,
	)

	if err := row.Scan(&response.Id); err != nil {
		return err
	}

	return nil
}
