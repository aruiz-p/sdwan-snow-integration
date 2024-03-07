package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"snow-sdwan/config"
	"strconv"
)

func createIncident(data map[string]interface{}) error {

	issueId := data["uuid"].(string)
	title := data["message"].(string)
	severity := data["severity_number"].(float64)
	severityStr := strconv.FormatFloat(severity, 'f', -1, 64)
	device := " - Issue found in Device" + data["host_name"].(string) +
		"with System-ip" + data["system_ip"].(string)

	// Construct JSON payload for creating incident
	incidentData := map[string]interface{}{
		"category":          "network",
		"caller_id":         "vManage",
		"short_description": issueId,
		"description":       title + device,
		"urgency":           severityStr,
		"impact":            severityStr,
	}

	// Convert incident data to JSON
	jsonData, err := json.Marshal(incidentData)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %v", err)
	}

	client := &http.Client{}

	// Send HTTP POST request to ServiceNow API endpoint to create incident
	req, err := http.NewRequest("POST", config.SNOW_INSTANCE+"/api/now/v1/table/incident", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	// Add basic authentication header
	auth := config.SNOW_USER + ":" + config.SNOW_PASS
	basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Set("Authorization", basicAuth)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		// Read response body to get more detailed error message
		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error reading response body: %v", err)
		}
		return fmt.Errorf("error creating incident. Status code: %d, Response body: %s", resp.StatusCode, responseBody)
	}

	// Log or report success status
	fmt.Printf("Incident with ID %s was created", issueId)

	return nil
}

func getIncident(issueId string) (bool, error) {

	// Send HTTP GET request to ServiceNow API endpoint to get incident
	resp, err := http.Get(fmt.Sprintf(config.SNOW_INSTANCE+"/api/now/v1/table/%s", issueId))
	if err != nil {
		return false, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status code
	switch resp.StatusCode {
	case http.StatusOK:
		// Incident exists
		return true, nil
	case http.StatusNotFound:
		// Incident does not exist
		return false, nil
	default:
		// Other error occurred
		return false, fmt.Errorf("error retrieving incident. Status code: %d", resp.StatusCode)
	}
}

func closeIncident(issueId string) error {
	// Construct JSON payload for updating incident status
	updateData := map[string]interface{}{
		"state":       "6",
		"close_notes": "Incident closed automatically through Webhooks",
		"close_code":  "Auto-resolved by vManage",
	}

	// Convert update data to JSON
	jsonData, err := json.Marshal(updateData)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %v", err)
	}

	// Create HTTP client
	client := &http.Client{}

	// Create PUT request to update incident status
	req, err := http.NewRequest("PUT", fmt.Sprintf(config.SNOW_INSTANCE+"/api/now/v1/table/incident/%s", issueId), bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(config.SNOW_USER, config.SNOW_PASS) // Replace with your ServiceNow credentials

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != http.StatusOK {
		// Read response body to get more detailed error message
		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error reading response body: %v", err)
		}
		return fmt.Errorf("error closing incident. Status code: %d, Response body: %s", resp.StatusCode, responseBody)
	}

	fmt.Printf("Incident %s closed successfully", issueId)

	return nil
}
