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
