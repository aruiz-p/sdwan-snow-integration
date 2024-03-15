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
	"strings"
	"time"
)

func createIncident(data map[string]interface{}) error {

	// Retrieve information to open incident
	issueId := data["uuid"].(string)
	ruleName := data["rule_name_display"].(string)
	title := data["message"].(string)
	severity := data["severity_number"].(float64)
	severityStr := strconv.FormatFloat(severity, 'f', -1, 64)
	device := ". Device " + data["host_name"].(string) +
		", System-ip " + data["system_ip"].(string)

	// Construct JSON payload for creating incident in SNOW
	incidentData := map[string]interface{}{
		"category":          "network",
		"caller_id":         "vManage",
		"short_description": issueId,
		"description":       ruleName + " - " + title + device,
		"urgency":           severityStr,
		"impact":            severityStr,
	}

	// Convert incident data to JSON
	jsonData, err := json.Marshal(incidentData)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %v", err)
	}

	client := &http.Client{}

	// Create incident
	req, err := http.NewRequest("POST", config.SNOW_INSTANCE+"/api/now/v1/table/incident", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	auth := config.SNOW_USER + ":" + config.SNOW_PASS
	basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Set("Authorization", basicAuth)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return fmt.Errorf("error reading response body: %v", err)
	}
	// Read response
	var incidentResponse map[string]interface{}
	err = json.Unmarshal(body, &incidentResponse)
	if err != nil {
		return fmt.Errorf("error decoding JSON response: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		// Read response body to get more detailed error message
		return fmt.Errorf("error creating incident. Status code: %d, Response body: %s", resp.StatusCode, body)

	}

	// Check if the result exists
	result, ok := incidentResponse["result"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing or invalid result field in the response")
	}

	incidentId := result["sys_id"].(string)

	fmt.Printf("Incident %s was created", incidentId)

	return nil
}

func getIncidentWithId(issueId string) (bool, string, error) {

	client := &http.Client{}

	// Send HTTP GET request to ServiceNow API endpoint to get incident
	req, err := http.NewRequest("GET", fmt.Sprintf(config.SNOW_INSTANCE+"/api/now/v1/table/incident"), nil)
	if err != nil {
		return false, "", fmt.Errorf("error creating request: %v", err)
	}

	// Add basic authentication header
	auth := config.SNOW_USER + ":" + config.SNOW_PASS
	basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Set("Authorization", basicAuth)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return false, "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status code
	switch resp.StatusCode {
	case http.StatusOK:
		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, "", fmt.Errorf("error reading response body: %v", err)
		}

		// Unmarshal JSON response into a map
		var incidentResponse map[string]interface{}
		err = json.Unmarshal(body, &incidentResponse)
		if err != nil {
			return false, "", fmt.Errorf("error decoding JSON response: %v", err)
		}

		// Check if the result exists
		results, ok := incidentResponse["result"].([]interface{})
		if !ok {
			return false, "", fmt.Errorf("missing or invalid result field in the response")
		}
		for _, incident := range results {
			incidentMap, ok := incident.(map[string]interface{})
			if !ok {
				return false, "", fmt.Errorf("incident is not a map[string]interface{}")
			}

			// Check if the incident has the "short_description" field
			shortDescription, ok := incidentMap["short_description"].(string)
			if !ok {
				continue
			}

			// Compare the short_description with the issueId
			if strings.Contains(shortDescription, issueId) {
				// If the short_description matches the issueId, return the incident id
				sys_id, ok := incidentMap["sys_id"].(string)
				if !ok {
					continue
				}
				return true, sys_id, nil
			}
		}

		// If no incident with the matching short_description was found, return false
		return false, "", nil

	default:
		// Other error occurred
		return false, "", fmt.Errorf("error retrieving incident. Status code: %d", resp.StatusCode)
	}
}

func getIncidentWoutId(ruleName, sysIp string, openTime float64) (bool, string, error) {
	client := &http.Client{}

	// Send HTTP GET request to ServiceNow API endpoint to get incident
	req, err := http.NewRequest("GET", fmt.Sprintf(config.SNOW_INSTANCE+"/api/now/v1/table/incident"), nil)
	if err != nil {
		return false, "", fmt.Errorf("error creating request: %v", err)
	}

	// Add basic authentication header
	auth := config.SNOW_USER + ":" + config.SNOW_PASS
	basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Set("Authorization", basicAuth)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return false, "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status code
	switch resp.StatusCode {
	case http.StatusOK:

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, "", fmt.Errorf("error reading response body: %v", err)
		}

		// Unmarshal JSON response into a map
		var incidentResponse map[string]interface{}
		err = json.Unmarshal(body, &incidentResponse)
		if err != nil {
			return false, "", fmt.Errorf("error decoding JSON response: %v", err)
		}

		// Check if the result exists
		results, ok := incidentResponse["result"].([]interface{})

		if !ok {
			return false, "", fmt.Errorf("missing or invalid result field in the response")
		}
		for _, incident := range results {
			incidentMap, ok := incident.(map[string]interface{})
			if !ok {
				return false, "", fmt.Errorf("incident is not a map[string]interface{}")
			}

			// Check if the incident has the "short_description" field
			description := incidentMap["description"].(string)
			snowTime := incidentMap["opened_at"].(string)
			layout := "2006-01-02 15:04:05"

			// Parse the string into a time.Time object
			parsedTime, err := time.Parse(layout, snowTime)

			if err != nil {
				fmt.Println("Error parsing time:", err)
			}

			// Convert time.Time object to Unix epoch time in milliseconds
			snowTimeEpoch := parsedTime.UnixNano() / int64(time.Millisecond)
			webhookTimeEpoch := int64(openTime)
			// Calculate how many hours ago incident was opened
			diffMillis := webhookTimeEpoch - snowTimeEpoch
			diffHours := float64(diffMillis) / (1000 * 60 * 60)

			// Convert Rule Name to look on Service Now responses
			newRuleName := strings.Replace(ruleName, "Up", "Down", -1)

			// Compare Rule Name, system ip and time
			if strings.Contains(description, newRuleName) &&
				strings.Contains(description, sysIp) &&
				diffHours < 12 {

				sys_id, ok := incidentMap["sys_id"].(string)
				if !ok {
					continue
				}
				return true, sys_id, nil
			}
		}

		// If no incident with the matching short_description was found, return false
		return false, "", nil

	default:
		// Other error occurred
		return false, "", fmt.Errorf("error retrieving incident. Status code: %d", resp.StatusCode)
	}
}

func closeIncident(incidentId string) error {
	// Construct JSON payload for updating incident status
	updateData := map[string]interface{}{
		"state":       "6",
		"close_notes": "Incident closed automatically through Webhooks",
		"close_code":  "Resolved by caller",
	}

	// Convert update data to JSON
	jsonData, err := json.Marshal(updateData)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %v", err)
	}

	client := &http.Client{}

	// Create PUT request to update incident status
	req, err := http.NewRequest("PUT", fmt.Sprintf(config.SNOW_INSTANCE+"/api/now/v1/table/incident/%s", incidentId), bytes.NewBuffer(jsonData))
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

	fmt.Printf("Incident %s closed successfully", incidentId)

	return nil
}
