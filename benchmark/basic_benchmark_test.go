package benchmark

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"text/tabwriter"
	"time"
)

const (
	url         = "http://localhost:8080/"
	numReqs     = 30
	concurrency = 1
	totalUsers  = 2
)

func TestBenchmark(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	RunBenchmark(
		url,
		numReqs,
		concurrency,
		totalUsers,
	)
}

func pickUser(limit int, prefix string) string {
	pickedNumber := rand.Intn(limit) + 1 // Random number between 1 and limit
	return fmt.Sprintf("%s-%d", prefix, pickedNumber)
}

func sendRequest(client *http.Client, userID string, url string) (int, bool) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Request creation failed:", err)
		return 404, false
	}
	req.Header.Set("X-ID", userID)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Request failed:", err)
		return 404, false
	}
	defer resp.Body.Close()

	return resp.StatusCode, true
}

func PrintStats(successfulReqs, failedReqs, bouncedReqs, totalReqs int32) {
	// Create a new tabwriter that writes to stdout
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Print the table header
	fmt.Fprintln(w, "+-------------------------+-------+")
	fmt.Fprintln(w, "| Stats                   | Count |")
	fmt.Fprintln(w, "+-------------------------+-------+")

	// Print the data rows
	fmt.Fprintf(w, "| Successful Requests     | %5d |\n", successfulReqs)
	fmt.Fprintf(w, "| Rate Limited Requests   | %5d |\n", failedReqs)
	fmt.Fprintf(w, "| Bounced Requests        | %5d |\n", bouncedReqs)
	fmt.Fprintln(w, "+-------------------------+-------+")
	fmt.Fprintf(w, "| Total Requests          | %5d |\n", totalReqs)
	fmt.Fprintln(w, "+-------------------------+-------+")

	// Flush the tabwriter to ensure everything is written
	w.Flush()
}

func RunBenchmark(url string, numReqs, concurrency, totalUsers int) {
	client := &http.Client{Timeout: 5 * time.Second}
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrency) // Limit concurrency

	var successfulReqs, failedReqs, bouncedReqs, totalReqs int32

	for i := 0; i < numReqs; i++ {
		wg.Add(1)
		semaphore <- struct{}{} // Block if concurrency is reached

		go func(id int) {
			defer func() {
				<-semaphore // Release slot
				wg.Done()
			}()

			resp, success := sendRequest(client, pickUser(totalUsers, "user"), url)
			if success {
				if resp == 200 {
					atomic.AddInt32(&successfulReqs, 1)
				} else if resp == 429 {
					atomic.AddInt32(&failedReqs, 1)
				} else {
					atomic.AddInt32(&bouncedReqs, 1)
				}
			} else {
				atomic.AddInt32(&bouncedReqs, 1)
			}
			atomic.AddInt32(&totalReqs, 1)
		}(i)
	}

	wg.Wait()

	PrintStats(successfulReqs, failedReqs, bouncedReqs, totalReqs)
}
