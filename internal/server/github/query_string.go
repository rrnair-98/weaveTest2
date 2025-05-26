package github

import (
	"fmt"
	"net/url"
	"strings"
)

// allowedQualifiers defines a list of supported query qualifiers for filtering search results in specific contexts.
var allowedQualifiers = []string{
	"in:file,path",
	"in:path,file",
	"in:file",
	"in:path",
	"user",
	"org",
	"repo",
	"path",
	"path",
	"language",
	"size",
	"filename",
	"extension",
}

var allowedOperators = []string{
	"AND",
	"OR",
	"NOT",
}

var replacementArray = make([]string, (len(allowedQualifiers)+len(allowedOperators))*2)

func init() {
	initReplacementArray()
}

func initReplacementArray() {
	j := 0
	for i, qualifier := range allowedQualifiers {
		replacementArray[i*2] = qualifier + ":"
		replacementArray[i*2+1] = ""

	}
	opIter := 0
	for i := j; i < j+3; i++ {
		replacementArray[i*2] = allowedOperators[opIter]
		replacementArray[i*2+1] = ""
		opIter++
	}
}

// allowedKeyWords represents a list of query params that this API supports,
// the assumption is that only the qualifier string will be passed to the GRPC server
var allowedKeywords = []string{
	"q",
	"sort",
	"order",
	"per_page",
	"page",
}

const (
	gitUrlFmt           = "https://api.github.com/search/code?q=%s"
	userQueryStringFmt  = "https://api.github.com/search/code?q=%s+%s"
	userQueryDefaultFmt = "user:%s"
)

// queryString is a wrapper around a string that represents a query string for github code search API.
// The query string is a string that can be appended to the gitUrlFmt to form a valid url.
// Content in q=() has to be URI encoded always.
// content after q= cannot exceed 255 chars(in my testing, the API ignores this), without qualifiers and operators being counted
// allowed keywords for code search are defined in allowedKeywords
// allowed qualifiers are defined in allowedQualifiers
// allowed operators are defined in allowedOperators.
// One thing to remember is that the API performs a case-insensitive search.
// Q, sort(default=indexed), order=(asc/desc| default=desc), per_page, page,
// q is required and can not be empty. The others are optional and have default values.
type queryString string

func (q queryString) Validate() error {
	if err := q.IsEmpty(); err != nil {
		return err
	}
	if err := q.hasValidNumAndsOrsNots(); err != nil {
		return err
	}
	if err := q.satisfiesQueryLenCapacity(); err != nil {
		return err
	}
	return nil
}

func (q queryString) IsEmpty() error {
	if q == "" {
		return fmt.Errorf("query string cant be empty")
	}
	return nil
}

func (q queryString) ToUrl() (string, error) {
	if err := q.Validate(); err != nil {
		return "", err
	}
	return fmt.Sprintf(gitUrlFmt, url.QueryEscape(string(q))), nil
}

func (q queryString) ToUrlWithUser(user string) (string, error) {
	if err := q.Validate(); err != nil {
		return "", err
	}
	if user == "" {
		return fmt.Sprintf(gitUrlFmt, url.QueryEscape(string(q))), nil
	}
	userStr := fmt.Sprintf(userQueryDefaultFmt, user)
	return fmt.Sprintf(userQueryStringFmt, url.QueryEscape(string(q)), url.QueryEscape(userStr)), nil
}

// hasValidNumAndsOrsNots checks if the query string contains more than 5 boolean operators (AND, OR, NOT).
// Mentioned here https://docs.github.com/en/rest/search/search?apiVersion=2022-11-28#limitations-on-query-length
func (q queryString) hasValidNumAndsOrsNots() error {
	queryStr := strings.ToUpper(string(q))

	// Count occurrences of each operator
	andCount := strings.Count(queryStr, " AND ")
	orCount := strings.Count(queryStr, " OR ")
	notCount := strings.Count(queryStr, " NOT ")
	totalOperators := andCount + orCount + notCount
	fmt.Printf("totalOperators: %d\n", totalOperators)
	if totalOperators > 5 {
		return fmt.Errorf("query string contains %d boolean operators (AND, OR, NOT), exceeding the maximum of 5", totalOperators)
	}
	return nil
}

func (q queryString) satisfiesQueryLenCapacity() error {
	// The API ignores this, but it is a good practice to check for it.
	strippedQuery := strings.NewReplacer(replacementArray...).Replace(string(q))
	if len(strippedQuery) > 255 {
		return fmt.Errorf("query string exceeds the maximum of 255 characters")
	}
	return nil
}
