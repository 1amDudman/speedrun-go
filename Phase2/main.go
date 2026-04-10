package main

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strconv"
	"sync"
)

type Transaction struct {
	ID     int    `json:"id"`
	Amount int    `json:"amount"`
	Status string `json:"status"`
}

func main() {
	// Validate command-line arguments
	if len(os.Args) < 3 {
		log.Fatalln("Usage: go run main.go <json_file.json> <csv_file.csv>")
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

	numJobs, numWorkers := 100, runtime.NumCPU()
	jobs := make(chan Transaction, numJobs)
	results := make(chan Transaction, numJobs)

	var wg sync.WaitGroup
	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for job := range jobs {
				if job.Status == "success" && job.Amount > 5000000 {
					// Simulate CPU-intensive processing by performing multiple SHA256 computations
					dataToHash := []byte(strconv.Itoa(job.ID))
					for i := 0; i < 1000; i++ {
						_ = sha256.Sum256(dataToHash)
					}

					results <- job
				}
			}
		}(i)
	}

	go func() {
		err = streamJSON(jsonFile, jobs)
		if err != nil {
			log.Println("Error processing JSON: ", err)
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	// Create CSV writer
	csvWriter := csv.NewWriter(csvFile)
	defer csvWriter.Flush()

	isHeaderWritten := false
	for res := range results {
		// Write header if not written yet
		if !isHeaderWritten {
			err := csvWriter.Write([]string{"ID", "Amount", "Status"})
			if err != nil {
				log.Fatalf("Error writing header to CSV: %v", err)
			}
			isHeaderWritten = true
		}

		record := []string{
			strconv.Itoa(res.ID),
			strconv.Itoa(res.Amount),
			res.Status,
		}
		err := csvWriter.Write(record)
		if err != nil {
			log.Fatalf("Error writing to CSV: %v", err)
		}
	}

	fmt.Println("Done! Check the CSV file.")
}

func streamJSON(r io.Reader, jobs chan<- Transaction) error {
	defer close(jobs)

	// Create a JSON decoder
	decoder := json.NewDecoder(r)

	// Read the opening bracket of the JSON array
	_, err := decoder.Token()
	if err != nil {
		return err // : handle error
	}

	recordCounter := 0
	for decoder.More() {
		var t Transaction
		err := decoder.Decode(&t)
		if err != nil {
			return err
		}

		// Record counter for logging memory usage
		recordCounter++
		if recordCounter%100000 == 0 {
			PrintMemUsage()
		}

		jobs <- t
	}

	return nil
}

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Alloc - это сколько памяти выделено и еще не очищено сборщиком мусора (наш главный показатель)
	// Sys - это сколько памяти программа суммарно запросила у операционной системы
	fmt.Printf("Alloc = %v MiB\tSys = %v MiB\tNumGC = %v\n",
		m.Alloc/1024/1024,
		m.Sys/1024/1024,
		m.NumGC)
}
