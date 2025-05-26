package github

import (
	"strings"
	"testing"
)

func TestQueryString_IsEmpty(t *testing.T) {
	tests := []struct {
		name    string
		query   queryString
		wantErr bool
	}{
		{
			name:    "Empty query string",
			query:   queryString(""),
			wantErr: true,
		},
		{
			name:    "Non-empty query string",
			query:   queryString("test"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.query.IsEmpty()
			if (err != nil) != tt.wantErr {
				t.Errorf("IsEmpty() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestQueryString_hasValidNumAndsOrsNots(t *testing.T) {
	tests := []struct {
		name    string
		query   queryString
		wantErr bool
	}{
		{
			name:    "No boolean operators",
			query:   queryString("test"),
			wantErr: false,
		},
		{
			name:    "Five boolean operators",
			query:   queryString("test AND test1 AND test2 OR test3 OR test4 NOT test5"),
			wantErr: false,
		},
		{
			name:    "Six boolean operators",
			query:   queryString("test AND test1 AND test2 OR test3 OR test4 NOT test5 AND test6"),
			wantErr: true,
		},
		{
			name:    "Mixed case operators",
			query:   queryString("test AnD test1 and test2 or test3 OR test4 not test5"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.query.hasValidNumAndsOrsNots()
			if (err != nil) != tt.wantErr {
				t.Errorf("hasValidNumAndsOrsNots() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestQueryString_satisfiesQueryLenCapacity(t *testing.T) {
	tests := []struct {
		name    string
		query   queryString
		wantErr bool
	}{
		{
			name:    "Short query",
			query:   queryString("test"),
			wantErr: false,
		},
		{
			name:    "Query with qualifiers but still under limit",
			query:   queryString("language:go repo:user/repo size:1000"),
			wantErr: false,
		},
		{
			name:    "Query exceeding limit after stripping qualifiers",
			query:   queryString("language:go AND language:ruby" + strings.Repeat("a", 256)),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.query.satisfiesQueryLenCapacity()
			if (err != nil) != tt.wantErr {
				t.Errorf("satisfiesQueryLenCapacity() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestQueryString_Validate(t *testing.T) {
	tests := []struct {
		name    string
		query   queryString
		wantErr bool
	}{
		{
			name:    "Valid query",
			query:   queryString("test"),
			wantErr: false,
		},
		{
			name:    "Empty query",
			query:   queryString(""),
			wantErr: true,
		},
		{
			name:    "Too many boolean operators",
			query:   queryString("test AND test1 AND test2 OR test3 OR test4 NOT test5 AND test6"),
			wantErr: true,
		},
		{
			name:    "Query too long",
			query:   queryString("language:go " + strings.Repeat("a", 256)),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.query.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestQueryString_ToUrl(t *testing.T) {
	tests := []struct {
		name    string
		query   queryString
		want    string
		wantErr bool
	}{
		{
			name:    "Valid query",
			query:   queryString("test"),
			want:    "https://api.github.com/search/code?q=test",
			wantErr: false,
		},
		{
			name:    "Query with spaces",
			query:   queryString("hello world"),
			want:    "https://api.github.com/search/code?q=hello+world",
			wantErr: false,
		},
		{
			name:    "Query with special characters",
			query:   queryString("test&query=value"),
			want:    "https://api.github.com/search/code?q=test%26query%3Dvalue",
			wantErr: false,
		},
		{
			name:    "Empty query",
			query:   queryString(""),
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.query.ToUrl()
			if (err != nil) != tt.wantErr {
				t.Errorf("ToUrl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ToUrl() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQueryString_ToUrlWithUser(t *testing.T) {
	tests := []struct {
		name    string
		query   queryString
		user    string
		want    string
		wantErr bool
	}{
		{
			name:    "Valid query with user",
			query:   queryString("test"),
			user:    "github-user",
			want:    "https://api.github.com/search/code?q=test+user%3Agithub-user",
			wantErr: false,
		},
		{
			name:    "Query with spaces and user",
			query:   queryString("hello world"),
			user:    "github-org",
			want:    "https://api.github.com/search/code?q=hello+world+user%3Agithub-org",
			wantErr: false,
		},
		{
			name:    "Query with special characters",
			query:   queryString("test&query"),
			user:    "octocat",
			want:    "https://api.github.com/search/code?q=test%26query+user%3Aoctocat",
			wantErr: false,
		},
		{
			name:    "Query with qualifiers",
			query:   queryString("language:go filename:main.go"),
			user:    "google",
			want:    "https://api.github.com/search/code?q=language%3Ago+filename%3Amain.go+user%3Agoogle",
			wantErr: false,
		},
		{
			name:    "Empty query",
			query:   queryString(""),
			user:    "github-user",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Empty user",
			query:   queryString("test"),
			user:    "",
			want:    "https://api.github.com/search/code?q=test",
			wantErr: false,
		},
		{
			name:    "User with special characters",
			query:   queryString("test"),
			user:    "org/repo",
			want:    "https://api.github.com/search/code?q=test+user%3Aorg%2Frepo",
			wantErr: false,
		},
		{
			// user:a+user:b works in the API, it treats it like OR user:a or user:b
			name:    "Query with existing user qualifier",
			query:   queryString("user:existinguser test"),
			user:    "newuser",
			want:    "https://api.github.com/search/code?q=user%3Aexistinguser+test+user%3Anewuser",
			wantErr: false,
		},
		{
			name:    "Query with boolean operators",
			query:   queryString("test AND debug OR log NOT error"),
			user:    "github",
			want:    "https://api.github.com/search/code?q=test+AND+debug+OR+log+NOT+error+user%3Agithub",
			wantErr: false,
		},
		{
			name:    "Query exceeding length limit",
			query:   queryString("language:go " + strings.Repeat("a", 256)),
			user:    "testuser",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Query with too many boolean operators",
			query:   queryString("a AND b AND c AND d AND e AND f AND g"),
			user:    "testuser",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.query.ToUrlWithUser(tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToUrlWithUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ToUrlWithUser() got = %v, want %v", got, tt.want)
			}
		})
	}
}
