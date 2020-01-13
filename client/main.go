package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	const DefaultURL = "http://localhost:8080/api"

	url, found := os.LookupEnv("CLIENT_URL")
	if !found {
		url = DefaultURL
	}

	bearerToken, _ := os.LookupEnv("CLIENT_BEARER_TOKEN")

	count := 10
	if countString, found := os.LookupEnv("CLIENT_COUNT"); found {
		var err error
		count, err = strconv.Atoi(countString)
		if err != nil {
			panic(err)
		}
	}

	parallel := 1
	if parallelString, found := os.LookupEnv("CLIENT_PARALLEL"); found {
		var err error
		parallel, err = strconv.Atoi(parallelString)
		if err != nil {
			panic(err)
		}
	}

	doneChan := make(chan bool, parallel)

	statsArray := make([]map[string]int, parallel)
	for loopIndex := 0; loopIndex < parallel; loopIndex++ {
		stats := map[string]int{}
		statsArray[loopIndex] = stats
		go func() {
			for index := 0; index < count; index++ {
				status, err := makeRequest(url, bearerToken)
				if err != nil {
					panic(err)
				}
				t := time.Now().UTC()
				key := fmt.Sprintf("%s|%d", t.Format("2006-01-02T15:04:05"), status)
				stats[key]++
			}
			doneChan <- true
		}()
	}

	// Wait for loops to finish
	for remainingLoops := parallel; remainingLoops > 0; remainingLoops-- {
		<- doneChan
	}

	aggregatedStats := map[string]int{}
	for index := 0; index < parallel; index++ {
		stats := statsArray[index]
		for k := range stats {
			aggregatedStats[k] += stats[k]
		}
	}

	for k, v := range aggregatedStats {
		fmt.Printf("%s %d\n", k, v)
	}
}
func makeRequest(url string, bearerToken string) (int, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	if bearerToken != "" {
		req.Header.Add("Authorization", "Bearer "+bearerToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()
	_, _ = ioutil.ReadAll(resp.Body)

	return resp.StatusCode, nil
}
