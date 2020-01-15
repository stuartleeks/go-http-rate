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

	notifyKeyDone := make(chan string, parallel)     // used to notify on time key changes
	notifyKeyStarting := make(chan string, parallel) // used to notify on time key changes
	doneChan := make(chan bool, parallel)

	statsArray := make([]map[string]Stats, parallel)
	for loopIndex := 0; loopIndex < parallel; loopIndex++ {
		stats := map[string]Stats{}
		statsArray[loopIndex] = stats
		lastKey := ""
		go func() {
			for index := 0; index < count; index++ {
				t := time.Now().UTC()
				key := t.Format("2006-01-02T15:04:05")
				if lastKey != key {
					notifyKeyStarting <- key
				}
				status, err := makeRequest(url, bearerToken)
				if err != nil {
					panic(err)
				}
				if stats[key] == nil {
					stats[key] = Stats{}
				}
				stats[key][status]++
				if lastKey != key {
					if lastKey != "" {
						// notify main func that we're done with that key
						notifyKeyDone <- lastKey
					}
					lastKey = key
				}
			}
			notifyKeyDone <- lastKey
			doneChan <- true
		}()
	}

	notifyCounts := map[string]int{}

	// Wait for loops to finish
	for remainingLoops := parallel; remainingLoops > 0; {
		select {
		case key := <-notifyKeyStarting:
			notifyCounts[key]++
		case key := <-notifyKeyDone:
			notifyCounts[key]--
			if notifyCounts[key] == 0 {
				// time period done with by all goroutines
				// aggregate and output
				aggregatedStats := Stats{}
				for i := 0; i < parallel; i++ {
					timeStats := statsArray[i][key]
					for k, v := range timeStats {
						aggregatedStats[k] += v
					}
				}
				for k, v := range aggregatedStats {
					fmt.Printf("%s|%d %d\n", key, k, v)
				}
			}
		case <-doneChan:
			remainingLoops--
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
