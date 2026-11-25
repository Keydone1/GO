https://github.com/Keydone1/GO

## Структура
```
lab4/
─ go.mod
─ main.go
```

## Задания

### 1. Карта «имя → возраст»
```go
people := map[string]int{
    "Анна":   25,
    "Борис":  30,
    "Виктор": 17,
}
people["Галина"] = 42   // добавление
// вывод через for name, age := range people
```

### 2. Средний возраст
```go
func AverageAge(people map[string]int) float64
```
Защита от пустой карты (деление на ноль).

### 3. Удаление по имени
```go
func RemoveByName(people map[string]int, name string) // delete(people, name)
```

### 4. Ввод строки → верхний регистр
`bufio.Reader.ReadString('\n')` + `strings.ToUpper`.

### 5. Сумма чисел из строки
`strings.Fields` + `strconv.Atoi`, нечисловые токены пропускаются.

### 6. Массив чисел в обратном порядке
Ввод строкой → `strings.Fields` → `strconv.Atoi` → вывод в обратном порядке.

## Запуск

```bash
cd lab4
go run .
```

## Тестирование

| Задание | Ввод | Ожидание |
|---------|------|----------|
| 1–3 | — | вывод карты, средний возраст, удаление «Борис» |
| 4 | `привет, мир` | `ПРИВЕТ, МИР` |
| 5 | `10 20 30 40` | `Сумма: 100` |
| 5 | `10 abc 20` | `Пропущено (не число): "abc"`, `Сумма: 30` |
| 6 | `1 2 3 4 5` | `В обратном порядке: 5 4 3 2 1` |

## Пример вывода

```
=== Задача 1: Карта имён и возрастов ===
Анна — 25 лет
Борис — 30 лет
...
Галина — 42 лет

=== Задача 2: Средний возраст ===
Средний возраст: 28.50

=== Задача 3: Удаление записи ===
Карта после удаления Бориса: map[Анна:25 Виктор:17 Галина:42]

=== Задача 4: Строка в верхнем регистре ===
Введите строку: привет, мир
В верхнем регистре: ПРИВЕТ, МИР

=== Задача 5: Сумма чисел ===
Введите числа через пробел: 10 20 30 40
Сумма: 100

=== Задача 6: Числа в обратном порядке ===
Введите числа через пробел: 1 2 3 4 5
В обратном порядке: 5 4 3 2 1
```

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
