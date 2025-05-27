package client

//
//import (
//	"fmt"
//	"go.uber.org/zap"
//	"io"
//	"net/http"
//	"weaveTest/internal/proto/generated"
//	appErrors "weaveTest/internal/server/github/errors"
//)
//
//type RemoteClient struct {
//	url string
//	token string
//	logger zap.Logger
//}
//
//func NewRemoteClient(url string, token string) *RemoteClient {
//	return &RemoteClient{
//		url: url,
//		token: token,
//	}
//}
//
//func (r *RemoteClient) Fetch(url string) (http.Response, error) {
//	client := &http.Client{}
//	req, err := http.NewRequest(httpMethod, url, nil)
//	if err != nil {
//		dataFetcher.logger.Error("failed to create httpClient", zap.Error(err))
//		return nil, appErrors.NewInternalError(err, appErrors.InvalidHttpClient, "failed to create httpClient")
//	}
//	// dont need to set this accept header, since text_matches is not being used right now
//	// req.Header.Add("Accept", "application/vnd.github+json")
//	req.Header.Add("Authorization", fmt.Sprintf(bearerFmt, dataFetcher.accessToken))
//	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")
//
//	dataFetcher.logger.Debug("performing http request for url: ", zap.String("url", url))
//	res, err := client.Do(req)
//	if err != nil {
//		dataFetcher.logger.Error("failed to perform http request: ", zap.Error(err))
//		return nil, appErrors.NewInternalError(err, appErrors.InvalidHttpClient, "failed to perform http request")
//	}
//	defer res.Body.Close()
//
//	body, err := io.ReadAll(res.Body)
//	if err != nil {
//		dataFetcher.logger.Error("failed to read response body: ", zap.Error(err))
//		return nil, appErrors.NewInternalError(err, appErrors.InvalidJSONBody, "failed to read response body")
//	}
//
//	if dataFetcher.paginated {
//		// TODO: implement pagination
//		return nil, nil
//	}
//
//	if err := dataFetcher.handleHttpErrors(res.StatusCode, body, url); err != nil {
//		return nil, err
//	}
//	dataFetcher.logger.Debug("successfully fetched data from remote")
//	var response CodeSearchResponse
//	err = json.Unmarshal(body, &response)
//	if err != nil {
//		dataFetcher.logger.Error("failed to unmarshal response body: ", zap.Error(err))
//		return nil, appErrors.NewInternalError(err, appErrors.InvalidJSONBody, string(body))
//	}
//	var results []*generated.Result
//	for _, item := range response.RepositoryItems {
//		results = append(results, &generated.Result{
//			FileUrl: item.HTMLURL,
//			Repo:    item.Repo.HTMLURL,
//		})
//	}
//	dataFetcher.logger.Debug("successfully fetched data from remote", zap.Int("numResults", len(results)))
//	return &generated.SearchResponse{Results: results}, nil
//}
