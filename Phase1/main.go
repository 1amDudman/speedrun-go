package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strconv"
)

type Transaction struct {
	ID     int    `json:"id"`
	Amount int    `json:"amount"`
	Status string `json:"status"`
}

func main() {
	// Validate command-line arguments
	if len(os.Args) < 3 {
		log.Fatalf("Usage: go run main.go <json_file.json> <csv_file.csv>")
	}

	// Get file paths from command-line arguments
	jsonFilePath := os.Args[1]
	csvFilePath := os.Args[2]

	// Open JSON file
	jsonFile, err := os.Open(jsonFilePath)
	if err != nil {
		log.Fatalln("Json file opening error: ", err)
	}
	defer jsonFile.Close()

	// Open CSV file
	csvFile, err := os.Create(csvFilePath)
	if err != nil {
		log.Fatalln("Csv file creation error: ", err)
	}
	defer csvFile.Close()

	// Create CSV writer
	csvWriter := csv.NewWriter(csvFile)
	defer csvWriter.Flush()

	fmt.Println("Starting to process JSON...")

	// Process JSON data and write to CSV
	isHeaderWritten := false
	recordCounter := 0
	err = streamJSON(jsonFile, func(t Transaction) error {
		recordCounter++
		if recordCounter%100000 == 0 {
			PrintMemUsage()
		}

		if !isHeaderWritten {
			err := csvWriter.Write([]string{"ID", "Amount", "Status"})
			if err != nil {
				return fmt.Errorf("Error writing header to CSV: %w", err)
			}
			isHeaderWritten = true
		}

		if t.Status == "success" && t.Amount > 5000000 {
			record := []string{
				strconv.Itoa(t.ID),
				strconv.Itoa(t.Amount),
				t.Status,
			}
			err := csvWriter.Write(record)
			if err != nil {
				return fmt.Errorf("Error writing to CSV: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		log.Fatalln("Error processing JSON: ", err)
	}

	fmt.Println("Done! Check the CSV file.")
}

func streamJSON(r io.Reader, process func(t Transaction) error) error {
	// Create decoder
	decoder := json.NewDecoder(r)

	// Read opening bracket
	_, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("Error reading opening bracket: %w", err)
	}

	// Parse data and pass to callback for processing
	for decoder.More() {
		var tx Transaction
		err := decoder.Decode(&tx)
		if err != nil {
			return fmt.Errorf("Error decoding JSON: %w", err)
		}

		err = process(tx)
		if err != nil {
			return fmt.Errorf("Error processing transaction: %w", err)
		}
	}

	// Read closing bracket
	_, err = decoder.Token()
	if err != nil {
		return fmt.Errorf("Error reading closing bracket: %w", err)
	}

	return nil
}

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Printf("Alloc = %v MiB\tSys = %v MiB\tNumGC = %v\n",
		m.Alloc/1024/1024,
		m.Sys/1024/1024,
		m.NumGC)
}
