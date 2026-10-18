package main

import (
	"bufio"
	"fmt"
	"os"
)

func main1() {
	in := bufio.NewScanner(os.Stdin)
	for in.Scan() {
		txt := in.Text()
		fmt.Println("Вы ввели:", txt)
	}
	if err := in.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка:", err)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		// Мы принудительно просим читать ровно до символа '\n'
		line, err := reader.ReadString('\n')
		if err != nil {
			break // Конец файла или ошибка
		}

		// ВНИМАНИЕ: line будет содержать символ '\n' на конце!
		// Его нужно удалять вручную, если он не нужен.
		fmt.Print("Вы ввели: ", line)
	}

}
