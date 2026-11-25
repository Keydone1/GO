package main

import (
    "bufio"
    "fmt"
    "math"
    "math/rand"
    "os"
    "strconv"
    "strings"
    "sync"
    "time"
)

// Задача 1. Три параллельные горутины
func calcFactorial(n int) int {
    result := 1
    for i := 2; i <= n; i++ {
        result *= i
    }
    return result
}

func task1() {
    fmt.Println(" Задача 1: Три параллельные горутины")

    var wg sync.WaitGroup

    // 1. Факториал
    wg.Add(1)
    go func() {
        defer wg.Done()
        fmt.Println("[Горутина 1] Считаю факториал...")
        time.Sleep(500 * time.Millisecond) // имитация задержки
        fmt.Printf("[Горутина 1] 10! = %d\n", calcFactorial(10))
    }()

    // 2. Генерация случайных чисел
    wg.Add(1)
    go func() {
        defer wg.Done()
        fmt.Println("[Горутина 2] Генерирую случайные числа...")
        time.Sleep(500 * time.Millisecond)
        nums := make([]int, 5)
        for i := range nums {
            nums[i] = rand.Intn(100)
        }
        fmt.Printf("[Горутина 2] Случайные числа: %v\n", nums)
    }()

    // 3. Сумма гармонического ряда 1 + 1/2 + 1/3 + ... + 1/n
    wg.Add(1)
    go func() {
        defer wg.Done()
        fmt.Println("[Горутина 3] Считаю сумму ряда...")
        time.Sleep(500 * time.Millisecond)
        sum := 0.0
        for i := 1; i <= 1000; i++ {
            sum += 1.0 / float64(i)
        }
        fmt.Printf("[Горутина 3] Сумма ряда (n=1000): %.4f\n", sum)
    }()

    start := time.Now()
    wg.Wait() // ждём завершения всех трёх
    fmt.Printf("Все горутины завершены за %v\n\n", time.Since(start))
}

// Задача 2. Каналы + close()
func fibonacciGenerator(n int, out chan<- int) {
    a, b := 0, 1
    for i := 0; i < n; i++ {
        out <- a
        a, b = b, a+b
    }
    close(out) // сигнал «больше данных не будет»
}

func task2() {
    fmt.Println("Задача 2: Каналы и числа Фибоначчи")

    ch := make(chan int)
    go fibonacciGenerator(10, ch)

    // range по каналу автоматически завершится после close()
    for num := range ch {
        fmt.Println("Получено:", num)
    }
    fmt.Println("Канал закрыт — чтение завершено\n")
}

// Задача 3. select с двумя каналами
func task3() {
    fmt.Println("Задача 3: select и управление каналами")

    chNumbers := make(chan int)
    chParity := make(chan string)

    // Горутина 1: случайные числа
    go func() {
        for i := 0; i < 5; i++ {
            time.Sleep(200 * time.Millisecond)
            chNumbers <- rand.Intn(100)
        }
        close(chNumbers)
    }()

    // Горутина 2: сообщения о чётности случайных чисел
    go func() {
        for i := 0; i < 5; i++ {
            time.Sleep(300 * time.Millisecond)
            n := rand.Intn(100)
            if n%2 == 0 {
                chParity <- fmt.Sprintf("Число %d — чётное", n)
            } else {
                chParity <- fmt.Sprintf("Число %d — нечётное", n)
            }
        }
        close(chParity)
    }()

    // select читает из обоих каналов, пока они не закрыты
    for chNumbers != nil || chParity != nil {
        select {
        case num, ok := <-chNumbers:
            if !ok {
                chNumbers = nil
                continue
            }
            fmt.Printf("[Числа]    %d\n", num)
        case msg, ok := <-chParity:
            if !ok {
                chParity = nil
                continue
            }
            fmt.Printf("[Чётность] %s\n", msg)
        }
    }
    fmt.Println()
}

// Задача 4. Мьютекс: гонка данных vs синхронизация
func task4() {
    fmt.Println("Задача 4: Мьютекс и гонка данных")

    const goroutines = 100
    const increments = 1000
    const expected = goroutines * increments

    //  Без мьютекса (гонка данных) 
    var counterUnsafe int
    var wg1 sync.WaitGroup
    for i := 0; i < goroutines; i++ {
        wg1.Add(1)
        go func() {
            defer wg1.Done()
            for j := 0; j < increments; j++ {
                counterUnsafe++ // ← небезопасно!
            }
        }()
    }
    wg1.Wait()
    fmt.Printf("Без мьютекса: %d (ожидалось %d)\n", counterUnsafe, expected)

    //  С мьютексом 
    var counterSafe int
    var mu sync.Mutex
    var wg2 sync.WaitGroup
    for i := 0; i < goroutines; i++ {
        wg2.Add(1)
        go func() {
            defer wg2.Done()
            for j := 0; j < increments; j++ {
                mu.Lock()
                counterSafe++
                mu.Unlock()
            }
        }()
    }
    wg2.Wait()
    fmt.Printf("С мьютексом:  %d (ожидалось %d)\n\n", counterSafe, expected)
}

// Задача 5. Многопоточный калькулятор
type CalcRequest struct {
    A, B     float64
    Op       string
    ResultCh chan float64
}

func calcServer(requests <-chan CalcRequest) {
    for req := range requests {
        var result float64
        switch req.Op {
        case "+":
            result = req.A + req.B
        case "-":
            result = req.A - req.B
        case "*":
            result = req.A * req.B
        case "/":
            if req.B == 0 {
                result = math.Inf(1)
            } else {
                result = req.A / req.B
            }
        default:
            continue
        }
        req.ResultCh <- result
    }
}

func task5() {
    fmt.Println("Задача 5: Многопоточный калькулятор")

    requests := make(chan CalcRequest)
    go calcServer(requests) // сервер в отдельной горутине

    operations := []struct {
        A, B float64
        Op   string
    }{
        {10, 5, "+"},
        {10, 5, "-"},
        {10, 5, "*"},
        {10, 5, "/"},
        {7, 2, "/"},
    }

    var wg sync.WaitGroup
    for _, op := range operations {
        wg.Add(1)
        go func(a, b float64, oper string) {
            defer wg.Done()
            resultCh := make(chan float64)
            requests <- CalcRequest{A: a, B: b, Op: oper, ResultCh: resultCh}
            result := <-resultCh
            fmt.Printf("%.1f %s %.1f = %.2f\n", a, oper, b, result)
        }(op.A, op.B, op.Op)
    }
    wg.Wait()
    close(requests)
    fmt.Println()
}

// Задача 6. Пул воркеров
func reverseString(s string) string {
    runes := []rune(s)
    for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
        runes[i], runes[j] = runes[j], runes[i]
    }
    return string(runes)
}

type Job struct {
    ID   int
    Text string
}

type JobResult struct {
    JobID    int
    Original string
    Reversed string
    WorkerID int
}

func worker(id int, jobs <-chan Job, results chan<- JobResult, wg *sync.WaitGroup) {
    defer wg.Done()
    for job := range jobs {
        time.Sleep(100 * time.Millisecond) // имитация работы
        results <- JobResult{
            JobID:    job.ID,
            Original: job.Text,
            Reversed: reverseString(job.Text),
            WorkerID: id,
        }
    }
}

func task6() {
    fmt.Println("Задача 6: Пул воркеров")

    // Ввод количества воркеров
    reader := bufio.NewReader(os.Stdin)
    numWorkers := 3
    fmt.Print("Введите количество воркеров (Enter — 3): ")
    line, _ := reader.ReadString('\n')
    line = strings.TrimSpace(line)
    if line != "" {
        if n, err := strconv.Atoi(line); err == nil && n > 0 {
            numWorkers = n
        }
    }

    tasks := []string{
        "hello world",
        "goal",
        "ronaldo",
        "barsa pozorniki",
        "mutex",
        "worker pool",
        "привет",
    }

    jobs := make(chan Job, len(tasks))
    results := make(chan JobResult, len(tasks))

    // Запускаем воркеров
    var wg sync.WaitGroup
    for i := 1; i <= numWorkers; i++ {
        wg.Add(1)
        go worker(i, jobs, results, &wg)
    }

    // Отправляем задачи
    for i, t := range tasks {
        jobs <- Job{ID: i + 1, Text: t}
    }
    close(jobs)

    // Отдельная горутина закрывает results после завершения всех воркеров
    go func() {
        wg.Wait()
        close(results)
    }()

    // Читаем результаты
    fmt.Printf("\nРаботают %d воркеров:\n", numWorkers)
    for res := range results {
        fmt.Printf("Воркер %d | Задача %d: %q → %q\n",
            res.WorkerID, res.JobID, res.Original, res.Reversed)
    }
    fmt.Println()
}

// main
func main() {
    task1()
    task2()
    task3()
    task4()
    task5()
    task6()
}