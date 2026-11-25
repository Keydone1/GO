package main

import "fmt"

// Задача 1. Структура Person с методом для вывода информации

type Person struct {
    Name string
    Age  int
}

// Info выводит информацию о человеке
func (p Person) Info() {
    fmt.Printf("Имя: %s, Возраст: %d\n", p.Name, p.Age)
}

// Задача 2. Метод Birthday увеличивает возраст на 1
// Используем указатель (*Person), чтобы изменения сохранились
func (p *Person) Birthday() {
    p.Age++
}

// Задача 3. Структура Circle с методом площади
type Circle struct {
    Radius float64
}

// Area вычисляет площадь круга
func (c Circle) Area() float64 {
    return 3.141592653589793 * c.Radius * c.Radius
}

// Задача 4. Интерфейс Shape + реализации для Rectangle и Circle
type Shape interface {
    Area() float64
}

// Rectangle уже реализует Area() float64 — этого достаточно для Shape
type Rectangle struct {
    Width  float64
    Height float64
}

func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}

// Задача 5. Функция принимает срез Shape и выводит площади
func PrintAreas(shapes []Shape) {
    for i, s := range shapes {
        fmt.Printf("Фигура %d: площадь = %.2f\n", i+1, s.Area())
    }
}


// Задача 6. Интерфейс Stringer + реализация для Book
type Stringer interface {
    String() string
}

type Book struct {
    Title  string
    Author string
    Year   int
}

// String реализует интерфейс Stringer
func (b Book) String() string {
    return fmt.Sprintf("«%s» — %s (%d)", b.Title, b.Author, b.Year)
}


// main
func main() {
    // Задача 1. Person.Info
    fmt.Println("Задача 1: Информация о человеке")
    p := Person{Name: "Кирилл", Age: 23}
    p.Info()
    fmt.Println()

    //Задача 2. Person.Birthday
    fmt.Println("Задача 2: День рождения")
    fmt.Printf("До:    %s - %d years\n", p.Name, p.Age)
    p.Birthday()
    fmt.Printf("После: %s - %d years\n", p.Name, p.Age)
    fmt.Println()

    // Задача 3. Circle.Area 
    fmt.Println("Задача 3: Площадь круга")
    c := Circle{Radius: 5}
    fmt.Printf("Радиус: %.1f, площадь: %.2f\n", c.Radius, c.Area())
    fmt.Println()

    //  Задача 4. Интерфейс Shape 
    fmt.Println("Задача 4: Интерфейс Shape")
    var s1 Shape = Rectangle{Width: 4, Height: 6}
    var s2 Shape = Circle{Radius: 3}
    fmt.Printf("Rectangle: %.2f\n", s1.Area())
    fmt.Printf("Circle:    %.2f\n", s2.Area())
    fmt.Println()

    //  Задача 5. PrintAreas для среза Shape 
    fmt.Println("Задача 5: Площади всех фигур")
    shapes := []Shape{
        Rectangle{Width: 2, Height: 3},
        Circle{Radius: 1},
        Rectangle{Width: 5, Height: 5},
        Circle{Radius: 2.5},
    }
    PrintAreas(shapes)
    fmt.Println()

    // Задача 6. Интерфейс Stringer для Book 
    fmt.Println("Задача 6: Интерфейс Stringer")
    b := Book{Title: "Каштанка", Author: "А. П. Чехов", Year: 1892}
    fmt.Println(b.String())

    
}