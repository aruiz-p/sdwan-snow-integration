package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {

	// Listen upcoming requests
	http.HandleFunc("/webhook", handleWebhook)
	fmt.Println("Server listening on port 8080...")
	if err := http.ListenAndServe("0.0.0.0:8080", nil); err != nil {
		fmt.Printf("Failed to start server: %v", err)
	}

}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	fmt.Println("Received webhook payload:", string(body))

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return
	}

	//Verify if issue is active and create incident
	if data["active"] == true {

		fmt.Println("Opening Service Now incident...")
		err := createIncident(data)
		if err != nil {
			fmt.Println("Error creating incident:", err)
			return
		}

	} else {
		// Check if request has "cleared_events" information
		if _, ok := data["cleared_events"]; ok {
			clearedEvents, ok := data["cleared_events"].([]interface{})

			if !ok {
				// Handle case where cleared_events is not an array
				fmt.Println("Error: cleared_events is not an array")
				return
			}
			// Store the uuid and call getInciddentWithId to get "sys_id"
			eventId := clearedEvents[0].(string)
			incidentExists, incident_id, err := getIncidentWithId(eventId)

			if err != nil {
				fmt.Printf("Error getting incident %s: %v\n", eventId, err)
				return
			}
			// If there is a match, close the incident with "sys_id"
			if incidentExists {
				err := closeIncident(incident_id)
				if err != nil {
					fmt.Printf("Error closing incident: %v\n", err)
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
				// Handle the case where system ip is inside "devices" key
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

			// Try and find incident without UUID reference
			incidentExists, incident_id, err := getIncidentWoutId(ruleName, sysIp, entryTime)

			if err != nil {
				fmt.Printf("Error getting incident %v\n", err)
				return
			}

			// If incident without UUID is found, close it
			if incidentExists {
				err := closeIncident(incident_id)
				if err != nil {
					fmt.Printf("Error closing incident: %v\n", err)
					return
				}
			} else {
				fmt.Printf("Incident doesn't exist or is older than 12 hours")
			}

		}

	}

}
