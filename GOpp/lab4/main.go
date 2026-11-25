package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Задача 2. Функция: средний возраст по карте
func AverageAge(people map[string]int) float64 {
	if len(people) == 0 {
		return 0
	}
	sum := 0
	for _, age := range people {
		sum += age
	}
	return float64(sum) / float64(len(people))
}

// Задача 3. Функция: удалить запись из карты по имени
func RemoveByName(people map[string]int, name string) {
	delete(people, name)
}

// main
func main() {
	reader := bufio.NewReader(os.Stdin)

	// Задача 1. Карта: имена и возрасты
	fmt.Println(" Задача 1: Карта имён и возрастов ")
	people := map[string]int{
		"Анна":   25,
		"Борис":  30,
		"Виктор": 17,
	}

	// Добавляем нового человека
	people["Галина"] = 42

	// Выводим все записи
	for name, age := range people {
		fmt.Printf("%s — %d лет\n", name, age)
	}
	fmt.Println()

	// Задача 2. Средний возраст
	fmt.Println(" Задача 2: Средний возраст ")
	fmt.Printf("Средний возраст: %.2f\n", AverageAge(people))
	fmt.Println()

	// Задача 3. Удаление записи по имени
	fmt.Println(" Задача 3: Удаление записи по имени")
	fmt.Println("Карта до удаления:", people)
	RemoveByName(people, "Борис")
	fmt.Println("Карта после удаления Бориса:", people)
	fmt.Println()

	// Задача 4. Ввод строки  ВЕРХНИЙ РЕГИСТР
	fmt.Println("Задача 4: Строка в верхнем регистре ")
	fmt.Print("Введите строку: ")
	line, _ := reader.ReadString('\n')
	line = strings.TrimRight(line, "\r\n")
	fmt.Printf("В верхнем регистре: %s\n", strings.ToUpper(line))
	fmt.Println()

	// Задача 5. Сумма нескольких чисел
	fmt.Println("Задача 5: Сумма чисел")
	fmt.Print("Введите числа через пробел: ")
	numbersLine, _ := reader.ReadString('\n')
	numbersLine = strings.TrimSpace(numbersLine)

	fields := strings.Fields(numbersLine)
	sum := 0
	for _, f := range fields {
		num, err := strconv.Atoi(f)
		if err != nil {
			fmt.Printf("Пропущено (не число): %q\n", f)
			continue
		}
		sum += num
	}
	fmt.Printf("Сумма: %d\n", sum)
	fmt.Println()

	// Задача 6. Массив целых чисел в обратном порядке
	fmt.Println("Задача 6: Числа в обратном порядке")
	fmt.Print("Введите числа через пробел: ")
	reverseLine, _ := reader.ReadString('\n')
	reverseLine = strings.TrimSpace(reverseLine)

	reverseFields := strings.Fields(reverseLine)
	numbers := make([]int, 0, len(reverseFields))
	for _, f := range reverseFields {
		num, err := strconv.Atoi(f)
		if err != nil {
			fmt.Printf("Пропущено (не число): %q\n", f)
			continue
		}
		numbers = append(numbers, num)
	}

	fmt.Println("Исходный массив:", numbers)
	fmt.Print("В обратном порядке: ")
	for i := len(numbers) - 1; i >= 0; i-- {
		fmt.Printf("%d ", numbers[i])
	}
	fmt.Println()
}
