package main

import (
    "fmt"
    "time"
)

func worker(id int, jobs <-chan int, results chan<- int) {
    for i := range jobs {
        fmt.Println("worker", id, "processing job", i)
        time.Sleep(time.Second)
        results <- i * 2
    }
}

func main() {

    jobs := make(chan int, 100)
    results := make(chan int, 100)

    for w := 1; w <= 3; w++ {
        go worker(w, jobs, results)
    }

    for i := 1; i <= 9; i++ {
        jobs <- i
    }
    close(jobs)

    for a := 1; a <= 9; a++ {
        <-results
    }
}
