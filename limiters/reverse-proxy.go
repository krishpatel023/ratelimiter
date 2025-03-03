package limiters

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/krishpatel023/ratelimiter/internal/helper"
)

// ReverseProxyConfig holds configuration for the reverse proxy
// It is used to set up the reverse proxy by the rate limiter middleware
func reverseProxy(TargetURL string) http.Handler {
	// Define the backend server URL
	targetURL, err := url.Parse(TargetURL)
	if err != nil || targetURL.String() == "" {
		helper.Log(fmt.Sprintf("Failed to parse target URL: %v", "Please add/check the target URL"), "error")
		os.Exit(1)
	}

	// Create a reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Set up the handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})

	return handler
}
