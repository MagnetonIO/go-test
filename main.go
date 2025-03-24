package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Pair represents a trading pair
type Pair struct {
	Base    string
	Quote   string
	Display string
}

// LTPResponse represents the response structure
type LTPResponse struct {
	LTP []LTPData `json:"ltp"`
}

// LTPData represents the price data for a pair
type LTPData struct {
	Pair   string  `json:"pair"`
	Amount float64 `json:"amount"`
}

// KrakenTickerResponse represents the response from Kraken API
type KrakenTickerResponse struct {
	Error  []string               `json:"error"`
	Result map[string]KrakenPair `json:"result"`
}

// KrakenPair contains ticker information for a pair
type KrakenPair struct {
	C []string `json:"c"` // c = last trade closed (price, lot volume)
}

// Supported pairs
var supportedPairs = map[string]Pair{
	"BTC/USD": {Base: "XBT", Quote: "USD", Display: "BTC/USD"},
	"BTC/CHF": {Base: "XBT", Quote: "CHF", Display: "BTC/CHF"},
	"BTC/EUR": {Base: "XBT", Quote: "EUR", Display: "BTC/EUR"},
}

// Cache for last traded prices
var priceCache = struct {
	sync.RWMutex
	data          map[string]LTPData
	lastUpdated   time.Time
	updateRunning bool
}{
	data: make(map[string]LTPData),
}

func main() {
	// Initialize and do first cache update
	updatePriceCache()

	// Start periodic updates every 30 seconds
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for range ticker.C {
			updatePriceCache()
		}
	}()

	// Setup API routes
	http.HandleFunc("/api/v1/ltp", handleLTP)

	// Start server
	port := ":8080"
	fmt.Printf("Starting server on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

// handleLTP handles the LTP API endpoint
func handleLTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get requested pairs
	requestedPairs := r.URL.Query()["pair"]

	// If no pairs specified, return all supported pairs
	if len(requestedPairs) == 0 {
		requestedPairs = make([]string, 0, len(supportedPairs))
		for pair := range supportedPairs {
			requestedPairs = append(requestedPairs, pair)
		}
	}

	response := LTPResponse{LTP: []LTPData{}}
	priceCache.RLock()

	// Check if cache needs updating (older than 1 minute)
	needsUpdate := time.Since(priceCache.lastUpdated) > time.Minute

	// Add requested pairs to response
	for _, pairName := range requestedPairs {
		pairName = strings.ToUpper(pairName)
		if _, exists := supportedPairs[pairName]; !exists {
			continue // Skip unsupported pairs
		}

		if data, found := priceCache.data[pairName]; found {
			response.LTP = append(response.LTP, data)
		}
	}
	priceCache.RUnlock()

	// Update cache if needed
	if needsUpdate && !priceCache.updateRunning {
		go updatePriceCache()
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// updatePriceCache fetches latest prices from Kraken and updates cache
func updatePriceCache() {
	priceCache.Lock()
	if priceCache.updateRunning {
		priceCache.Unlock()
		return
	}
	priceCache.updateRunning = true
	priceCache.Unlock()

	defer func() {
		priceCache.Lock()
		priceCache.updateRunning = false
		priceCache.Unlock()
	}()

	// Build comma-separated list of Kraken pairs
	var krakenPairs []string
	for _, pair := range supportedPairs {
		// Kraken uses XBT instead of BTC
		krakenPairs = append(krakenPairs, fmt.Sprintf("%s%s", pair.Base, pair.Quote))
	}

	// Make request to Kraken API
	resp, err := http.Get(fmt.Sprintf("https://api.kraken.com/0/public/Ticker?pair=%s", strings.Join(krakenPairs, ",")))
	if err != nil {
		log.Printf("Error fetching from Kraken: %v", err)
		return
	}
	defer resp.Body.Close()

	var krakenResponse KrakenTickerResponse
	if err := json.NewDecoder(resp.Body).Decode(&krakenResponse); err != nil {
		log.Printf("Error decoding Kraken response: %v", err)
		return
	}

	if len(krakenResponse.Error) > 0 {
		log.Printf("Kraken API error: %v", krakenResponse.Error)
		return
	}

	// Update cache with new prices
	priceCache.Lock()
	defer priceCache.Unlock()

	for displayPair, pairInfo := range supportedPairs {
		krakenPair := fmt.Sprintf("%s%s", pairInfo.Base, pairInfo.Quote)

		// For some pairs, Kraken might add "X" or "Z" prefixes
		altKrakenPair := fmt.Sprintf("X%sZ%s", pairInfo.Base, pairInfo.Quote)

		var ticker KrakenPair
		var found bool

		// Try both formats
		if t, ok := krakenResponse.Result[krakenPair]; ok {
			ticker = t
			found = true
		} else if t, ok := krakenResponse.Result[altKrakenPair]; ok {
			ticker = t
			found = true
		}

		if found && len(ticker.C) >= 1 {
			// Extract and convert price to float
			var price float64
			fmt.Sscanf(ticker.C[0], "%f", &price)

			priceCache.data[displayPair] = LTPData{
				Pair:   displayPair,
				Amount: price,
			}
		}
	}

	priceCache.lastUpdated = time.Now()
}
