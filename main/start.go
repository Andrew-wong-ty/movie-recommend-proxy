package main

import (
	"flag"
	"proxy/proxy"
)

func main() {
	// Define flags
	maxRequest := flag.Int("maxRequest", 10, "Maximum number of requests to allow within 'expireTime' time")
	expireTime := flag.Int("expireTime", 1, "Time (in seconds) until a request quota expires")
	rateLimitKey := flag.String("rateLimitKey", "MLiP-recommend-endpoint", "Redis key for rate limiting")
	redisAddr := flag.String("redisAddr", "localhost:6379", "Redis server address")
	password := flag.String("password", "secretpassword", "Password for Redis server")
	backendURL := flag.String("backendURL", "http://localhost:8082", "URL of the backend server")
	proxyPort := flag.String("proxyPort", ":8083", "Port for the proxy server to listen on")
	endpoint := flag.String("endpoint", "/recommend/", "Endpoint for recommendation")

	// Parse the flags
	flag.Parse()

	// Create a request rate limiter using the parsed flags
	limiter := proxy.NewRateLimiter(*maxRequest, *expireTime, *rateLimitKey, *redisAddr, *password)

	// Initialize the proxy server with parsed flags
	proxyServer := proxy.NewServer(*backendURL, *proxyPort, *endpoint, limiter)

	// Start the proxy server
	proxyServer.Start()
}
