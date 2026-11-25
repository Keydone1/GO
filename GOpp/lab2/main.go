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
