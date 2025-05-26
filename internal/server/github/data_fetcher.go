package github

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
	"weaveTest/internal/proto/generated"
	appErrors "weaveTest/internal/server/github/errors"
)

const (
	httpMethod = "GET"
	bearerFmt  = "Bearer %s"
)

type RepositoryDataFetcher struct {
	accessToken        string
	logger             zap.Logger
	paginated          bool
	handleRateLimiting bool
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

func (dataFetcher *RepositoryDataFetcher) Fetch(request *generated.SearchRequest) (*generated.SearchResponse, appErrors.AppError) {

	queryEscapedUrl, err := dataFetcher.genQualifiedUrl(request)
	if err != nil {
		dataFetcher.logger.Error("failed to generate qualified url", zap.Error(err))
		return nil, err
	}
	if dataFetcher.handleRateLimiting {
		// TODO: handle rate limiting
	}
	return dataFetcher.fetchDataFromRemote(queryEscapedUrl)
}

func (dataFetcher *RepositoryDataFetcher) genQualifiedUrl(request *generated.SearchRequest) (string, appErrors.AppError) {
	var query = queryString(request.SearchTerm)
	qualifiedUrl := ""
	if err := query.Validate(); err != nil {
		return "", appErrors.NewInternalError(err, appErrors.InvalidQuery, err.Error())
	}
	if request.User == "" {
		qualifiedUrl, _ = query.ToUrl()
	} else {
		qualifiedUrl, _ = query.ToUrlWithUser(request.User)
	}
	queryEscapedUrl := url.QueryEscape(qualifiedUrl)
	dataFetcher.logger.Debug("url verified: ", zap.String("qualifiedUrl", qualifiedUrl),
		zap.String("escapedUrl", queryEscapedUrl))
	return qualifiedUrl, nil
}

func (dataFetcher *RepositoryDataFetcher) fetchDataFromRemote(url string) (*generated.SearchResponse, appErrors.AppError) {
	client := &http.Client{}
	req, err := http.NewRequest(httpMethod, url, nil)

	if err != nil {
		dataFetcher.logger.Error("failed to create httpClient", zap.Error(err))
		return nil, appErrors.NewInternalError(err, appErrors.InvalidHttpClient, "failed to create httpClient")
	}
	// dont need to set this accept header, since text_matches is not being used right now
	// req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("Authorization", fmt.Sprintf(bearerFmt, dataFetcher.accessToken))
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	dataFetcher.logger.Debug("performing http request for url: ", zap.String("url", url))
	res, err := client.Do(req)
	if err != nil {
		dataFetcher.logger.Error("failed to perform http request: ", zap.Error(err))
		return nil, appErrors.NewInternalError(err, appErrors.InvalidHttpClient, "failed to perform http request")
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		dataFetcher.logger.Error("failed to read response body: ", zap.Error(err))
		return nil, appErrors.NewInternalError(err, appErrors.InvalidJSONBody, "failed to read response body")
	}

	if dataFetcher.paginated {
		// TODO: implement pagination
		return nil, nil
	}

	if err := dataFetcher.handleHttpErrors(res.StatusCode, body, url); err != nil {
		return nil, err
	}
	dataFetcher.logger.Debug("successfully fetched data from remote")
	var response CodeSearchResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		dataFetcher.logger.Error("failed to unmarshal response body: ", zap.Error(err))
		return nil, appErrors.NewInternalError(err, appErrors.InvalidJSONBody, string(body))
	}
	var results []*generated.Result
	for _, item := range response.RepositoryItems {
		results = append(results, &generated.Result{
			FileUrl: item.HTMLURL,
			Repo:    item.Repo.HTMLURL,
		})
	}
	dataFetcher.logger.Debug("successfully fetched data from remote", zap.Int("numResults", len(results)))
	return &generated.SearchResponse{Results: results}, nil
}

func (dataFetcher *RepositoryDataFetcher) handleHttpErrors(statusCode int, body []byte, url string) appErrors.AppError {
	dataFetcher.logger.Debug("handling http errors", zap.Int("statusCode", statusCode))
	if statusCode == 200 {
		return nil
	}
	switch statusCode {
	// TODO: Wrap errors in a custom error type
	case http.StatusNotAcceptable:
		return appErrors.NewRemoteError(fmt.Errorf("query string was wrongly formatted: %s", url), statusCode, string(body))
	case http.StatusGatewayTimeout:
		return appErrors.NewRemoteError(fmt.Errorf("gateway timed out for the search API"), statusCode, "")
	case http.StatusTooManyRequests:
		return appErrors.NewRemoteError(fmt.Errorf("rate limit exceeded"), statusCode, "")
	case http.StatusUnauthorized:
		return appErrors.NewRemoteError(fmt.Errorf("unauthorized, the token being used could either have expired or is invalid"), statusCode, string(body))
	case http.StatusUnprocessableEntity:
		// https://docs.github.com/en/rest/search/search?apiVersion=2022-11-28#access-errors-or-missing-search-results
		return appErrors.NewRemoteError(fmt.Errorf("either the qualifer provided was invalid or the resource specified in the qualifier cant be accessed"), statusCode, string(body))
	default:
		return nil
	}

}

func (dataFetcher *RepositoryDataFetcher) bodyBytesToError(body []byte) (*HttpErrorResponse, error) {
	var errResponse HttpErrorResponse
	err := json.Unmarshal(body, &errResponse)
	if err != nil {
		dataFetcher.logger.Error("failed to unmarshal error response body: ", zap.Error(err))
		return nil, err
	}
	return &errResponse, nil
}
