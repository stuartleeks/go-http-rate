package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Stats map[int]int // map of

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

	// notifyChan := make(chan string, parallel) // used to notify on time key changes
	doneChan := make(chan bool, parallel)

	statsArray := make([]map[string]Stats, parallel)
	for loopIndex := 0; loopIndex < parallel; loopIndex++ {
		stats := map[string]Stats{}
		statsArray[loopIndex] = stats
		go func() {
			for index := 0; index < count; index++ {
				status, err := makeRequest(url, bearerToken)
				if err != nil {
					panic(err)
				}
				t := time.Now().UTC()
				key := t.Format("2006-01-02T15:04:05")
				if stats[key] == nil {
					stats[key] = Stats{}
				}
				stats[key][status]++
			}
			doneChan <- true
		}()
	}

	// Wait for loops to finish
	for remainingLoops := parallel; remainingLoops > 0; remainingLoops-- {
		<-doneChan
	}

	aggregatedStats := map[string]Stats{}
	for index := 0; index < parallel; index++ {
		stats := statsArray[index]
		for k := range stats {
			for k2 := range stats[k] {
				if aggregatedStats[k] == nil {
					aggregatedStats[k] = Stats{}
				}
				aggregatedStats[k][k2] += stats[k][k2]
			}
		}
	}

	for k, v := range aggregatedStats {
		for k2, v2 := range v {
			fmt.Printf("%s|%d %d\n", k, k2, v2)
		}
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
