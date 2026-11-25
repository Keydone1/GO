https://github.com/Keydone1/GO
## Задания

### 1. Чётное или нечётное (ввод с консоли)
Пользователь вводит число один раз — программа отвечает и про чётность, и про знак (задачи 1+2 объединены).

### 2. Функция знака числа
```go
func GetSign(num int) string // "Positive" | "Negative" | "Zero"
```

### 3. Числа от 1 до 10
Классический `for i := 1; i <= 10; i++`.

### 4. Длина строки (ввод с консоли)
```go
func StringLength(s string) int // len([]rune(s)) — Unicode-корректно
```
Ввод через `bufio.Reader`, чтобы читать строку с пробелами.

### 5. Структура Rectangle
```go
type Rectangle struct { Width, Height float64 }
func (r Rectangle) Area() float64
```

### 6. Среднее двух целых
```go
func Average(a, b int) float64 // float64(a+b) / 2
```
Ввод двух чисел с консоли.

## Запуск

```bash
cd lab2
go run .
```
## Тестирование

Программа интерактивная — вводите значения:

 Шаг  Ввод Ожидание 

| Задачи 1–2 | `8` | `Число четное`, `8 — Positive` |
| | `-7` | `Число нечетное`, `-7 — Negative` |
| | `0` | `Число четное`, `0 — Zero` |
| Задача 4 | `Hello World` | `Длина строки = 11` |
| | `Привет` | `Длина строки = 6` |
| Задача 6 | `4` и `7` | `Среднее значение 4 и 7 =550` |
| | `-2` и `2` | `Среднее значение -2 и 2 = 0.00` |

## Пример вывода

```
=== Задачи 1 и 2: Четность и знак числа ===
Введите число: 8
Число четное
Число 8 — Positive

=== Задача 3: Числа от 1 до 10 ===
1
2
...
10

=== Задача 4: Длина строки ===
Введите текст: Hello World
Длина строки = 11

=== Задача 5: Площадь прямоугольника ===
Прямоугольник 5.0 x 3.0, площадь = 15.00

=== Задача 6: Среднее значение ===
Введите первое число: 4
Введите второе число: 7
Среднее значение 4 и 7 = 5.50
```

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Задача 2. Функция определения знака числа
func GetSign(num int) string {
	switch {
	case num > 0:
		return "Positive"
	case num < 0:
		return "Negative"
	default:
		return "Zero"
	}
}

// Задача 4. Функция длины строки (в Unicode-символах)
func StringLength(s string) int {
	return len([]rune(s))
}

// Задача 5. Структура Rectangle и метод для площади
type Rectangle struct {
	Width  float64
	Height float64
}

func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

// Задача 6. Среднее значение двух целых чисел
func Average(a, b int) float64 {
	return float64(a+b) / 2
}

func main() {
	// Общий reader для чтения строк целиком (с пробелами)
	reader := bufio.NewReader(os.Stdin)

	//  Задачи 1 и 2. Один ввод — два ответа
	fmt.Println("Задачи 1 и 2: Четность и знак числа")
	var num int
	fmt.Print("Введите число: ")
	fmt.Scan(&num)

	if num%2 == 0 {
		fmt.Println("Число четное")
	} else {
		fmt.Println("Число нечетное")
	}
	fmt.Printf("Число %d — %s\n", num, GetSign(num))
	fmt.Println()

	//Задача 3. Числа от 1 до 10
	fmt.Println("Задача 3: Числа от 1 до 10 ")
	for i := 1; i <= 10; i++ {
		fmt.Println(i)
	}
	fmt.Println()

	// Задача 4. Длина строки
	fmt.Println("Задача 4: Длина строки")
	reader.ReadString('\n')

	fmt.Print("Введите текст: ")
	text, _ := reader.ReadString('\n')
	text = strings.TrimRight(text, "\r\n")
	fmt.Printf("Длина строки = %d\n", StringLength(text))
	fmt.Println()

	// Задача 5. Площадь прямоугольника
	fmt.Println("Задача 5: Площадь прямоугольника")
	rect := Rectangle{Width: 5.0, Height: 3.0}
	fmt.Printf("Прямоугольник %.1f x %.1f, площадь = %.2f\n",
		rect.Width, rect.Height, rect.Area())
	fmt.Println()

	//Задача 6. Среднее значениеч
	fmt.Println("Задача 6: Среднее значение")
	var a, b int
	fmt.Print("Введите первое число: ")
	fmt.Scan(&a)
	fmt.Print("Введите второе число: ")
	fmt.Scan(&b)
	fmt.Printf("Среднее значение %d и %d = %.2f\n", a, b, Average(a, b))
}
