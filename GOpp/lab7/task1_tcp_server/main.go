package main

import (
    "bufio"
    "fmt"
    "log"
    "net"
    "strings"
)

func main() {
    ln, err := net.Listen("tcp", ":8081")
    if err != nil {
        log.Fatal("Listen:", err)
    }
    defer ln.Close()
    fmt.Println("TCP-сервер слушает :8081")

    for {
        conn, err := ln.Accept()
        if err != nil {
            log.Println("Accept:", err)
            continue
        }
        // Каждое соединение — в своей горутине
        go handleConn(conn)
    }
}

func handleConn(conn net.Conn) {
    defer conn.Close()
    fmt.Printf("[+] Подключён %s\n", conn.RemoteAddr())

    reader := bufio.NewReader(conn)
    msg, err := reader.ReadString('\n')
    if err != nil {
        log.Println("Read:", err)
        return
    }

    fmt.Printf("Получено от %s: %q\n", conn.RemoteAddr(), strings.TrimSpace(msg))

    // Отправляем подтверждение
    fmt.Fprintf(conn, "ACK: сообщение получено (%d байт)\n", len(msg))
    fmt.Printf("[-] Отключён %s\n", conn.RemoteAddr())
}