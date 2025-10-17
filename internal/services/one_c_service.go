package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OneCService handles communication with the 1C proxy service.
type OneCService struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewOneCService creates a new instance of OneCService.
func NewOneCService(baseURL string, apiKey string) *OneCService {
	return &OneCService{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// executeRequest handles making an authorized request to the 1C service.
func (s *OneCService) executeRequest(url string) (interface{}, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for 1C service: %w", err)
	}

	if s.apiKey != "" {
		req.Header.Set("Authorization", "Basic "+s.apiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to 1C service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("1C service returned non-OK status: %d - %s", resp.StatusCode, string(body))
	}

	var data interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response from 1C service: %w", err)
	}

	return data, nil
}

// GetSchedule fetches a patient's appointment time by phone number from the 1C service.
func (s *OneCService) GetSchedule(phone string) (interface{}, error) {
	// Remove the leading '+' if it exists
	cleanedPhone := strings.TrimPrefix(phone, "+")
	reqURL := fmt.Sprintf("%s/getschedule?phone=%s", s.baseURL, cleanedPhone)
	return s.executeRequest(reqURL)
}

// GetDoctorSchedule fetches the general schedule from the 1C service.
func (s *OneCService) GetDoctorSchedule() (interface{}, error) {
	reqURL := s.baseURL + "/getdoctorschedule"
	return s.executeRequest(reqURL)
}
