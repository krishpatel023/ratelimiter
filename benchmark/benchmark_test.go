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
	concurrency = 3
	totalUsers  = 600 // Increased user diversity
	headerName  = "X-ID"
	duration    = 60 * time.Second // Run the benchmark for x seconds
	iterations  = 2                // Number of times to run the benchmark
)

func TestBenchmark(t *testing.T) {
	successful, failed, bounced, total, avgRespTime := int32(0), int32(0), int32(0), int32(0), 0*time.Second
	for range iterations {

		s, f, b, t, a := RunBenchmark(url, concurrency, totalUsers, duration)

		successful += s
		failed += f
		bounced += b
		total += t
		avgRespTime += a
	}

	PrintStats(successful, failed, bounced, total, avgRespTime, duration, iterations)
}

func pickUser() string {
	return fmt.Sprintf("user-%d", rand.Intn(totalUsers)) // Random user from 1000
}

func sendRequest(client *http.Client, userID string, url string) (int, bool, time.Duration) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Request creation failed:", err)
		return 404, false, 0
	}
	req.Header.Set(headerName, userID)

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start)

	if err != nil {
		// Less verbose error logging
		fmt.Printf("Request failed: %v\n", err)
		return 404, false, elapsed
	}
	defer resp.Body.Close()

	return resp.StatusCode, true, elapsed
}

func PrintStats(successfulReqs, failedReqs, bouncedReqs, totalReqs int32, avgResponseTime time.Duration, duration time.Duration, total int) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, "+--------------------------+-------+")
	fmt.Fprintln(w, "| Stats                    | Count |")
	fmt.Fprintln(w, "+--------------------------+-------+")
	fmt.Fprintf(w, "| Successful Requests      | %5d |\n", successfulReqs)
	fmt.Fprintf(w, "| Rate Limited Requests    | %5d |\n", failedReqs)
	fmt.Fprintf(w, "| Bounced Requests         | %5d |\n", bouncedReqs)
	fmt.Fprintln(w, "+--------------------------+-------+")
	fmt.Fprintf(w, "| Total Requests           | %5d |\n", totalReqs)
	fmt.Fprintln(w, "+--------------------------+-------+")
	fmt.Fprintf(w, "| Avg Response Time (ms)   | %5.2f |\n", avgResponseTime.Seconds()*1000)
	fmt.Fprintf(w, "| Duration / benchmark (s) | %5.2f |\n", duration.Seconds())
	fmt.Fprintf(w, "| Total Benchmark Runs     | %5d |\n", total)
	fmt.Fprintln(w, "+--------------------------+-------+")

	w.Flush()
}

func RunBenchmark(url string, concurrency, totalUsers int, duration time.Duration) (int32, int32, int32, int32, time.Duration) {
	// Create a transport with proper connection pooling
	transport := &http.Transport{
		MaxIdleConns:        concurrency * 2,
		MaxIdleConnsPerHost: concurrency * 2,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false, // This is good, keep it false
		ForceAttemptHTTP2:   true,
	}

	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrency)

	var successfulReqs, failedReqs, bouncedReqs, totalReqs int32
	var totalResponseTime int64
	var responseTimeCount int32

	start := time.Now()
	for time.Since(start) < duration {
		wg.Add(1)
		semaphore <- struct{}{} // This blocks properly if concurrency limit is reached

		go func() {
			defer func() {
				<-semaphore // Free up a slot in the semaphore
				wg.Done()
			}()

			// Small random jitter to avoid connection collisions
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(10))) // Small random delay

			resp, success, elapsed := sendRequest(client, pickUser(), url)
			atomic.AddInt32(&totalReqs, 1)

			if success {
				atomic.AddInt32(&responseTimeCount, 1)
				atomic.AddInt64(&totalResponseTime, int64(elapsed))

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
		}()
	}

	wg.Wait()

	avgResponseTime := time.Duration(0)
	if responseTimeCount > 0 {
		avgResponseTime = time.Duration(atomic.LoadInt64(&totalResponseTime)) / time.Duration(responseTimeCount)
	}

	return successfulReqs, failedReqs, bouncedReqs, totalReqs, avgResponseTime
}
