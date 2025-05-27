package github

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
	"weaveTest/internal/config"
	"weaveTest/internal/proto/generated"
	internal "weaveTest/internal/server/github/client"
	"weaveTest/internal/server/github/client/errors"
)

type RepositoryDataFetcher struct {
	accessToken        string
	logger             zap.Logger
	paginated          bool
	handleRateLimiting bool
	paginator          *internal.Paginator
}

func NewDataFetcher(logger zap.Logger, accessToken string) *RepositoryDataFetcher {
	return NewDataFetcherWithPagination(logger, accessToken, false)
}

func NewDataFetcherWithPagination(logger zap.Logger, accessToken string, shouldPaginate bool) *RepositoryDataFetcher {
	return &RepositoryDataFetcher{
		logger:             logger,
		accessToken:        accessToken,
		paginated:          shouldPaginate,
		handleRateLimiting: false,
	}
}

func (dataFetcher *RepositoryDataFetcher) Fetch(ctx context.Context, request *generated.SearchRequest) (*generated.SearchResponse, errors.AppError) {

}

func (dataFetcher *RepositoryDataFetcher) genQualifiedUrl(request *generated.SearchRequest) (string, errors.AppError) {

}

func (dataFetcher *RepositoryDataFetcher) fetchDataFromRemote(ctx context.Context, url string) (*generated.SearchResponse, errors.AppError) {

}

func (dataFetcher *RepositoryDataFetcher) handleHttpErrors(statusCode int, body []byte, url string) errors.AppError {
	dataFetcher.logger.Debug("handling http errors", zap.Int("statusCode", statusCode))
	if statusCode == http.StatusOK {
		return nil
	}
	switch statusCode {
	// TODO: Wrap errors in a custom error type
	case http.StatusNotAcceptable:
		return errors.NewRemoteError(fmt.Errorf("query string was wrongly formatted: %s", url), statusCode, string(body))
	case http.StatusGatewayTimeout:
		return errors.NewRemoteError(fmt.Errorf("gateway timed out for the search API"), statusCode, "")
	case http.StatusTooManyRequests:
		return errors.NewRemoteError(fmt.Errorf("rate limit exceeded"), statusCode, "")
	case http.StatusUnauthorized:
		return errors.NewRemoteError(fmt.Errorf("unauthorized, the token being used could either have expired or is invalid"), statusCode, string(body))
	case http.StatusUnprocessableEntity:
		// https://docs.github.com/en/rest/search/search?apiVersion=2022-11-28#access-errors-or-missing-search-results
		return errors.NewRemoteError(fmt.Errorf("either the qualifer provided was invalid or the resource specified in the qualifier cant be accessed"), statusCode, string(body))
	default:
		return nil
	}

}

func (dataFetcher *RepositoryDataFetcher) bodyBytesToError(body []byte) (*internal.HttpErrorResponse, error) {
	var errResponse internal.HttpErrorResponse
	err := json.Unmarshal(body, &errResponse)
	if err != nil {
		dataFetcher.logger.Error("failed to unmarshal error response body: ", zap.Error(err))
		return nil, err
	}
	return &errResponse, nil
}
