package nyc

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"parkit-data-ETL/internal/models"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

type Config struct {
	BaseURL string
	Key     string
}

type apiResponse []struct {
	ObjectID     string `json:"objectid"`
	MeterNumber  string `json:"meter_number"`
	Status       string `json:"status"`
	PayByCell    string `json:"pay_by_cell_number"`
	Hours        string `json:"meter_hours"`
	Facility     string `json:"facility"`
	FacilityName string `json:"facility_name"`
	Borough      string `json:"borough"`
	OnStreet     string `json:"on_street"`
	FromStreet   string `json:"from_street"`
	ToStreet     string `json:"to_street"`
	SideStreet   string `json:"side_of_street"`
	Latitude     string `json:"lat"`
	Longitude    string `json:"long"`
	Location     struct {
		Type        string    `json:"type"`
		Coordinates []float64 `json:"coordinates"`
	} `json:"location"`
}

func NewClient(config Config) *Client {
	return &Client{
		baseURL: config.BaseURL,
		apiKey:  config.Key,
		client:  &http.Client{},
	}
}

func (c *Client) FetchParkingMeters(offset int) ([]models.ParkingMeter, error) {
	// No limit to fetch all records at once
	url := fmt.Sprintf("%s?$offset=%d", c.baseURL, offset)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-App-Token", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	meters := make([]models.ParkingMeter, 0, len(apiResp))
	for _, result := range apiResp {
		// Debug logging for specific meter
		if result.MeterNumber == "4863002" {
			rawJSON, _ := json.MarshalIndent(result, "", "  ")
			log.Printf("Found meter 4863002 - Raw data:\n%s", string(rawJSON))
		}

		duration, vehicleType, meterDays, meterHours := parseMeterInfo(result.Hours)

		// Skip records with missing facility data
		if result.Facility == "" {
			log.Printf("Warning: Skipping meter %s due to missing facility data", result.MeterNumber)
			continue
		}

		// Check if it's an on-street meter
		isOnStreet := strings.EqualFold(result.Facility, "on street")
		if isOnStreet {
			// Validate required street information for on-street meters
			if result.OnStreet == "" || result.FromStreet == "" || result.ToStreet == "" || result.SideStreet == "" {
				log.Printf("Warning: Skipping on-street meter %s due to missing street information", result.MeterNumber)
				continue
			}
		} else {
			// Must be off-street, validate facility name
			if result.FacilityName == "" {
				// Try to construct facility name from street info, but only if they look like valid street names
				if result.FromStreet != "" && result.OnStreet != "" &&
				   !strings.EqualFold(result.FromStreet, "null") && !strings.EqualFold(result.OnStreet, "null") &&
				   !strings.EqualFold(result.FromStreet, "n/a") && !strings.EqualFold(result.OnStreet, "n/a") {
					// Format street names with proper case and handle ordinal suffixes
					fromStreet := strings.Title(strings.ToLower(result.FromStreet))
					onStreet := strings.Title(strings.ToLower(result.OnStreet))
					// Convert "31 STREET" to "31st Street"
					if strings.Contains(onStreet, " Street") {
						parts := strings.Split(onStreet, " ")
						if num, err := strconv.Atoi(parts[0]); err == nil {
							suffix := "th"
							switch num % 10 {
							case 1:
								if num%100 != 11 {
									suffix = "st"
								}
							case 2:
								if num%100 != 12 {
									suffix = "nd"
								}
							case 3:
								if num%100 != 13 {
									suffix = "rd"
								}
							}
							onStreet = fmt.Sprintf("%d%s %s", num, suffix, parts[1])
						}
					}
					result.FacilityName = fmt.Sprintf("%s-%s", fromStreet, onStreet)
				} else {
					log.Printf("Warning: Skipping off-street meter %s due to missing facility name and invalid street info", result.MeterNumber)
					continue
				}
			}
		}

		objectID, err := strconv.Atoi(result.ObjectID)
		if err != nil {
			log.Printf("Warning: Failed to convert ObjectID '%s' to integer: %v", result.ObjectID, err)
			continue
		}

		meters = append(meters, models.ParkingMeter{
			ObjectID:     objectID,
			MeterNumber:  result.MeterNumber,
			Status:       result.Status,
			PayByCell:    result.PayByCell,
			VehicleType:  vehicleType,
			Duration:     duration,
			MeterDays:    meterDays,
			MeterHours:   meterHours,
			Facility:     !isOnStreet, // false for on-street, true for off-street
			FacilityName: result.FacilityName,
			Borough:      result.Borough,
			OnStreet:     result.OnStreet,
			FromStreet:   result.FromStreet,
			ToStreet:     result.ToStreet,
			SideOfStreet: result.SideStreet,
			Location: models.Point{
				Type:        result.Location.Type,
				Coordinates: result.Location.Coordinates,
			},
		})
	}

	return meters, nil
}

func (c *Client) GetTotalCount() (int, error) {
	req, err := http.NewRequest("GET", c.baseURL+"?$select=count(*)", nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("X-App-Token", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result []struct {
		Count string `json:"count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	if len(result) == 0 {
		return 0, fmt.Errorf("no count result returned")
	}

	count, err := strconv.Atoi(result[0].Count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func parseMeterInfo(hours string) (duration int, vehicleType models.VehicleType, meterDays models.MeterDays, meterHours models.MeterHours) {
	if hours == "" {
		return 0, models.PAS, models.MeterDays{}, models.MeterHours{}
	}

	parts := strings.Fields(hours) // Split by whitespace
	if len(parts) == 0 {
		return 0, models.PAS, models.MeterDays{}, models.MeterHours{}
	}

	// Parse duration (first number)
	duration, _ = strconv.Atoi(parts[0])

	// Parse vehicle type
	vehicleType = models.PAS // default
	for _, part := range parts {
		switch strings.ToUpper(part) {
		case "PAS":
			vehicleType = models.PAS
		case "COM", "COMM":
			vehicleType = models.COMM
		}
	}

	// Parse days
	meterDays = models.MeterDays{} // all false by default
	for _, part := range parts {
		switch strings.ToUpper(part) {
		case "MON-FRI":
			meterDays.Monday = true
			meterDays.Tuesday = true
			meterDays.Wednesday = true
			meterDays.Thursday = true
			meterDays.Friday = true
		case "MON-SAT":
			meterDays.Monday = true
			meterDays.Tuesday = true
			meterDays.Wednesday = true
			meterDays.Thursday = true
			meterDays.Friday = true
			meterDays.Saturday = true
		case "MON-SUN":
			meterDays.Monday = true
			meterDays.Tuesday = true
			meterDays.Wednesday = true
			meterDays.Thursday = true
			meterDays.Friday = true
			meterDays.Saturday = true
			meterDays.Sunday = true
		}
	}

	// Parse hours (last part, format: "0900-1900")
	for _, part := range parts {
		if strings.Contains(part, "-") && len(part) == 9 {
			timeRange := strings.Split(part, "-")
			if len(timeRange) == 2 {
				startHour, _ := strconv.Atoi(timeRange[0][:2])
				startMin, _ := strconv.Atoi(timeRange[0][2:])
				endHour, _ := strconv.Atoi(timeRange[1][:2])
				endMin, _ := strconv.Atoi(timeRange[1][2:])

				meterHours = models.MeterHours{
					StartTime: time.Date(0, 1, 1, startHour, startMin, 0, 0, time.UTC),
					EndTime:   time.Date(0, 1, 1, endHour, endMin, 0, 0, time.UTC),
				}
			}
		}
	}

	return
}
