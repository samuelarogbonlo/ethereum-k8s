package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Configuration
type Config struct {
	GethURL      string
	LighthouseURL string
	LogFile      string
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

// Send a JSON-RPC request
func sendRPCRequest(url string, method string, params interface{}) (*RPCResponse, error) {
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

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBytes))
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	var response RPCResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return &response, nil
}

// Get the latest block number
func getLatestBlockNumber(url string) (string, error) {
	response, err := sendRPCRequest(url, "eth_blockNumber", []interface{}{})
	if err != nil {
		return "", err
	}

	if response.Error != nil {
		return "", fmt.Errorf("RPC error: %v", response.Error)
	}

	return fmt.Sprintf("%v", response.Result), nil
}

// Get the number of peers
func getPeerCount(url string) (int, error) {
	response, err := sendRPCRequest(url, "net_peerCount", []interface{}{})
	if err != nil {
		return 0, err
	}

	if response.Error != nil {
		return 0, fmt.Errorf("RPC error: %v", response.Error)
	}

	// Convert hex to int
	hexStr := response.Result.(string)
	var count int
	fmt.Sscanf(hexStr, "0x%x", &count)
	return count, nil
}

// Check Node Health
func checkNodeHealth(config Config, logger *log.Logger) int {
	timestamp := time.Now().Format(time.RFC3339)

	// Check Geth
	blockNumber, err := getLatestBlockNumber(config.GethURL)
	if err != nil {
		logger.Printf("[%s] ERROR: Failed to get latest block number: %v", timestamp, err)
		return 1
	}

	peerCount, err := getPeerCount(config.GethURL)
	if err != nil {
		logger.Printf("[%s] ERROR: Failed to get peer count: %v", timestamp, err)
		return 1
	}

	// Log results
	logger.Printf("[%s] INFO: Latest block number: %s", timestamp, blockNumber)
	logger.Printf("[%s] INFO: Connected peers: %d", timestamp, peerCount)

	// Health check logic
	if peerCount == 0 {
		logger.Printf("[%s] WARNING: No peers connected", timestamp)
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
		GethURL:      "http://localhost:8545",
		LighthouseURL: "http://localhost:5052",
		LogFile:      filepath.Join(homeDir, ".ethereum", "health-check.log"),
	}

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

	logger := initLogger(config.LogFile)

	// Run health check
	exitCode := checkNodeHealth(config, logger)
	os.Exit(exitCode)
}