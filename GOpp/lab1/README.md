https://github.com/Keydone1/GO
## Задания

### 1. Текущее время и дата
Выводит текущую дату и время через `time.Now()` с форматированием.

### 2. Переменные разных типов
Объявляет `int`, `float64`, `string`, `bool` через `var` и выводит значения и типы (`%T`).

### 3. Краткая форма объявления
Демонстрирует `:=` для тех же типов.

### 4. Арифметические операции с двумя целыми
Сложение, вычитание, умножение, деление, остаток (`+ - * / %`).

### 5. Функция суммы и разности float64
```go
func SumAndDiff(a, b float64) (float64, float64)
```
Возвращает два значения.

### 6. Среднее значение трёх чисел
```go
func Average3(a, b, c float64) float64
```

## Запуск

```bash
cd lab1
go run .
```

## Тестирование

Программа не требует ввода — сразу выводит результаты. Проверьте:

1. **Задание 1** — выводится актуальная дата/время.
2. **Задание 2–3** — значения и типы совпадают с ожидаемыми.
3. **Задание 4** — `10 / 3 = 3` (целочисленное деление), `10 % 3 = 1`.
4. **Задание 5** — `SumAndDiff(7.5, 2.5)` даёт `(10.0, 5.0)`.
5. **Задание 6** — `Average3(3, 6, 9) = 6.0`.

## Пример вывода

```
=== Задача 1: Текущее время и дата ===
Сейчас: 15.01.2025 14:30:45

=== Задача 2: Переменные разных типов ===
int: 42, тип: int
float64: 3.14, тип: float64
string: Go, тип: string
bool: true, тип: bool

=== Задача 4: Арифметика ===
10 + 3 = 13
10 - 3 = 7
10 * 3 = 30
10 / 3 = 3
10 % 3 = 1

=== Задача 5: Сумма и разность ===
SumAndDiff(7.5, 2.5) = (10.00, 5.00)

=== Задача 6: Среднее трёх чисел ===
Average3(3, 6, 9) = 6.00
```



package main

import (
	"fmt"
	"time"
)

// Функция для вычисления суммы и разности двух чисел с плавающей запятой
func sumAndDiff(a, b float64) (sum, diff float64) {
	sum = a + b
	diff = a - b
	return
}

// Функция для вычисления среднего значения трех чисел
func average(a, b, c float64) float64 {
	return (a + b + c) / 3
}

func main() {
	// 1. Вывод текущего времени и даты
	fmt.Println("Текущее дата и время")
	now := time.Now()
	fmt.Printf("Текущая дата и время: %v\n", now)
	fmt.Printf("Дата: %d-%02d-%02d\n", now.Year(), now.Month(), now.Day())
	fmt.Printf("Время: %02d:%02d:%02d\n\n", now.Hour(), now.Minute(), now.Second())

	  //2. Создание переменных различных типов (явное объявление)
	fmt.Println(" Переменные различных типов (явное объявление)")
	var age int = 25
	var pi float64 = 3.14159
	var name string = "Кирилл"
	var isStudent bool = false

	fmt.Printf("int: %d\n", age)
	fmt.Printf("float64: %.5f\n", pi)
	fmt.Printf("string: %s\n", name)
	fmt.Printf("bool: %t\n\n", isStudent)

	// 3. Краткая форма объявления переменных
	fmt.Println("Краткая форма объявления переменных")
	city := "Самара"
	temperature := 21.5
	dayOfWeek := 3
	isWeekend := false

	fmt.Printf("Город: %s\n", city)
	fmt.Printf("Температура: %.1f°C\n", temperature)
	fmt.Printf("День недели (число): %d\n", dayOfWeek)
	fmt.Printf("Выходной? %t\n\n", isWeekend)

	// 4. Арифметические операции с двумя целыми числами
	fmt.Println("Арифметические операции с целыми числами")
	x := 15
	y := 4

	fmt.Printf("Число 1: %d\n", x)
	fmt.Printf("Число 2: %d\n", y)
	fmt.Printf("Сумма: %d + %d = %d\n", x, y, x+y)
	fmt.Printf("Разность: %d - %d = %d\n", x, y, x-y)
	fmt.Printf("Произведение: %d * %d = %d\n", x, y, x*y)
	fmt.Printf("Частное (целочисленное): %d / %d = %d\n", x, y, x/y)
	fmt.Printf("Остаток от деления: %d %% %d = %d\n\n", x, y, x%y)

	// 5. Функция для вычисления суммы и разности чисел с плавающей запятой
	fmt.Println("Сумма и разность чисел с плавающей запятой")
	a := 15.7
	b := 4.3

	sum, diff := sumAndDiff(a, b)
	fmt.Printf("Число 1: %.2f\n", a)
	fmt.Printf("Число 2: %.2f\n", b)
	fmt.Printf("Сумма: %.2f + %.2f = %.2f\n", a, b, sum)
	fmt.Printf("Разность: %.2f - %.2f = %.2f\n\n", a, b, diff)

	// 6. Вычисление среднего значения трех чисел
	fmt.Println("Среднее значение трех чисел")
	num1 := 10.0
	num2 := 20.0
	num3 := 30.0

	avg := average(num1, num2, num3)
	fmt.Printf("Числа: %.2f, %.2f, %.2f\n", num1, num2, num3)
	fmt.Printf("Среднее значение: %.2f\n", avg)

	// Дополнительно: среднее с целыми числами
	fmt.Println("\n Среднее с целыми числами")
	n1 := 8
	n2 := 15
	n3 := 22
	avgInt := average(float64(n1), float64(n2), float64(n3))
	fmt.Printf("Целые числа: %d, %d, %d\n", n1, n2, n3)
	fmt.Printf("Среднее значение: %.2f\n", avgInt)
}