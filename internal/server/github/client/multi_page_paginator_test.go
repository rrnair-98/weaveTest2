package client

import (
	"context"
	"google.golang.org/grpc/codes"
	"net/http"
	"path/filepath"
	"runtime"
	"testing"
	"weaveTest/internal/config"
	"weaveTest/internal/server/github/client/fixtures"

	"go.uber.org/zap"
	"weaveTest/internal/proto/generated"
	appError "weaveTest/internal/server/github/client/errors"
)

// MockRequestMaker implements the RequestMaker interface for testing
type MockRequestMaker struct {
}

func (m *MockRequestMaker) Perform(ctx context.Context, url string) (*http.Response, appError.AppError) {
	return fixtures.SampleResponses[url], nil
	// Default error if no more responses
}

func getProjectRoot() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("Could not get caller information")
	}
	// Assuming the test file is in a subdirectory of the project root
	return filepath.Dir(filepath.Dir(filename))
}

func TestMultiPagePaginator_Paginate_Success(t *testing.T) {
	// Setup test environment
	logger, _ := zap.NewDevelopment()
	t.Log("project root: ", getProjectRoot())
	var fp = filepath.Join(getProjectRoot(), "..", "..", "..", ".env.test.json")
	t.Log("env file path: ", fp)
	err := config.InitEnvFromFile(fp)
	if err != nil {

		t.Fatal(err)
	}
	// Create logger

	// Create paginator
	paginator := &MultiPagePaginator{
		logger:       logger,
		requestMaker: &MockRequestMaker{},
	}

	// Create search request with correct properties
	request := &generated.SearchRequest{
		SearchTerm: "badWithHttp422",
		User:       "testuser",
	}
	tests := []struct {
		name       string
		query      queryString
		pageNumber int
		perPage    int
		grpcStatus codes.Code
		wantErr    bool
	}{
		{
			name:       "success",
			query:      "badWithHttp422",
			pageNumber: 1,
			perPage:    30,
			grpcStatus: codes.InvalidArgument,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := paginator.Paginate(context.Background(), request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Paginate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.grpcStatus != err.GrpcStatus() {
				t.Errorf("Paginate() error = %v, wantErr %v", err, tt.grpcStatus)
			}
		})
	}

}
