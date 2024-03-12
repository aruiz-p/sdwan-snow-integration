package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {

	// Open the JSON file for reading
	file, err := os.Open("./examples/intf_up.json")
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
		err := createIncident(data)
		if err != nil {
			fmt.Println("Error creating incident:", err)
			return
		}

	} else {

		if _, ok := data["cleared_events"]; ok {
			clearedEvents, ok := data["cleared_events"].([]interface{})

			if !ok {
				// Handle case where cleared_events is not an array
				fmt.Println("Error: cleared_events is not an array")
				return
			}

			eventId := clearedEvents[0].(string)

			incidentExists, incident_id, err := getIncidentWithId(eventId)

			if err != nil {
				fmt.Printf("Error getting incident %s: %v\n", eventId, err)
				return
			}

			if incidentExists {
				err := closeIncident(incident_id)
				if err != nil {
					fmt.Printf("Error closing incident: %v\n", err)
					// Handle the error accordingly (e.g., log it, return, etc.)
					return
				}
			}
		} else {
			ruleName := data["rule_name_display"].(string)
			entryTime := data["entry_time"].(float64)

			sysIp := ""

			if data["system_ip"] != nil {
				sysIp = data["system_ip"].(string)

			} else {
				devices, ok := data["devices"].([]interface{})
				if !ok {
					// Handle the case where "devices" is not of type []interface{}
					fmt.Println("devices key is not of type []interface{}")
					return
				}
				// Assuming there's at least one device in the devices slice
				device := devices[0].(map[string]interface{})
				sysIp, ok = device["system-ip"].(string)

				if !ok {
					// Handle the case where "sysIp" is not of type string
					fmt.Println("sysIp is not of type string")
					return
				}
			}

			incidentExists, incident_id, err := getIncidentWoutId(ruleName, sysIp, entryTime)

			if err != nil {
				fmt.Printf("Error getting incident %v\n", err)
				return
			}

			if incidentExists {
				err := closeIncident(incident_id)
				if err != nil {
					fmt.Printf("Error closing incident: %v\n", err)
					// Handle the error accordingly (e.g., log it, return, etc.)
					return
				}
			} else {
				fmt.Printf("Incident doesn't exist")
			}

		}

	}

}
