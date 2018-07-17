package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArticlesHandler(t *testing.T) {
	tests := []struct {
		testBody       string
		expectedOutput string
	}{
		{
			testBody: `{"data": {"facet1": {"facet3": {"facet4": {"facet6": {"count": 20},"facet7": {"count": 30}},"facet5": {"count": 50}}}, "facet2": {"count": 0}}}`,
			expectedOutput: `{ "result": [
                {"facet1": 100},
                {"facet2": 0},
                {"facet3": 100},
                {"facet4": 50},
                {"facet5": 50},
                {"facet6": 20},
                {"facet7": 30}
                ]
            }`,
		},
		{
			testBody:       `{}`,
			expectedOutput: `{ "result": []}`,
		},
		{
			testBody:       `{"data": {}}`,
			expectedOutput: `{ "result": []}`,
		},
		{
			testBody: `{"data": {"facet1": {"count": 100}}}`,
			expectedOutput: `{ "result": [
				{"facet1": 100}
				]}`,
		},
		{
			testBody: `{"data": {"facet4": {"facet6": {"count": 20},"facet7": {"count": 30}}}}`,
			expectedOutput: `{ "result": [
				{"facet4": 50},
				{"facet6": 20},
				{"facet7": 30}
				]}`,
		},
	}

	for _, tt := range tests {
		req, err := http.NewRequest("POST", "/challenge", strings.NewReader(tt.testBody))
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		http.HandlerFunc(challengeHandler).ServeHTTP(rr, req)
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
		}
		assert.JSONEq(t, tt.expectedOutput, rr.Body.String(), "Response body differs")
	}
}
