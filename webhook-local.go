package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	// Open the JSON file for reading
	file, err := os.Open("open.json")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()

	// Create a variable to hold the decoded JSON data
	var data map[string]interface{}

	// Create a decoder
	decoder := json.NewDecoder(file)

	// Call Decode to decode the JSON data into the variable
	if err := decoder.Decode(&data); err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(data["active"])

	//Verify if issue is active and create incident
	if data["active"] == true {
		createIncident(data)
	}

	clearedEvents, ok := data["cleared_events"].([]interface{})
	if !ok {
		// Handle case where cleared_events is not an array
		fmt.Println("Error: cleared_events is not an array")
		return
	}

	eventId := clearedEvents[0].(string)

	incidentExists, err := getIncident(eventId)

	if err != nil {
		fmt.Printf("Error getting incident %s: %v\n", eventId, err)
		// Handle the error accordingly (e.g., log it, return, etc.)
		return
	}

	if incidentExists {
		err := closeIncident(eventId)
		if err != nil {
			fmt.Printf("Error closing incident: %v\n", err)
			// Handle the error accordingly (e.g., log it, return, etc.)
			return
		}
	}

	fmt.Println("Event ID:", eventId)
}
