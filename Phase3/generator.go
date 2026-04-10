//go:build ignore

package main

import (
	"fmt"
	"os"
)

func main() {
	file, err := os.Create("heavy_data.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	fmt.Println("Heavy JSON generation started...")

	file.WriteString("[\n")

	total := 1000000
	for i := 1; i <= total; i++ {
		status := "success"
		if i%3 == 0 {
			status = "failed"
		} else if i%5 == 0 {
			status = "pending"
		}

		line := fmt.Sprintf(`  {"id": %d, "amount": %d, "type": "payment", "status": "%s"}`, i, i*10, status)

		if i < total {
			line += ",\n"
		} else {
			line += "\n"
		}

		file.WriteString(line)
	}

	file.WriteString("]\n")

	fmt.Println("Ready! Check size of heavy_data.json")
}
