package main

import "fmt"
import "time"
import "strconv"

func main() {
    messages := make(chan string)

    go func() {
        for i := 0; i < 10; i++ {
            messages <- "ping" + strconv.Itoa(i)
            time.Sleep(time.Second)
        }
    }()

    for i := 0; i < 10; i++ {
        fmt.Println(<-messages)
    }
}
