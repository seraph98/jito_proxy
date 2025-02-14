package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var jito_urls = []string{
	"https://amsterdam.mainnet.block-engine.jito.wtf/api/v1/transactions?bundleOnly=true",
	"https://frankfurt.mainnet.block-engine.jito.wtf/api/v1/transactions?bundleOnly=true",
	"https://ny.mainnet.block-engine.jito.wtf/api/v1/transactions?bundleOnly=true",
	"https://tokyo.mainnet.block-engine.jito.wtf/api/v1/transactions?bundleOnly=true",
	"https://slc.mainnet.block-engine.jito.wtf/api/v1/transactions?bundleOnly=true",
}

type JitoResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Error   struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    any    `json:"data"`
	} `json:"error"`
	Result struct {
		Context struct {
			Slot int `json:"slot"`
		} `json:"context"`
		Value struct {
			Err        interface{} `json:"err"`
			Accounts   interface{} `json:"accounts"`
			Logs       []string    `json:"logs"`
			ReturnData struct {
				Data      []string `json:"data"`
				ProgramID string   `json:"programId"`
			} `json:"returnData"`
			UnitsConsumed int `json:"unitsConsumed"`
		} `json:"value"`
	} `json:"result"`
	ID int `json:"id"`
}

func sendTransactionHandler(w http.ResponseWriter, r *http.Request) {
	var data map[string]interface{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, status_code := sendJito(data)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status_code)
	json.NewEncoder(w).Encode(resp)
}

func sendJito(data map[string]interface{}) (resp JitoResponse, status int) {
	start := time.Now().UnixMilli()
	defer func() {
		end := time.Now().UnixMilli()
		emitQps("send_jito", fmt.Sprintf("%d", status))
		emitLatency(float64(end-start)/1000, "send_jito", fmt.Sprintf("%d", status))
	}()
	jito_url := jito_urls[rand.Intn(len(jito_urls))]

	// Choose a random proxy from the list
	proxy := proxy_urls[rand.Intn(len(proxy_urls))]
	client := &http.Client{Timeout: 10 * time.Second}

	// Prepare the data for the request
	jsonData, err := json.Marshal(data)
	if err != nil {
		resp.Error.Message = fmt.Sprintf("error marshalling data: %v", err)
		return resp, http.StatusInternalServerError
	}

	// Create the POST request
	req, err := http.NewRequest("POST", jito_url, strings.NewReader(string(jsonData)))
	if err != nil {
		resp.Error.Message = fmt.Sprintf("error creating request: %v", err)
		return resp, http.StatusInternalServerError
	}

	// Add necessary headers
	req.Header.Add("Content-Type", "application/json")

	// Assign proxy if not using local address
	if proxy != "" {
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			resp.Error.Message = fmt.Sprintf("error parsing proxy URL: %v", err)
			return resp, http.StatusInternalServerError
		}
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
		client.Transport = transport
	}

	// Send the request using the client
	raw_resp, err := client.Do(req)
	if err != nil {
		resp.Error.Message = fmt.Sprintf("error sending request through proxy %s: %v", proxy, err)
		return resp, http.StatusInternalServerError
	}
	defer raw_resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(raw_resp.Body)
	if err != nil {
		resp.Error.Message = fmt.Sprintf("error reading response body: %v", err)
		return resp, http.StatusInternalServerError
	}

	// If the response is in JSON format, we try to parse it into a map
	if err := json.Unmarshal(body, &resp); err != nil {
		resp.Error.Message = fmt.Sprintf("error unmarshalling JSON response: %v", err)
		return resp, http.StatusInternalServerError
	}
	return resp, raw_resp.StatusCode
}

func main() {
	http.HandleFunc("/", sendTransactionHandler)

	// Start the server
	fmt.Println("Starting the Go server on port 6555...")
	if err := http.ListenAndServe(":6555", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
