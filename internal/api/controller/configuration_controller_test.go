package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bassista/go_spin/internal/config"
	"github.com/gin-gonic/gin"
)

func TestConfigurationController_GetConfiguration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name            string
		baseUrl         string
		refreshInterval int
		expectedStatus  int
		expectedBody    ConfigurationResponse
	}{
		{
			name:            "returns configuration with baseUrl",
			baseUrl:         "https://example.com",
			refreshInterval: 42,
			expectedStatus:  http.StatusOK,
			expectedBody:    ConfigurationResponse{BaseUrl: "https://example.com", RefreshIntervalSec: 42},
		},
		{
			name:            "returns configuration with empty baseUrl",
			baseUrl:         "",
			refreshInterval: 60,
			expectedStatus:  http.StatusOK,
			expectedBody:    ConfigurationResponse{BaseUrl: "", RefreshIntervalSec: 60},
		},
		{
			name:            "returns configuration with baseUrl containing $1 token",
			baseUrl:         "https://$1.my.domain.com",
			refreshInterval: 30,
			expectedStatus:  http.StatusOK,
			expectedBody:    ConfigurationResponse{BaseUrl: "https://$1.my.domain.com", RefreshIntervalSec: 30},
		},
		{
			name:            "returns configuration with localhost baseUrl",
			baseUrl:         "http://localhost/",
			refreshInterval: 15,
			expectedStatus:  http.StatusOK,
			expectedBody:    ConfigurationResponse{BaseUrl: "http://localhost/", RefreshIntervalSec: 15},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config with the test baseUrl and refreshInterval
			cfg := &config.Config{
				Data: config.DataConfig{
					BaseUrl:             tt.baseUrl,
					RefreshIntervalSecs: tt.refreshInterval,
				},
			}

			// Create controller
			controller := NewConfigurationController(cfg)

			// Create test router
			router := gin.New()
			router.GET("/configuration", controller.GetConfiguration)

			// Create request
			req, err := http.NewRequest(http.MethodGet, "/configuration", nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			// Perform request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Parse response body
			var response ConfigurationResponse
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			// Check response
			if response.BaseUrl != tt.expectedBody.BaseUrl {
				t.Errorf("expected baseUrl %q, got %q", tt.expectedBody.BaseUrl, response.BaseUrl)
			}
			if response.RefreshIntervalSec != tt.expectedBody.RefreshIntervalSec {
				t.Errorf("expected refreshIntervalSec %d, got %d", tt.expectedBody.RefreshIntervalSec, response.RefreshIntervalSec)
			}
		})
	}
}

func TestNewConfigurationController(t *testing.T) {
	cfg := &config.Config{
		Data: config.DataConfig{
			BaseUrl: "https://test.com",
		},
	}

	controller := NewConfigurationController(cfg)

	if controller == nil {
		t.Error("expected controller to be created, got nil")
	}

	if controller.config != cfg {
		t.Error("expected controller config to match provided config")
	}
}
