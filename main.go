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

	// Check if the response is a string or raw JSON data
	switch resp := resp.(type) {
	case string:
		// If it's a string, return it directly
		w.Write([]byte(resp))
	case []byte:
		// If it's raw JSON, write the raw bytes directly
		w.Write(resp)
	default:
		// Otherwise, assume it's a map and return it as JSON
		json.NewEncoder(w).Encode(resp)
	}
}

func sendJito(data map[string]interface{}) (interface{}, int) {
	jito_url := jito_urls[rand.Intn(len(jito_urls))]

	// Choose a random proxy from the list
	proxy := proxy_urls[rand.Intn(len(proxy_urls))]
	client := &http.Client{Timeout: 10 * time.Second}

	// Prepare the data for the request
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshalling data: %v", err), http.StatusInternalServerError
	}

	// Create the POST request
	req, err := http.NewRequest("POST", jito_url, strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err), http.StatusInternalServerError
	}

	// Add necessary headers
	req.Header.Add("Content-Type", "application/json")

	// Assign proxy if not using local address
	if proxy != "" {
		req.URL.Host = proxy
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			return fmt.Errorf("error parsing proxy URL: %v", err), http.StatusInternalServerError
		}
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
		client.Transport = transport
	}

	// Send the request using the client
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request through proxy %s: %v", proxy, err), http.StatusInternalServerError
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err), http.StatusInternalServerError
	}

	// If the response is in JSON format, we try to parse it into a map
	if resp.Header.Get("Content-Type") == "application/json" {
		var jsonResponse map[string]interface{}
		if err := json.Unmarshal(body, &jsonResponse); err != nil {
			return fmt.Errorf("error unmarshalling JSON response: %v", err), http.StatusInternalServerError
		}
		return jsonResponse, resp.StatusCode
	}

	// If not JSON, return the raw body as a string
	return body, resp.StatusCode
}

func main() {
	http.HandleFunc("/send_transaction", sendTransactionHandler)

	// Start the server
	fmt.Println("Starting the Go server on port 5000...")
	if err := http.ListenAndServe(":5000", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
