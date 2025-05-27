package fixtures

import "net/http"

var SampleResponses = make(map[string]*http.Response)
func initSampleResponses() {
	SampleResponses["https://api.github.com/search/code?q=badWithHttp422+user:testuser&per_page=30&page=1"] = &http.Response{
		Status:           "unprocessable entity",
		StatusCode:       http.StatusUnprocessableEntity,
		Proto:            "",
		ProtoMajor:       0,
		ProtoMinor:       0,
		Header:           nil,
		Body:             nil,
		ContentLength:    0,
		TransferEncoding: nil,
		Close:            false,
		Uncompressed:     false,
		Trailer:          nil,
		Request:          nil,
		TLS:              nil,
	}
}

