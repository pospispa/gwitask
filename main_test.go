package main

import (
	"fmt"
	"math"
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

func TestBigRequest(t *testing.T) {
	binaryTreeHeight := 15
	bigRequest := constructBinaryTreeJSON(binaryTreeHeight)
	expectedOutput := constructExpectedJSONOutputOfBinaryTree(binaryTreeHeight)
	req, err := http.NewRequest("POST", "/challenge", strings.NewReader(bigRequest))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	http.HandlerFunc(challengeHandler).ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}
	assert.JSONEq(t, expectedOutput, rr.Body.String(), "Response body differs")
}

func TestConstructBinaryTreeJSON(t *testing.T) {
	tests := []struct {
		height     int
		wantedJSON string
	}{
		{height: 0, wantedJSON: `{"data": {}}`},
		{height: 1, wantedJSON: `{"data": {"facet00001": {"count": 1}}}`},
		{height: 2, wantedJSON: `{"data": {"facet00001": {"facet00002": {"count": 1}, "facet00003": {"count": 1}}}}`},
		{height: 3, wantedJSON: `{"data": {"facet00001": {"facet00002": {"facet00004": {"count": 1}, "facet00005": {"count": 1}}, "facet00003": {"facet00006": {"count": 1}, "facet00007": {"count": 1}}}}}`},
	}
	for _, tt := range tests {
		if got := constructBinaryTreeJSON(tt.height); got != tt.wantedJSON {
			t.Errorf("constructBinaryTree(%v) = (%s) WANT (%s)", tt.height, got, tt.wantedJSON)
		}
	}
}

func TestConstructExpectedJSONOutputOfBinaryTree(t *testing.T) {
	tests := []struct {
		height     int
		wantedJSON string
	}{
		{height: 0, wantedJSON: `{"result": []}`},
		{height: 1, wantedJSON: `{"result": [{"facet00001": 1}]}`},
		{height: 2, wantedJSON: `{"result": [{"facet00001": 2}, {"facet00002": 1}, {"facet00003": 1}]}`},
		{height: 3, wantedJSON: `{"result": [{"facet00001": 4}, {"facet00002": 2}, {"facet00003": 2}, {"facet00004": 1}, {"facet00005": 1}, {"facet00006": 1}, {"facet00007": 1}]}`},
	}
	for _, tt := range tests {
		if got := constructExpectedJSONOutputOfBinaryTree(tt.height); got != tt.wantedJSON {
			t.Errorf("constructBinaryTree(%v) = (%s) WANT (%s)", tt.height, got, tt.wantedJSON)
		}
	}
}

func constructBinaryTreeJSON(height int) string {
	tree := constructBinaryTree(height)
	return `{"data": {` + binaryTreeToJSON(tree) + `}}`
}

func binaryTreeToJSON(tree *tree) string {
	if tree == nil {
		return ""
	}
	if tree.left == nil && tree.right == nil {
		return fmt.Sprintf("\"facet%05d\": {\"count\": %d}", tree.facetNumber, tree.facetLeafSum)
	}
	return fmt.Sprintf("\"facet%05d\": {%s, %s}", tree.facetNumber, binaryTreeToJSON(tree.left), binaryTreeToJSON(tree.right))
}

type tree struct {
	facetNumber  int
	facetLeafSum int
	left, right  *tree
}

// constructBinaryTree will create a binary tree that will look like (the node number is facetNumber):
// 1                       all nodes on this level will have facetLeafSum == 8 (2^sub-tree height - 1)
// 2 3                     all nodes on this level will have facetLeafSum == 4 (2^sub-tree height - 1)
// 4 5 6 7                 all nodes on this level will have facetLeafSum == 2 (2^sub-tree height - 1)
// 8 9 10 11 12 13 14 15   all leaves will have facetLeafSum == 1 (2^sub-tree height - 1)
func constructBinaryTree(height int) *tree {
	if height == 0 {
		return nil
	}
	return constructBinaryTreeNode(1, height)
}

func constructBinaryTreeNode(nodeNumber, subTreeHeight int) *tree {
	leafSum := int(math.Pow(2, float64(subTreeHeight-1)))
	if subTreeHeight == 1 {
		return &tree{facetNumber: nodeNumber, facetLeafSum: leafSum, left: nil, right: nil}
	}
	return &tree{
		facetNumber:  nodeNumber,
		facetLeafSum: leafSum,
		left:         constructBinaryTreeNode(2*nodeNumber, subTreeHeight-1),
		right:        constructBinaryTreeNode(2*nodeNumber+1, subTreeHeight-1),
	}
}

const (
	lowestRow = 1
)

func constructExpectedJSONOutputOfBinaryTree(height int) string {
	var facets strings.Builder
	facets.WriteString(`{"result": [`)
	var firstInRow, lastInRow, facetValue int
	for i := height; i > 0; i-- {
		depth := height - i
		firstInRow = int(math.Pow(2, float64(depth)))
		lastInRow = int(math.Pow(2, float64(depth+1))) - 1
		facetValue = int(math.Pow(2, float64(i-1)))
		isLowestRow := i == lowestRow
		facets.WriteString(constructExpectedJSONOutputOfBinaryTreeRow(firstInRow, lastInRow, facetValue, isLowestRow))
	}
	facets.WriteString(`]}`)
	return facets.String()
}

func constructExpectedJSONOutputOfBinaryTreeRow(firstInRow, lastInRow, facetValue int, isLowestRow bool) string {
	var rowOfFacets strings.Builder
	for j := firstInRow; j <= lastInRow; j++ {
		if isLowestRow && j == lastInRow {
			rowOfFacets.WriteString(fmt.Sprintf("{\"facet%05d\": %d}", j, facetValue))
		} else {
			rowOfFacets.WriteString(fmt.Sprintf("{\"facet%05d\": %d}, ", j, facetValue))
		}
	}
	return rowOfFacets.String()
}
