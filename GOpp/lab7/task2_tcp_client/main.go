package main

import (
    "bufio"
    "fmt"
    "log"
    "net"
    "os"
    "strings"
)

func main() {
    conn, err := net.Dial("tcp", "localhost:8081")
    if err != nil {
        log.Fatal("Dial:", err)
    }
    defer conn.Close()

    fmt.Print("Введите сообщение: ")
    reader := bufio.NewReader(os.Stdin)
    msg, _ := reader.ReadString('\n')
    msg = strings.TrimRight(msg, "\r\n")

    // Отправляем сообщение с переводом строки — сервер читает до '\n'
    fmt.Fprintf(conn, "%s\n", msg)

    // Ждём ответ
    resp, err := bufio.NewReader(conn).ReadString('\n')
    if err != nil {
        log.Fatal("Read:", err)
    }
    fmt.Println("Ответ сервера:", strings.TrimSpace(resp))
}