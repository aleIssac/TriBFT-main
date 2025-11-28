package build

import (
	"encoding/json"
	"fmt"
	"os"
)

func readIpTable(ipTablePath string) map[uint64]map[uint64]string {
	// Read the contents of ipTable.json
	file, err := os.ReadFile(ipTablePath)
	if err != nil {
		// handle error
		fmt.Println(err)
	}
	// Create a map to store the IP addresses
	var ipMap map[uint64]map[uint64]string
	// Unmarshal the JSON data into the map
	err = json.Unmarshal(file, &ipMap)
	if err != nil {
		// handle error
		fmt.Println(err)
	}
	return ipMap
}

// make sure the file is uptodate
var fileUpToDataSet = make(map[string]struct{})

func attachLineToFile(filePath string, line string) error {
	if _, exist := fileUpToDataSet[filePath]; !exist {
		// Try to delete old file, ignore error if file doesn't exist
		os.Remove(filePath) // Ignore error, file may not exist
		fileUpToDataSet[filePath] = struct{}{}
	}
	// Open file in append mode, create if not exists
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close() // Ensure file is closed when function ends

	// Write content to file, append a newline
	if _, err := file.WriteString(line + "\n"); err != nil {
		return err
	}

	return nil
}
