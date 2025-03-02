package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Configuration
const (
	executionRPCEndpoint = "http://localhost:30545" // Update with your node's IP if not using local
	timeoutSeconds       = 10
	minPeers             = 1 // Minimum number of peers to consider healthy
)

// RPC request structure
type RPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

// RPC response structure
type RPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Result  interface{} `json:"result"`
	Error   *RPCError   `json:"error,omitempty"`
}

// RPC error structure
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// SendRPCRequest sends a JSON-RPC request to the Ethereum node
func SendRPCRequest(endpoint string, method string, params []interface{}) (*RPCResponse, error) {
	request := RPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: time.Duration(timeoutSeconds) * time.Second,
	}

	// Send request
	resp, err := client.Post(endpoint, "application/json", bytes.NewBuffer(requestJSON))
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	// Parse response
	var response RPCResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}

	// Check for RPC error
	if response.Error != nil {
		return nil, fmt.Errorf("RPC error: %s (code: %d)", response.Error.Message, response.Error.Code)
	}

	return &response, nil
}

// HexToInt converts a hex string to an integer
func HexToInt(hex string) (uint64, error) {
	// Remove "0x" prefix if present
	if len(hex) >= 2 && hex[0:2] == "0x" {
		hex = hex[2:]
	}
	return strconv.ParseUint(hex, 16, 64)
}

func main() {
	var exitCode = 0
	var healthStatus = "HEALTHY"

	fmt.Println("====================================")
	fmt.Println("  Ethereum Node Health Check")
	fmt.Println("====================================")

	// Check 1: Node connection
	fmt.Println("Checking node connection...")
	clientVersionResp, err := SendRPCRequest(executionRPCEndpoint, "web3_clientVersion", []interface{}{})
	if err != nil {
		fmt.Printf("❌ Node connection failed: %v\n", err)
		os.Exit(1)
	}
	clientVersion := clientVersionResp.Result.(string)
	fmt.Printf("✅ Connected to node: %s\n", clientVersion)

	// Check 2: Block height
	fmt.Println("\nChecking latest block...")
	blockResp, err := SendRPCRequest(executionRPCEndpoint, "eth_blockNumber", []interface{}{})
	if err != nil {
		fmt.Printf("❌ Failed to get latest block: %v\n", err)
		exitCode = 1
		healthStatus = "UNHEALTHY"
	} else {
		blockHex := blockResp.Result.(string)
		blockNum, err := HexToInt(blockHex)
		if err != nil {
			fmt.Printf("❌ Failed to parse block number: %v\n", err)
			exitCode = 1
			healthStatus = "UNHEALTHY"
		} else {
			fmt.Printf("✅ Latest block: %d\n", blockNum)

			// Check if block was updated in the last hour
			syncResp, err := SendRPCRequest(executionRPCEndpoint, "eth_syncing", []interface{}{})
			if err != nil {
				fmt.Printf("⚠️ Unable to check sync status: %v\n", err)
			} else {
				// If syncing is false, node is in sync or not syncing
				if syncBool, ok := syncResp.Result.(bool); ok && !syncBool {
					// Get block timestamp
					blockParamHex := fmt.Sprintf("0x%x", blockNum)
					blockInfoResp, err := SendRPCRequest(executionRPCEndpoint, "eth_getBlockByNumber", []interface{}{blockParamHex, false})
					if err == nil {
						blockInfo := blockInfoResp.Result.(map[string]interface{})
						timestampHex := blockInfo["timestamp"].(string)
						timestamp, err := HexToInt(timestampHex)
						if err == nil {
							blockTime := time.Unix(int64(timestamp), 0)
							timeSinceLastBlock := time.Since(blockTime)

							if timeSinceLastBlock.Hours() > 1 {
								fmt.Printf("⚠️ Last block is over an hour old (%s ago)\n", timeSinceLastBlock.Round(time.Second))
							} else {
								fmt.Printf("✅ Last block timestamp: %s (%s ago)\n",
									blockTime.Format(time.RFC3339),
									timeSinceLastBlock.Round(time.Second))
							}
						}
					}
				} else if syncObj, ok := syncResp.Result.(map[string]interface{}); ok {
					// Node is syncing
					currentBlockHex := syncObj["currentBlock"].(string)
					highestBlockHex := syncObj["highestBlock"].(string)

					currentBlock, _ := HexToInt(currentBlockHex)
					highestBlock, _ := HexToInt(highestBlockHex)

					syncPercentage := float64(currentBlock) / float64(highestBlock) * 100
					fmt.Printf("⚠️ Node is syncing: %.2f%% complete (%d/%d blocks)\n",
						syncPercentage, currentBlock, highestBlock)
				}
			}
		}
	}

	// Check 3: Peer count
	fmt.Println("\nChecking peer connections...")
	peerResp, err := SendRPCRequest(executionRPCEndpoint, "net_peerCount", []interface{}{})
	if err != nil {
		fmt.Printf("❌ Failed to get peer count: %v\n", err)
		exitCode = 1
		healthStatus = "UNHEALTHY"
	} else {
		peerCountHex := peerResp.Result.(string)
		peerCount, err := HexToInt(peerCountHex)
		if err != nil {
			fmt.Printf("❌ Failed to parse peer count: %v\n", err)
			exitCode = 1
			healthStatus = "UNHEALTHY"
		} else {
			if peerCount < minPeers {
				fmt.Printf("⚠️ Low peer count: %d (minimum: %d)\n", peerCount, minPeers)
				// Don't fail the health check just for low peer count
			} else {
				fmt.Printf("✅ Connected peers: %d\n", peerCount)
			}
		}
	}

	// Check 4: Network ID
	fmt.Println("\nChecking network ID...")
	netResp, err := SendRPCRequest(executionRPCEndpoint, "net_version", []interface{}{})
	if err != nil {
		fmt.Printf("❌ Failed to get network ID: %v\n", err)
	} else {
		networkID := netResp.Result.(string)
		var networkName string

		switch networkID {
		case "1":
			networkName = "Ethereum Mainnet"
		case "5":
			networkName = "Goerli Testnet"
		case "11155111":
			networkName = "Sepolia Testnet"
		default:
			networkName = "Unknown"
		}

		fmt.Printf("✅ Network: %s (ID: %s)\n", networkName, networkID)
	}

	// Final health status
	fmt.Println("\n====================================")
	if healthStatus == "HEALTHY" {
		fmt.Println("✅ Node status: HEALTHY")
	} else {
		fmt.Println("❌ Node status: UNHEALTHY")
	}
	fmt.Println("====================================")

	// Log result
	logFileName := fmt.Sprintf("healthcheck-%s.log", time.Now().Format("20060102-150405"))
	logFile, err := os.Create(logFileName)
	if err == nil {
		defer logFile.Close()
		log.SetOutput(logFile)
		log.Printf("Health check result: %s\n", healthStatus)
		fmt.Printf("Results logged to: %s\n", logFileName)
	}

	os.Exit(exitCode)
}