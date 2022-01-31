package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"

	es "github.com/elastic/go-elasticsearch/v7"
)

type ElasticClient struct {
	esClient *es.Client
}

func New() (*ElasticClient, error) {
	es, err := es.NewDefaultClient()
	if err != nil {
		return nil, err
	}
	ec := &ElasticClient{
		esClient: es,
	}
	return ec, nil
}
func getFields(searchType string) []string {
	switch searchType {
	case "address":
		return []string{"city", "state", "country"}
	case "name":
		return []string{"fname", "middleName", "lname"}
	}
	return nil
}
func getMatchType(searchBy string) string {
	switch searchBy {
	case "term":
		return "multi_match"
	case "infix":
		return "match"
	}
	return ""
}
func getSearchType(searchType string) string {
	switch searchType {
	case "address":
		return "full_address"
	case "name":
		return "full_name"
	}
	return ""
}
func createMatchQuery(matchType, text, searchType string) map[string]interface{} {

	switch matchType {
	case "multi_match":
		return map[string]interface{}{
			"query": text,
			"type":  "bool_prefix",
			"fields": []string{
				fmt.Sprintf("%s_term", getSearchType(searchType)),
				fmt.Sprintf("%s_term._2gram", getSearchType(searchType)),
				fmt.Sprintf("%s_term._3gram", getSearchType(searchType)),
			}}
	case "match":
		return map[string]interface{}{
			getSearchType(searchType): map[string]interface{}{
				"query":    text,
				"operator": "and",
			},
		}
	}
	return nil
}
func getFullQuery(searchBy, searchType, text string) map[string]interface{} {
	switch searchBy {
	case "prefix":
		return map[string]interface{}{
			"prefix": map[string]interface{}{
				getSearchType(searchType) + "_prefix": map[string]interface{}{
					"value":            text,
					"case_insensitive": true,
				},
			},
		}
	case "infix", "term":
		return map[string]interface{}{
			getMatchType(searchBy): createMatchQuery(getMatchType(searchBy), text, searchType),
		}
	}
	return nil
}
func (esp *ElasticClient) Search(index, text, searchBy, searchType string, fields []string) []map[string]interface{} {
	var combine []map[string]interface{}
	var result map[string]interface{}
	// Build the request body.
	fields = getFields(searchType)
	var buf bytes.Buffer
	query := map[string]interface{}{
		"size":  5000,
		"query": getFullQuery(searchBy, searchType, text),
	}
	fmt.Println("Query:: ", query)

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}

	// Perform the search request.
	res, err := esp.esClient.Search(
		esp.esClient.Search.WithContext(context.Background()),
		esp.esClient.Search.WithIndex(index),
		esp.esClient.Search.WithBody(&buf),
		esp.esClient.Search.WithTrackTotalHits(true),
		esp.esClient.Search.WithPretty(),
	)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
		} else {
			// Print the response status and error information.
			log.Fatalf("[%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
	}
	// Print the response status, number of results, and request duration.
	log.Printf(
		"[%s] %d hits; took: %dms",
		res.Status(),
		int(result["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
		int(result["took"].(float64)),
	)
	// Print the ID and document source for each hit.
	for _, hit := range result["hits"].(map[string]interface{})["hits"].([]interface{}) {
		temp := map[string]interface{}{}
		var outputStr string
		for i, field := range fields {
			if i < len(fields)-1 {
				outputStr += fmt.Sprintf("%v,", hit.(map[string]interface{})["_source"].(map[string]interface{})[field])
			} else {
				outputStr += fmt.Sprintf("%v", hit.(map[string]interface{})["_source"].(map[string]interface{})[field])
			}
		}
		temp["output"] = outputStr
		combine = append(combine, temp)
	}
	// fmt.Println(combine)
	return combine
}
