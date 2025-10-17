package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OneCService handles communication with the 1C proxy service.
type OneCService struct {
	baseURL    string
	httpClient *http.Client
}

// NewOneCService creates a new instance of OneCService.
func NewOneCService(baseURL string) *OneCService {
	return &OneCService{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetSchedule fetches the general schedule from the 1C service.
func (s *OneCService) GetSchedule() (interface{}, error) {
	resp, err := s.httpClient.Get(s.baseURL + "/getschedule")
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

// GetDoctorSchedule fetches a specific doctor's schedule by phone number from the 1C service.
func (s *OneCService) GetDoctorSchedule(phone string) (interface{}, error) {
	reqURL := fmt.Sprintf("%s/getdoctorschedule?phone=%s", s.baseURL, phone)
	resp, err := s.httpClient.Get(reqURL)
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
