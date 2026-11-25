https://github.com/Keydone1/GO

## Структура
```
lab5/
─ go.mod
─ main.go
```

## Задания

### 1. `Person` + метод `Info()`
```go
type Person struct { Name string; Age int }
func (p Person) Info()
```

### 2. Метод `Birthday()` (pointer-receiver!)
```go
func (p *Person) Birthday() { p.Age++ }
```
 Без `*` возраст не изменится снаружи.

### 3. `Circle` + метод `Area()`
```go
type Circle struct { Radius float64 }
func (c Circle) Area() float64 // math.Pi * r^2
```

### 4. Интерфейс `Shape`
```go
type Shape interface { Area() float64 }
// реализуют и Circle, и Rectangle — неявно
type Rectangle struct { Width, Height float64 }
func (r Rectangle) Area() float64 // W * H
```

### 5. Функция над срезом `[]Shape`
```go
func PrintAreas(shapes []Shape) // полиморфный вызов s.Area()
```

### 6. Интерфейс `Stringer` + `Book`
```go
type Stringer interface { String() string }
func (b Book) String() string // fmt.Println вызовет автоматически
```

## Запуск

```bash
cd lab5
go run .
```

## Тестирование

Не требует ввода. Проверьте:

1. **Задача 1** — `Имя: Анна, Возраст: 25`.
2. **Задача 2** — после `Birthday()` возраст стал `26`. Если `25` — забыли `*` в получателе.
3. **Задача 3** — `Circle{Radius: 5}.Area() ≈ 78.54`.
4. **Задача 4** — `Rectangle{4,6} = 24`, `Circle{3} ≈ 28.27`.
5. **Задача 5** — 4 фигуры, площади перечислены.
6. **Задача 6** — книга выводится как `«Название» — Автор (год)` — и через `String()`, и через `fmt.Println(b)`.

## Пример вывода

```
=== Задача 1: Информация о человеке ===
Имя: Анна, Возраст: 25

=== Задача 2: День рождения ===
До:    Анна — 25 лет
После: Анна — 26 лет

=== Задача 3: Площадь круга ===
Радиус: 5.0, площадь: 78.54

=== Задача 4: Интерфейс Shape ===
Rectangle: 24.00
Circle:    28.27

=== Задача 5: Площади всех фигур ===
Фигура 1: площадь = 6.00
Фигура 2: площадь = 3.14
Фигура 3: площадь = 25.00
Фигура 4: площадь = 19.63

=== Задача 6: Интерфейс Stringer ===
«Преступление и наказание» — Ф. М. Достоевский (1866)
«Преступление и наказание» — Ф. М. Достоевский (1866)
```

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