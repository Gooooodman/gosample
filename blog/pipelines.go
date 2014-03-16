package main

import (
    "fmt"
    "sync"
    "time"
)

func gen(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        for _, n := range nums {
            out <- n
        }
        close(out)
    }()
    return out
}

func sq(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        for n := range in {
            out <- n * n
        }
        close(out)
    }()
    return out
}

func merge(cs ...<-chan int) <-chan int {
    var wg sync.WaitGroup
    out := make(chan int)

    // Start ant output goroutine for each input channel in cs. output
    // copies values from c to out until c is closed, then calls wg.Done.
    output := func(c <-chan int) {
        for n := range c {
            out <- n
        }
        wg.Done()
    }
    wg.Add(len(cs))
    for _, c := range cs {
        go output(c)
    }

    // Start a goroutine to close out once all the output goroutines are
    // done. This must start after the wg.Add call.
    go func() {
        wg.Wait()
        close(out)
    }()
    return out
}

func main() {
    c := gen(2, 3)
    out := sq(c)
    fmt.Println("第一种实现:")
    fmt.Println(<-out)
    fmt.Println(<-out)

    fmt.Println("第二种实现:")
    for n := range sq(sq(gen(2, 3))) {
        fmt.Println(n)
    }

    fmt.Println("第三种实现:")
    in := gen(2, 3)
    c1 := sq(in)
    c2 := sq(in)
    /*
       for n1 := range c1 {
           fmt.Println("1_", n1)
       }
       for n2 := range c2 {
           fmt.Println("2_", n2)
       }
    */
    for n := range merge(c1, c2) {
        fmt.Println(n)
    }
}
