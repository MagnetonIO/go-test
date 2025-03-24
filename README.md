# Bitcoin Last Traded Price (LTP) Service

This service provides a REST API to retrieve the Last Traded Price (LTP) of Bitcoin for USD, EUR, and CHF currency pairs.

## Features

- Retrieves real-time Bitcoin price data from Kraken API
- Supports querying individual or multiple currency pairs
- Caches data for 1 minute to reduce API calls
- Dockerized for easy deployment
- Integration tests included

## API Reference

### Get Last Traded Price

```
GET /api/v1/ltp
```

**Query Parameters:**

- `pair` (optional): Currency pair(s) to get prices for. Can be specified multiple times.
  - Supported values: `BTC/USD`, `BTC/EUR`, `BTC/CHF`
  - If no pair is specified, all supported pairs are returned

**Response:**

```json
{
  "ltp": [
    {
      "pair": "BTC/USD",
      "amount": 52000.12
    },
    {
      "pair": "BTC/EUR",
      "amount": 50000.12
    },
    {
      "pair": "BTC/CHF",
      "amount": 49000.12
    }
  ]
}
```

## Building and Running

### Local Development

1. Ensure Go 1.20+ is installed on your machine
2. Clone the repository:
   ```
   git clone https://github.com/yourusername/btc-ltp-service.git
   cd btc-ltp-service
   ```
3. Build the application:
   ```
   go build -o btc-ltp-service
   ```
4. Run the application:
   ```
   ./btc-ltp-service
   ```
5. The service will be available at `http://localhost:8080`

### Using Docker

1. Build the Docker image:
   ```
   docker build -t btc-ltp-service .
   ```
2. Run the container:
   ```
   docker run -p 8080:8080 btc-ltp-service
   ```
3. The service will be available at `http://localhost:8080`

## Testing

### Running Integration Tests

```
go test -v
```

To skip the Kraken API integration test (useful for CI environments):

```
go test -v -short
```

## Implementation Details

- The service caches LTP data for 1 minute to ensure time accuracy as required
- Data is refreshed in the background every 30 seconds to minimize latency
- The service handles Kraken API's naming conventions (XBT instead of BTC)
- Error handling is implemented for network failures and API errors

## Dependencies

This project has minimal external dependencies:
- Go standard library for HTTP handling and JSON parsing
- Alpine Linux for the Docker container base image

## License

[MIT License](LICENSE)
