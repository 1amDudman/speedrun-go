package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type JSONData struct {
	ID     int    `json:"id"`
	Amount int    `json:"amount"`
	Status string `json:"status"`
}

func main() {
	// PGX deps
	ctx := context.Background()
	dsn := "postgres://admin:secret@localhost:5432/transactions"

	// PGX config
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatal("unable to parse database config: ", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	// Create a connection pool
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatal("error creating connection pool: ", err)
	}

	if err := pool.Ping(ctx); err != nil {
		log.Fatal("unable to ping database: ", err)
	}

	fmt.Println("Connected to database successfully!")
	_ = pool // TODO: temporary

	/* ----------------------------------------------------------------------- */

	numJobs, numWorkers := 1000, runtime.NumCPU()
	jobs := make(chan JSONData, numJobs)
	results := make(chan JSONData, numJobs)

	for i := 1; i <= numWorkers; i++ {
		go func() {
			for job := range jobs {
				if job.Status == "success" && job.Amount > 5000000 {
					results <- job
				}
			}
		}()
	}

	go func() {
		jsonData := make([]JSONData, 0, 1000)

		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case res := <-results:
				jsonData = append(jsonData, res)

				if len(jsonData) >= 1000 {
					flushBatch(jsonData, pool)
					jsonData = jsonData[:0]
				}
			case <-ticker.C:
				if len(jsonData) > 0 {
					log.Printf("Timer tick: %d records in buffer", len(jsonData))
					flushBatch(jsonData, pool)
					jsonData = jsonData[:0]
				}
			}
		}
	}()

	/* ----------------------------------------------------------------------- */

	mux := http.NewServeMux()
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		err := streamJSON(r.Body, jobs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("File successfully loaded into a queue for processing\n"))
	})

	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}

func flushBatch(jsonData []JSONData, pool *pgxpool.Pool) {
	batch := &pgx.Batch{}

	for _, data := range jsonData {
		batch.Queue("INSERT INTO successful_transactions (transaction_id, amount, status) VALUES ($1, $2, $3)",
			data.ID, data.Amount, data.Status)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	br := pool.SendBatch(ctx, batch)

	// Check for errors in batch results
	var batchErr error
	for i := 0; i < len(jsonData); i++ {
		_, err := br.Exec()
		if err != nil {
			batchErr = err
			break
		}
	}

	if err := br.Close(); err != nil {
		log.Printf("Warning: failed to close batch gracefully: %v", err)
	}

	if batchErr != nil {
		log.Printf("Error executing batch: %v", batchErr)
	} else {
		log.Printf("Batch of %d records inserted successfully", len(jsonData))
	}

	cancel()
}

func streamJSON(r io.Reader, jobs chan<- JSONData) error {
	decoder := json.NewDecoder(r)

	_, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("error reading JSON array start: %v", err)
	}

	for decoder.More() {
		var jsonData JSONData
		err := decoder.Decode(&jsonData)
		if err != nil {
			return fmt.Errorf("error decoding JSON object: %v", err)
		}
		jobs <- jsonData
	}

	_, err = decoder.Token()
	if err != nil {
		return fmt.Errorf("error reading JSON array end: %v", err)
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
