package client

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"weaveTest/internal/config"
	appError "weaveTest/internal/server/github/client/errors"
)

type RequestMaker interface {
	Perform(ctx context.Context, url string) (*http.Response, appError.AppError)
}

type DefaultRequestMaker struct {
	logger *zap.Logger
}

func (r *DefaultRequestMaker) Perform(ctx context.Context, url string) (*http.Response, appError.AppError) {
	return performRequest(ctx, url, r.logger)
}

func NewDefaultRequestMaker(logger *zap.Logger) *DefaultRequestMaker {
	return &DefaultRequestMaker{
		logger: logger,
	}
}

func performRequest(ctx context.Context, url string, logger *zap.Logger) (*http.Response, appError.AppError) {
	env := config.GetEnv()
	client := &http.Client{}
	req, err := http.NewRequest(httpMethod, url, nil)
	if err != nil {
		logger.Error("failed to create httpClient", zap.Error(err))
		return nil, appError.NewInternalError(err, appError.InvalidHttpClient, "failed to create httpClient")
	}
	// dont need to set this accept header, since text_matches is not being used right now
	// req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("Authorization", fmt.Sprintf(bearerFmt, env.GitToken))
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	logger.Debug("performing http request for url: ", zap.String("url", url))
	if env.Paginator.RateLimited {
		rateLimiter := GetRateLimiter()
		logger.Debug("rate limiter wait started")
		err := rateLimiter.WaitWithContext(ctx)
		logger.Debug("rate limiter wait completed")
		if err != nil {
			logger.Error("rate limiter failed, limit exceeded", zap.Error(err))
			return nil, appError.NewRemoteError(err, http.StatusTooManyRequests, "rate limiter failed, limit exceeded")
		}
	}
	res, err := client.Do(req)
	if err != nil {
		logger.Error("failed to perform http request: ", zap.Error(err))
		return nil, appError.NewInternalError(err, appError.InvalidHttpClient, "failed to perform http request")
	}
	return res, nil
}
