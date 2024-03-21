package proxy

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// Server represents a proxy server configuration
type Server struct {
	backendURL  string
	proxyPort   string
	endpoint    string // the recommend endpoint
	rateLimiter *RateLimiter
}

// NewServer creates a new instance of ProxyServer with the given backend URL and proxy port
func NewServer(backendURL, proxyPort, endpoint string, limiter *RateLimiter) *Server {
	return &Server{
		backendURL:  backendURL,
		proxyPort:   proxyPort,
		endpoint:    endpoint,
		rateLimiter: limiter,
	}
}

func (srv *Server) copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// proxyHandler handles the incoming requests and proxies them to the backend server
func (srv *Server) proxyHandler(w http.ResponseWriter, req *http.Request) {
	// Extract userid from the URL path
	userID := req.URL.Path[len(srv.endpoint):]

	// Check if userid is an integer
	if _, err := strconv.Atoi(userID); err != nil {
		// If not an integer, return an error message
		http.Error(w, "invalid userid", http.StatusBadRequest)
		return
	}

	// check rate limiter
	if !srv.rateLimiter.CanProcess() {
		http.Error(w, "server is busy", http.StatusBadRequest)
		return
	}

	// Forward the request to the backend server
	backendURL := srv.backendURL + req.URL.Path
	proxyReq, err := http.NewRequest(req.Method, backendURL, req.Body)
	if err != nil {
		http.Error(w, "Error creating request to backend", http.StatusInternalServerError)
		return
	}

	srv.copyHeader(proxyReq.Header, req.Header)

	// Execute the request to the backend
	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, "Error forwarding request to backend", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	srv.copyHeader(w.Header(), resp.Header)

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// Start initializes the proxy server and starts listening for incoming requests
func (srv *Server) Start() {
	handler := http.HandlerFunc(srv.proxyHandler)
	http.Handle(srv.endpoint, handler)

	// Start the server on the specified port
	fmt.Printf("Starting server on %s\n", srv.proxyPort)
	if err := http.ListenAndServe(srv.proxyPort, nil); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}
