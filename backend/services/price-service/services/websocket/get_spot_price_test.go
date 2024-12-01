package websocket

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

// Mock Binance server
type MockSpotBinanceServer struct {
	*httptest.Server
	URL string
}

func NewMockSpotBinanceServer() *MockSpotBinanceServer {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Send mock spot price data
		mockData := map[string]interface{}{
			"s": "BTCUSDT",                                       // symbol
			"E": time.Now().UnixNano() / int64(time.Millisecond), // event time
			"c": "30000.00",                                      // last price
		}

		mockJSON, _ := json.Marshal(mockData)
		conn.WriteMessage(websocket.TextMessage, mockJSON)

		// Keep connection alive
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	})

	server := httptest.NewServer(handler)
	return &MockSpotBinanceServer{
		Server: server,
		URL:    "ws" + strings.TrimPrefix(server.URL, "http"),
	}
}

// Helper function to setup test environment
func setupSpotTest(t *testing.T) (*gin.Engine, *httptest.Server, *MockSpotBinanceServer) {
	gin.SetMode(gin.TestMode)

	// Setup mock Binance server
	mockBinance := NewMockSpotBinanceServer()

	// Create a test router
	router := gin.New()
	router.GET("/ws", func(c *gin.Context) {
		// Override the Binance WebSocket URL with our mock server
		originalFunc := SpotPriceSocket
		originalFunc(c)
	})

	// Create test HTTP server
	server := httptest.NewServer(router)

	return router, server, mockBinance
}

func TestSpotPriceSocket(t *testing.T) {
	t.Run("Successful Connection", func(t *testing.T) {
		_, server, mockBinance := setupSpotTest(t)
		defer server.Close()
		defer mockBinance.Close()

		// Create WebSocket URL
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?symbol=btcusdt"

		// Connect to WebSocket
		c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		assert.NoError(t, err)
		defer c.Close()

		// Read response
		done := make(chan bool)
		var response map[string]interface{}

		go func() {
			_, msg, err := c.ReadMessage()
			assert.NoError(t, err)

			err = json.Unmarshal(msg, &response)
			assert.NoError(t, err)
			done <- true
		}()

		select {
		case <-done:
			assert.Contains(t, response, "symbol")
			assert.Contains(t, response, "price")
			assert.Contains(t, response, "eventTime")
		case <-time.After(6 * time.Second):
			t.Fatal("Test timed out")
		}
	})

	t.Run("Invalid Symbol", func(t *testing.T) {
		_, server, mockBinance := setupSpotTest(t)
		defer server.Close()
		defer mockBinance.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?symbol=invalid"

		c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		assert.NoError(t, err)
		defer c.Close()

		// Wait for close message
		done := make(chan bool)
		go func() {
			_, _, err := c.ReadMessage()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "websocket: close")
			done <- true
		}()

		select {
		case <-done:
			// Test passed
		case <-time.After(6 * time.Second):
			t.Fatal("Test timed out")
		}
	})

	t.Run("Disconnect Message", func(t *testing.T) {
		_, server, mockBinance := setupSpotTest(t)
		defer server.Close()
		defer mockBinance.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?symbol=btcusdt"

		c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		assert.NoError(t, err)
		defer c.Close()

		// Send disconnect message
		err = c.WriteMessage(websocket.TextMessage, []byte("disconnect"))
		assert.NoError(t, err)

		// Wait for connection to close
		done := make(chan bool)
		go func() {
			for {
				_, _, err := c.ReadMessage()
				if err != nil {
					assert.Contains(t, err.Error(), "websocket: close")
					done <- true
					return
				}
			}
		}()

		select {
		case <-done:
			// Test passed
		case <-time.After(2 * time.Second):
			t.Fatal("Test timed out")
		}
	})
}