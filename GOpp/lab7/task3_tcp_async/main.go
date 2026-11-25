package main

import (
    "bufio"
    "context"
    "fmt"
    "log"
    "net"
    "os"
    "os/signal"
    "strings"
    "sync"
    "sync/atomic"
    "syscall"
    "time"
)

func main() {
    ln, err := net.Listen("tcp", ":8082")
    if err != nil {
        log.Fatal("Listen:", err)
    }
    fmt.Println("Асинхронный TCP-сервер слушает :8082")
    fmt.Println("Для остановки нажмите Ctrl+C")

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    var wg sync.WaitGroup
    var activeConnections int64

    // ---- Accept loop ----
    go func() {
        for {
            conn, err := ln.Accept()
            if err != nil {
                select {
                case <-ctx.Done():
                    return // ожидаемое завершение после Close()
                default:
                    log.Println("Accept:", err)
                    continue
                }
            }
            atomic.AddInt64(&activeConnections, 1)
            wg.Add(1)
            go handleConn(ctx, conn, &wg, &activeConnections)
        }
    }()

    // ---- Ожидание сигнала ----
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    <-sigCh
    log.Println("Получен сигнал, начинаем graceful shutdown...")

    // Перестаём принимать новые соединения
    cancel()
    ln.Close()

    // Ждём завершения активных соединений с таймаутом
    done := make(chan struct{})
    go func() {
        wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        log.Println("Все соединения закрыты штатно")
    case <-time.After(5 * time.Second):
        log.Printf("Таймаут! Осталось активных соединений: %d\n",
            atomic.LoadInt64(&activeConnections))
    }
}

func handleConn(ctx context.Context, conn net.Conn, wg *sync.WaitGroup, active *int64) {
    defer wg.Done()
    defer conn.Close()
    defer atomic.AddInt64(active, -1)

    remote := conn.RemoteAddr().String()
    log.Printf("[+] %s подключён", remote)

    reader := bufio.NewReader(conn)
    for {
        // Проверяем отмену контекста — если shutdown, выходим
        select {
        case <-ctx.Done():
            fmt.Fprintf(conn, "Сервер выключается, до свидания!\n")
            return
        default:
        }

        // Таймаут на чтение, чтобы не блокироваться навсегда
        conn.SetReadDeadline(time.Now().Add(30 * time.Second))
        msg, err := reader.ReadString('\n')
        if err != nil {
            log.Printf("[-] %s отключён: %v", remote, err)
            return
        }

        msg = strings.TrimSpace(msg)
        if msg == "quit" {
            fmt.Fprintf(conn, "До свидания!\n")
            return
        }
        log.Printf("[%s] %s", remote, msg)
        fmt.Fprintf(conn, "ACK: %s (%d символов)\n", msg, len([]rune(msg)))
    }
}