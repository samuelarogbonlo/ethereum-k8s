package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Configuration
type Config struct {
	GethURL       string
	LighthouseURL string
	LogFile       string
	Timeout       time.Duration
	Retries       int
	RetryDelay    time.Duration
	Debug         bool
}

// RPCRequest represents a JSON-RPC request
type RPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int         `json:"id"`
}

// RPCResponse represents a JSON-RPC response
type RPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Error   interface{} `json:"error"`
	ID      int         `json:"id"`
}

// Initialize logger
func initLogger(logFile string) *log.Logger {
	// Create the directory for the log file if needed
	logDir := filepath.Dir(logFile)
	if logDir != "." && logDir != "" {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			log.Printf("Warning: Could not create log directory: %v", err)
			// Fallback to current directory
			logFile = "./ethereum-health.log"
		}
	}

	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}

	return log.New(file, "", log.LstdFlags)
}

// Send a JSON-RPC request with retries
func sendRPCRequest(url string, method string, params interface{}, config Config) (*RPCResponse, error) {
	request := RPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	if config.Debug {
		fmt.Printf("DEBUG: Sending request to %s\nMethod: %s\nRequest body: %s\n",
			url, method, string(requestBytes))
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}

	// Try with retries
	var resp *http.Response
	var lastError error

	for attempt := 0; attempt <= config.Retries; attempt++ {
		if attempt > 0 {
			fmt.Printf("Retrying connection to %s (attempt %d/%d)...\n", url, attempt, config.Retries)
			time.Sleep(config.RetryDelay)
		}

		// Create a new request for each attempt
		req, err := http.NewRequestWithContext(
			context.Background(),
			"POST",
			url,
			bytes.NewBuffer(requestBytes),
		)
		if err != nil {
			lastError = err
			continue
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err = client.Do(req)
		if err != nil {
			lastError = err
			if config.Debug {
				fmt.Printf("DEBUG: Connection failed: %v\n", err)
			}
			continue
		}

		// If we got here, we have a response
		break
	}

	if resp == nil {
		return nil, fmt.Errorf("failed after %d attempts: %v", config.Retries, lastError)
	}
	defer resp.Body.Close()

	if config.Debug {
		fmt.Printf("DEBUG: Response status: %s\n", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	if config.Debug {
		fmt.Printf("DEBUG: Response body: %s\n", string(body))
	}

	var response RPCResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return &response, nil
}

// Get the latest block number
func getLatestBlockNumber(url string, config Config) (string, error) {
	response, err := sendRPCRequest(url, "eth_blockNumber", []interface{}{}, config)
	if err != nil {
		return "", err
	}

	if response.Error != nil {
		return "", fmt.Errorf("RPC error: %v", response.Error)
	}

	return fmt.Sprintf("%v", response.Result), nil
}

// Get the sync status
func getSyncStatus(url string, config Config) (bool, error) {
	response, err := sendRPCRequest(url, "eth_syncing", []interface{}{}, config)
	if err != nil {
		return false, err
	}

	if response.Error != nil {
		return false, fmt.Errorf("RPC error: %v", response.Error)
	}

	// If result is false, node is up to date
	// If result is an object, node is syncing
	isSyncing := response.Result != nil && response.Result != false
	return isSyncing, nil
}

// Get the number of peers
func getPeerCount(url string, config Config) (int, error) {
	response, err := sendRPCRequest(url, "net_peerCount", []interface{}{}, config)
	if err != nil {
		return 0, err
	}

	if response.Error != nil {
		return 0, fmt.Errorf("RPC error: %v", response.Error)
	}

	// Convert hex to int
	hexStr, ok := response.Result.(string)
	if !ok {
		return 0, fmt.Errorf("unexpected response format for peer count: %v", response.Result)
	}

	var count int
	fmt.Sscanf(hexStr, "0x%x", &count)
	return count, nil
}

// Check if service is available
func isServiceAvailable(url string, timeout time.Duration) bool {
	client := &http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode < 500
}

// Check Node Health
func checkNodeHealth(config Config, logger *log.Logger) int {
	timestamp := time.Now().Format(time.RFC3339)

	// First check if service is reachable
	fmt.Println("Checking if Geth RPC endpoint is reachable...")
	if !isServiceAvailable(config.GethURL, 2*time.Second) {
		errMsg := fmt.Sprintf("Geth RPC endpoint at %s is not reachable", config.GethURL)
		logger.Printf("[%s] ERROR: %s", timestamp, errMsg)
		fmt.Printf("ERROR: %s\n", errMsg)

		// Additional diagnostics
		fmt.Println("\nTroubleshooting steps:")
		fmt.Println("1. Check if the Geth pod is running:")
		fmt.Println("   kubectl get pod geth-0")
		fmt.Println("2. Check Geth pod logs:")
		fmt.Println("   kubectl logs geth-0")
		fmt.Println("3. Verify the service definition:")
		fmt.Println("   kubectl get svc geth -o yaml")
		fmt.Println("4. Try port-forwarding directly to the pod:")
		fmt.Println("   kubectl port-forward pod/geth-0 8545:8545")

		return 1
	}

	// Check Geth
	fmt.Println("Checking Geth node health...")

	// Check if node is syncing
	isSyncing, err := getSyncStatus(config.GethURL, config)
	if err != nil {
		logger.Printf("[%s] ERROR: Failed to get sync status: %v", timestamp, err)
		fmt.Printf("ERROR: Failed to get sync status: %v\n", err)
	} else {
		syncStatus := "up to date"
		if isSyncing {
			syncStatus = "currently syncing"
		}
		logger.Printf("[%s] INFO: Node sync status: %s", timestamp, syncStatus)
		fmt.Printf("Node sync status: %s\n", syncStatus)
	}

	blockNumber, err := getLatestBlockNumber(config.GethURL, config)
	if err != nil {
		logger.Printf("[%s] ERROR: Failed to get latest block number: %v", timestamp, err)
		fmt.Printf("ERROR: Failed to get latest block number: %v\n", err)
		return 1
	}

	peerCount, err := getPeerCount(config.GethURL, config)
	if err != nil {
		logger.Printf("[%s] ERROR: Failed to get peer count: %v", timestamp, err)
		fmt.Printf("ERROR: Failed to get peer count: %v\n", err)
		return 1
	}

	// Log results
	logger.Printf("[%s] INFO: Latest block number: %s", timestamp, blockNumber)
	logger.Printf("[%s] INFO: Connected peers: %d", timestamp, peerCount)

	// Health check logic
	if peerCount == 0 {
		logger.Printf("[%s] WARNING: No peers connected", timestamp)
		fmt.Printf("WARNING: No peers connected\n")
		return 1
	}

	fmt.Printf("Latest block number: %s\n", blockNumber)
	fmt.Printf("Connected peers: %d\n", peerCount)

	if peerCount >= 2 {
		fmt.Println("Node health status: HEALTHY")
		logger.Printf("[%s] INFO: Node health status: HEALTHY", timestamp)
		return 0
	} else {
		fmt.Println("Node health status: WARNING (Low peer count)")
		logger.Printf("[%s] WARNING: Node health status: WARNING (Low peer count)", timestamp)
		return 1
	}
}

func main() {
	// Get home directory for default log location
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	// Configuration with better defaults
	config := Config{
		GethURL:       "http://localhost:8545",
		LighthouseURL: "http://localhost:5052",
		LogFile:       filepath.Join(homeDir, ".ethereum", "health-check.log"),
		Timeout:       10 * time.Second,
		Retries:       3,
		RetryDelay:    2 * time.Second,
		Debug:         false,
	}

	// Print more debugging info at startup
	fmt.Println("Starting health check with the following configuration:")
	fmt.Printf("  Geth URL: %s\n", config.GethURL)
	fmt.Printf("  Lighthouse URL: %s\n", config.LighthouseURL)
	fmt.Printf("  Log file: %s\n", config.LogFile)
	fmt.Printf("  Timeout: %s\n", config.Timeout)
	fmt.Printf("  Retries: %d\n", config.Retries)
	fmt.Printf("  Retry delay: %s\n", config.RetryDelay)

	// Override configuration from environment variables
	if gethURL := os.Getenv("GETH_URL"); gethURL != "" {
		config.GethURL = gethURL
	}

	if lighthouseURL := os.Getenv("LIGHTHOUSE_URL"); lighthouseURL != "" {
		config.LighthouseURL = lighthouseURL
	}

	if logFile := os.Getenv("LOG_FILE"); logFile != "" {
		config.LogFile = logFile
	}

	if os.Getenv("DEBUG") == "true" {
		config.Debug = true
		fmt.Println("Debug mode enabled")
	}

	logger := initLogger(config.LogFile)
	fmt.Printf("Starting Ethereum health check...\n")
	fmt.Printf("Geth URL: %s\n", config.GethURL)
	fmt.Printf("Lighthouse URL: %s\n", config.LighthouseURL)

	// Run health check
	exitCode := checkNodeHealth(config, logger)
	os.Exit(exitCode)
}