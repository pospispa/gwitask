package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"sort"
)

var (
	addr = flag.String("addr", ":8080", "http service address")
)

const (
	requestRootElementName = "data"
	countElementName       = "count"
	maxContentLength       = 1000000
)

func main() {
	flag.Parse()
	http.HandleFunc("/challenge", challengeHandler)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

// Sums represents the resulting structure of sums of facet values
type Sums struct {
	Result []map[string]int `json:"result"`
}

func challengeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	if r.ContentLength == -1 {
		http.Error(w, "Error: content length required", http.StatusLengthRequired)
	}
	if r.ContentLength > maxContentLength {
		http.Error(w, "Error: content too large", http.StatusRequestEntityTooLarge)
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	response, httpStatusCode, err := sumFacetValues(body)
	if err != nil {
		http.Error(w, err.Error(), httpStatusCode)
		return
	}
	w.Write(response)
}

func sumFacetValues(requestBody []byte) ([]byte, int, error) {
	var unmarshaledRequestBody, unmarshaledFacetsBody interface{}
	var unmarshaledRequestBodyMap, unmarshaledFacetsBodyMap map[string]interface{}
	var ok bool
	if err := json.Unmarshal(requestBody, &unmarshaledRequestBody); err != nil {
		return []byte{}, http.StatusBadRequest, err
	}

	unmarshaledRequestBodyMap = unmarshaledRequestBody.(map[string]interface{})
	if len(unmarshaledRequestBodyMap) == 0 {
		return resultToJSON(map[string]int{})
	}
	if unmarshaledFacetsBody, ok = unmarshaledRequestBodyMap[requestRootElementName]; !ok {
		return []byte{}, http.StatusBadRequest, fmt.Errorf("Missing data field in request")
	}
	unmarshaledFacetsBodyMap = unmarshaledFacetsBody.(map[string]interface{})
	if len(unmarshaledFacetsBodyMap) == 0 {
		return resultToJSON(map[string]int{})
	}

	facetSums := map[string]int{}
	var err error
	var facetValue int
	for facetName, treeOrValue := range unmarshaledFacetsBodyMap {
		if facetSums, facetValue, err = sumSingleFacetTree(facetName, treeOrValue, facetSums); err != nil {
			return []byte{}, http.StatusBadRequest, err
		}
		if facetSums, err = addFacetIntoSums(facetName, facetValue, facetSums); err != nil {
			return []byte{}, http.StatusBadRequest, err
		}
	}
	return resultToJSON(facetSums)
}

func sumSingleFacetTree(facetName string, tree interface{}, preSummedValues map[string]int) (map[string]int, int, error) {
	treeMap := tree.(map[string]interface{})
	if len(treeMap) == 0 {
		return map[string]int{}, 0, fmt.Errorf("Error: %s does not have count leaf", facetName)
	}
	facetValue, ok := treeMap[countElementName]
	if ok && len(treeMap) > 1 {
		return map[string]int{}, 0, fmt.Errorf("Incorrect format of %s, more tnat 1 item in the count leaf", facetName)
	}
	if ok {
		switch facetValueType := facetValue.(type) {
		case int:
			return preSummedValues, facetValue.(int), nil
		case float64:
			return preSummedValues, int(math.Round(facetValue.(float64))), nil
		default:
			fmt.Printf("facetValue.(type): (%T)", facetValueType)
			return map[string]int{}, 0, fmt.Errorf("Incorrect value of %s, count value is not an integer", facetName)
		}
	}

	facetSums := preSummedValues
	facetCount := 0
	var err error
	var currFacetValue int
	for currFacetName, currFacetTreeOrValue := range treeMap {
		if facetSums, currFacetValue, err = sumSingleFacetTree(currFacetName, currFacetTreeOrValue, facetSums); err != nil {
			return map[string]int{}, 0, err
		}
		if facetSums, err = addFacetIntoSums(currFacetName, currFacetValue, facetSums); err != nil {
			return map[string]int{}, 0, err
		}
		facetCount += currFacetValue
	}
	return facetSums, facetCount, nil
}

func addFacetIntoSums(facetName string, value int, sums map[string]int) (map[string]int, error) {
	if _, ok := sums[facetName]; ok {
		return map[string]int{}, fmt.Errorf("Error duplicate name %s in request", facetName)
	}
	sums[facetName] = value
	return sums, nil
}

func resultToJSON(results map[string]int) ([]byte, int, error) {
	sums := Sums{}
	sums.Result = make([]map[string]int, len(results))
	i := 0
	for k, v := range results {
		sums.Result[i] = map[string]int{k: v}
		i++
	}
	sort.Slice(sums.Result, func(i, j int) bool { return getFacetName(sums.Result[i]) < getFacetName(sums.Result[j]) })
	result, err := json.Marshal(sums)
	if err != nil {
		return []byte{}, http.StatusInternalServerError, err
	}
	return result, http.StatusOK, nil
}

func getFacetName(facetMap map[string]int) string {
	var facetName string
	for facetName = range facetMap {
		return facetName
	}
	return facetName
}
