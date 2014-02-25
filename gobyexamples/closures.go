package main

import "fmt"

func intSqe() func() int {
    i := 0
    return func() int {
        i += 1
        return i
    }
}

func main() {
    nextInt := intSqe()

    fmt.Println(nextInt())
    fmt.Println(nextInt())

    newInts := intSqe()
    fmt.Println(newInts())
}
