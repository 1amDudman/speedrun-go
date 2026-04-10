//go:build ignore

package main

import (
	"fmt"
	"os"
)

func main() {
	// Создаем файл
	file, err := os.Create("heavy_data.json")
	if err != nil {
		panic(err) // Здесь паника оправдана, это разовый скрипт
	}
	defer file.Close()

	fmt.Println("Генерация тяжелого JSON началась...")

	// Открываем массив
	file.WriteString("[\n")

	// Генерируем 1 000 000 записей
	total := 1000000
	for i := 1; i <= total; i++ {
		// Имитируем разные статусы транзакций
		status := "success"
		if i%3 == 0 {
			status = "failed"
		} else if i%5 == 0 {
			status = "pending"
		}

		// Пишем JSON строку напрямую в файл
		line := fmt.Sprintf(`  {"id": %d, "amount": %d, "type": "payment", "status": "%s"}`, i, i*10, status)

		if i < total {
			line += ",\n"
		} else {
			line += "\n"
		}

		file.WriteString(line)
	}

	// Закрываем массив
	file.WriteString("]\n")

	fmt.Println("Готово! Проверь размер файла heavy_data.json")
}
