package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLTPEndpoint(t *testing.T) {
	// Initialize cache with some test data
	priceCache.Lock()
	priceCache.data = map[string]LTPData{
		"BTC/USD": {Pair: "BTC/USD", Amount: 52000.12},
		"BTC/EUR": {Pair: "BTC/EUR", Amount: 50000.12},
		"BTC/CHF": {Pair: "BTC/CHF", Amount: 49000.12},
	}
	priceCache.lastUpdated = time.Now()
	priceCache.Unlock()

	// Test cases
	tests := []struct {
		name           string
		url            string
		expectedPairs  []string
		expectedStatus int
	}{
		{
			name:           "Get all pairs",
			url:            "/api/v1/ltp",
			expectedPairs:  []string{"BTC/USD", "BTC/EUR", "BTC/CHF"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get single pair",
			url:            "/api/v1/ltp?pair=BTC/USD",
			expectedPairs:  []string{"BTC/USD"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get multiple pairs",
			url:            "/api/v1/ltp?pair=BTC/USD&pair=BTC/EUR",
			expectedPairs:  []string{"BTC/USD", "BTC/EUR"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get with unsupported pair",
			url:            "/api/v1/ltp?pair=BTC/JPY",
			expectedPairs:  []string{},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tc.url, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(handleLTP)
			handler.ServeHTTP(rr, req)

			// Check status code
			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tc.expectedStatus)
			}

			// Check response body
			if tc.expectedStatus == http.StatusOK {
				var response LTPResponse
				err = json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Could not unmarshal response: %v", err)
				}

				// Check if all expected pairs are in the response
				if len(response.LTP) != len(tc.expectedPairs) {
					t.Errorf("Expected %d pairs, got %d", len(tc.expectedPairs), len(response.LTP))
				}

				// Create a map for easier lookup
				pairsMap := make(map[string]bool)
				for _, item := range response.LTP {
					pairsMap[item.Pair] = true
				}

				// Check if all expected pairs are present
				for _, pair := range tc.expectedPairs {
					if !pairsMap[pair] {
						t.Errorf("Expected pair %s not found in response", pair)
					}
				}
			}
		})
	}
}

func TestKrakenIntegration(t *testing.T) {
	// Skip this test in CI environments or add a flag to enable/disable
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Update cache from Kraken
	updatePriceCache()

	// Check if cache was updated
	priceCache.RLock()
	defer priceCache.RUnlock()

	if time.Since(priceCache.lastUpdated) > time.Minute {
		t.Error("Cache was not updated")
	}

	// Check if we have data for all pairs
	for pair := range supportedPairs {
		if _, ok := priceCache.data[pair]; !ok {
			t.Errorf("Expected data for pair %s not found in cache", pair)
		}
	}

	// Check if amounts are reasonable (BTC should be worth something)
	for pair, data := range priceCache.data {
		if data.Amount <= 0 {
			t.Errorf("Expected positive amount for pair %s, got %f", pair, data.Amount)
		}
	}
}
