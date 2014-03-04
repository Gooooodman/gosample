package main

import (
    "fmt"
    "net"
    "os"
)

func main() {
    conn, err := net.Dial("tcp", "127.0.0.1:9000")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
        os.Exit(1)
    }
    defer conn.Close()

    for i := 0; i < 10; i++ {
        value := "hello world"
        if i == 9 {
            value = "exit"
        }
        _, err := conn.Write([]byte(value))
        if err != nil {
            fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
            os.Exit(1)
        }

        var buf [512]byte
        n, err := conn.Read(buf[0:])
        if err != nil {
            fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
            os.Exit(1)
        }
        fmt.Println(string(buf[0:n]))
    }
    os.Exit(0)
}
