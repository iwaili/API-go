package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
)

func main() {
	// Load existing allowed keys
	var allowedKeys []string
	allowedKeysFile := "allowed_keys.json"

	// Check if the allowed_keys.json file exists
	if _, err := os.Stat(allowedKeysFile); os.IsNotExist(err) {
		// If the file doesn't exist, create an empty list of allowed keys
		allowedKeys = []string{}
	} else {
		// If the file exists, read and load the allowed keys
		data, err := os.ReadFile(allowedKeysFile)
		if err != nil {
			log.Fatalf("Failed to load allowed keys: %v", err)
		}
		err = json.Unmarshal(data, &allowedKeys)
		if err != nil {
			log.Fatalf("Failed to parse allowed keys: %v", err)
		}
	}

	// Get the key to add from command line arguments
	if len(os.Args) < 2 {
		log.Fatal("Please provide a key to add.")
	}
	key := os.Args[1]

	// Check if the key already exists in allowedKeys
	for _, k := range allowedKeys {
		if k == key {
			log.Fatalf("Key '%s' already exists.", key)
		}
	}

	// Add the new key to the list
	allowedKeys = append(allowedKeys, key)

	// Sort the keys in ascending order
	sort.Strings(allowedKeys)

	// Save the updated allowed keys to the JSON file
	allowedData, err := json.MarshalIndent(allowedKeys, "", "  ")
	if err != nil {
		log.Fatalf("Failed to serialize allowed keys: %v", err)
	}

	err = os.WriteFile(allowedKeysFile, allowedData, 0644)
	if err != nil {
		log.Fatalf("Failed to write allowed keys to file: %v", err)
	}

	fmt.Println("Key added successfully.")
}
