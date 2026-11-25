https://github.com/Keydone1/GO
## Задания

### 1–2. Пакет `mathutils` + факториал с вводом
```go
// mathutils/mathutils.go
package mathutils
func Factorial(n int) int
```
В `main` импортируется `lab3/mathutils`, пользователь вводит число — выводится факториал.

### 3. Пакет `stringutils` + переворот строки
```go
// stringutils/stringutils.go
package stringutils
func Reverse(s string) string // []rune — Unicode-корректно
```

### 4. Массив из 5 целых
```go
var arr [5]int
// заполнение: 10, 20, 30, 40, 50
```

### 5. Срез из массива + добавление/удаление
```go
slice := arr[:]
slice = append(slice, 60, 70)         // добавление
slice = append(slice[:i], slice[i+1:]...) // удаление по индексу
```

### 6. Самая длинная строка в срезе
```go
words := []string{"Go", "программирование", "кот", "разработка", "я"}
// поиск через len([]rune(w))
```

## Запуск

```bash
cd lab3
go run .
```

Если `go run .` выдаёт `package lab3/mathutils is not in std`:
```bash
go mod init lab3     # если go.mod отсутствует
go mod tidy
go run .
```

## Тестирование

| Шаг | Ввод | Ожидание |
|-----|------|----------|
| Факториал | `5` | `5! = 120` |
| Переворот | `Привет, мир` | `рим ,тевирП` |
| Массив | — | `[10 20 30 40 50]` |
| Срез + append | — | `[10 20 30 40 50 60 70]` |
| Удаление idx=2 | — | `[10 20 40 50 60 70]` |
| Длинная строка | — | `"программирование" (длина 16)` |


## Пример вывода

```
=== Факториал числа ===
Введите число для факториала: 5
5! = 120

=== Переворот строки ===
Введите строку: Привет, мир
Оригинал:   "Привет, мир"
Переворот:  "рим ,тевирП"

=== Массив из 5 целых чисел ===
arr[0] = 10
arr[1] = 20
...
arr[4] = 50

=== Срез: добавление и удаление ===
Исходный срез:             [10 20 30 40 50]
После append(60, 70):      [10 20 30 40 50 60 70]
После удаления индекса 2:  [10 20 40 50 60 70]

=== Самая длинная строка ===
Срез строк: [Go программирование кот разработка я]
Самая длинная строка: "программирование" (длина 16)
```



package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"lab3/mathutils" // путь = module + /имя_пакета
	"lab3/stringutils"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	// Задача 1+2. Факториал введённого числа (пакет mathutils)

	fmt.Println(" Факториал числа (пакет mathutils) ")
	var n int
	fmt.Print("Введите число для факториала: ")
	fmt.Scan(&n)

	if n < 0 {
		fmt.Println("Факториал для отрицательных чисел не определён")
	} else {
		fmt.Printf("%d! = %d\n", n, mathutils.Factorial(n))
	}
	fmt.Println()

	// Задача 3. Переворот строки (пакет stringutils)

	fmt.Println("Переворот строки (пакет stringutils)")
	// Сброс буфера после fmt.Scan(&n) — там остался "\n"
	reader.ReadString('\n')

	fmt.Print("Введите строку: ")
	text, _ := reader.ReadString('\n')
	text = strings.TrimRight(text, "\r\n")

	fmt.Printf("Оригинал:   %s\n", text)
	fmt.Printf("Переворот:  %s\n", stringutils.Reverse(text))
	fmt.Println()

	// Задача 4. Массив из 5 целых чисел

	fmt.Println("Массив из 5 целых чисел")
	var arr [5]int
	// Заполняем значениями (например, квадратами индекса)
	for i := 0; i < len(arr); i++ {
		arr[i] = (i + 1) * 10 // 10, 20, 30, 40, 50
	}
	// Выводим на экран
	for i, v := range arr {
		fmt.Printf("arr[%d] = %d\n", i, v)
	}
	fmt.Println()

	// Задача 5. Срез из массива: добавление и удаление элементов

	fmt.Println("Срез: добавление и удаление элементов ")
	// Создаём срез из массива arr
	slice := arr[:]
	fmt.Println("Исходный срез:      ", slice)

	// Добавление элементов через append
	slice = append(slice, 60, 70)
	fmt.Println("После append 60, 70:", slice)

	// Удаление элемента по индексу (например, индекс 2)
	idx := 2
	slice = append(slice[:idx], slice[idx+1:]...)
	fmt.Printf("После удаления индекса %d: %v\n", idx, slice)
	fmt.Println()

	// Задача 6. Срез строк — найти самую длинную

	fmt.Println("Самая длинная строка в срезе ")
	words := []string{"Go", "программирование", "молодогвардейцы", "кинетик", "якут"}
	fmt.Println("Срез строк:", words)

	longest := ""
	for _, w := range words {
		if len([]rune(w)) > len([]rune(longest)) {
			longest = w
		}
	}
	fmt.Printf("Самая длинная строка: %q (длина %d)\n",
		longest, len([]rune(longest)))
}



// Пакет mathutils содержит математические утилиты.
package mathutils

// Factorial вычисляет факториал числа n (n!).
// Для n < 0 возвращает 0, так как факториал не определён.
// Для n = 0 и n = 1 возвращает 1 (по определению 0! = 1).
func Factorial(n int) int {
	if n < 0 {
		return 0
	}
	result := 1
	for i := 2; i <= n; i++ {
		result *= i
	}
	return result
}


package stringutils

// Reverse возвращает строку, символы которой идут в обратном порядке.
// Использует []rune, чтобы корректно работать с Unicode (кириллица, эмодзи).
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
