package main

import (
    "fmt"
    "math/rand"
    "runtime"
    "time"
)

func makeBuffer() []byte {
    return make([]byte, rand.Intn(5000000)+50000000)
}

func main() {
    pool := make([][]byte, 20)

    buffer := make(chan []byte, 5)

    fmt.Println("向操作系统申请字节|实际占用字节|当前堆中分配字节|当前堆中未使用|向操作系统返回字节|运行次数")
    var m runtime.MemStats
    makes := 0
    for {
        var b []byte
        select {
        case b = <-buffer:
        default:
            makes += 1
            b = makeBuffer()
        }

        i := rand.Intn(len(pool))
        if pool[i] != nil {
            select {
            case buffer <- pool[i]:
                pool[i] = nil
            default:
            }
        }

        pool[i] = b

        time.Sleep(time.Second)

        bytes := 0

        for i := 0; i < len(pool); i++ {
            if pool[i] != nil {
                bytes += len(pool[i])
            }
        }

        runtime.ReadMemStats(&m)
        fmt.Printf("%dM, %dM, %dM, %dM, %dM, %d\n", m.HeapSys/1024/1024, bytes/1024/1024,
            m.HeapAlloc/1024/1024, m.HeapIdle/1024/1024, m.HeapReleased/1024/1024, makes)
    }
}
