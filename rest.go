// Rest API Implementations

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
)

// restWakeUpWithComputerName - REST Handler for Processing URLS /api/computer/<computerName>
func restWakeUpWithComputerName(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	computerName := vars["computerName"]

	var result WakeUpResponseObject
	result.Success = false

	// Ensure computerName is not empty
	if computerName == "" {
		result.Message = "Empty Computername is not allowed"
		result.ErrorObject = nil
		w.WriteHeader(http.StatusBadRequest)
	} else {
		// Get Computer from List
		for _, c := range ComputerList {
			if c.Name == computerName {
				// We found the Computername
				if err := SendMagicPacket(c.Mac, c.BroadcastIPAddress, ""); err != nil {
					// We got an internal Error on SendMagicPacket
					w.WriteHeader(http.StatusInternalServerError)
					result.Success = false
					result.Message = "Internal error on Sending the Magic Packet"
					result.ErrorObject = err
				} else {
					// Horray we send the WOL Packet succesfully
					result.Success = true
					result.Message = fmt.Sprintf("Succesfully Wakeup Computer %s with Mac %s on Broadcast IP %s", c.Name, c.Mac, c.BroadcastIPAddress)
					result.ErrorObject = nil
				}
			}
		}

		if result.Success == false && result.ErrorObject == nil {
			// We could not find the Computername
			w.WriteHeader(http.StatusNotFound)
			result.Message = fmt.Sprintf("Computername %s could not be found", computerName)
		}
	}
	json.NewEncoder(w).Encode(result)
}

// restAddComputer - REST Handler for adding a new computer
func restAddComputer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var newComputer Computer
	if err := json.NewDecoder(r.Body).Decode(&newComputer); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate MAC and IP address formats
	macRegex := regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`)
	ipRegex := regexp.MustCompile(`^([0-9]{1,3}\.){3}[0-9]{1,3}$`)
	if !macRegex.MatchString(newComputer.Mac) {
		http.Error(w, "Invalid MAC address format", http.StatusBadRequest)
		return
	}
	if !ipRegex.MatchString(newComputer.BroadcastIPAddress) {
		http.Error(w, "Invalid IP address format", http.StatusBadRequest)
		return
	}

	// Check for duplicate name, MAC, or IP address
	for _, c := range ComputerList {
		if c.Name == newComputer.Name {
			http.Error(w, "Computer with this name already exists", http.StatusBadRequest)
			return
		}
		if c.Mac == newComputer.Mac {
			http.Error(w, "Computer with this MAC address already exists", http.StatusBadRequest)
			return
		}
		if c.BroadcastIPAddress == newComputer.BroadcastIPAddress {
			http.Error(w, "Computer with this IP address already exists", http.StatusBadRequest)
			return
		}
	}

	// Append the new computer to the list
	ComputerList = append(ComputerList, newComputer)

	// Save the updated list to the CSV file
	if err := saveComputerList(DefaultComputerFilePath, ComputerList); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Log the creation of a new computer
	log.Printf("Added new computer: %+v\n", newComputer)

	// Return success message
	response := WakeUpResponseObject{
		Success: true,
		Message: "Computer added successfully",
		ErrorObject: nil,
	}
	json.NewEncoder(w).Encode(response)
}

// restDeleteComputer - REST Handler for deleting a computer
func restDeleteComputer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	computerName := vars["computerName"]

	index := -1
	for i, c := range ComputerList {
		if c.Name == computerName {
			index = i
			break
		}
	}

	if index == -1 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(WakeUpResponseObject{
			Success: false,
			Message: fmt.Sprintf("Computername %s could not be found", computerName),
			ErrorObject: nil,
		})
		return
	}

	// Delete computer
	ComputerList = append(ComputerList[:index], ComputerList[index+1:]...)
	if err := saveComputerList(DefaultComputerFilePath, ComputerList); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Log the deletion of a computer
	log.Printf("Deleted computer: %s\n", computerName)

	// Return success message
	json.NewEncoder(w).Encode(WakeUpResponseObject{
		Success: true,
		Message: fmt.Sprintf("Successfully deleted computer %s", computerName),
		ErrorObject: nil,
	})
}
