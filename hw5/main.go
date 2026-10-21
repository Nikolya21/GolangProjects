package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func collectServerFile(f *os.File) (*os.File, error) {
	reader := bufio.NewReader(f)
	file, err1 := os.Create("ResultFile")
	if err1 != nil {
		return nil, fmt.Errorf("ошибка создания файла: %w", err1)
	}
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			lineS := strings.Fields(line)
			switch lineS[1] {
			case "[INFO]":
				fmt.Fprintln(file, 1)
			case "[ERROR]":
				fmt.Fprintln(file, 2)
			case "[WARN]":
				fmt.Fprintln(file, 3)
			}
		}
		if err == io.EOF {
			fmt.Fprintln(file, "Строк больше нет, файл завершен.")
			return file, err1
		}
	}
}

// Проблема в том, что мы каждый раз идем туда сюда в оперативку - долго.
// нужно воспользоваться NewWrite

func collectServerFile1(f *os.File) (*os.File, error) {
	file, err := os.Create("ResultFile1")
	if err != nil {
		return nil, fmt.Errorf("ошибка создания файла: %w", err)
	}
	scanner := bufio.NewScanner(f)
	writer := bufio.NewWriter(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		switch {
		case strings.Contains(line, "[INFO]"):
			writer.WriteString("1\n")
		case strings.Contains(line, "[ERROR]"):
			writer.WriteString("2\n")
		case strings.Contains(line, "[WARN]"):
			writer.WriteString("3\n")
		}
	}
	if err := scanner.Err(); err != nil {
		file.Close()
		return nil, fmt.Errorf("ошибка при чтении: %w", err)
	}
	if err := writer.Flush(); err != nil {
		file.Close()
		return nil, fmt.Errorf("ошибка при сбросе буфера записи: %w", err)
	}
	return file, nil
}

func main() {
	file, err := os.Open("/Users/nikola2104/Desktop/MiniProj/GolangProjects/hw5/server.log")
	if err != nil {
		fmt.Println("Ошибка открытия файла:", err)
		return
	}
	collectServerFile1(file)
}
