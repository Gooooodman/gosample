package main

import (
    "convertor"
    "fmt"
)

func main() {
    location := flag.String("location", "", "需转换的VS解决方案根目录")
    flag.Parse()

    if "" == *location {
        fmt.Println("参数[location]必须设置，并且是合法的目录。")
        return
    }

    convertor := &Convertor{*location}
    //err := convertor.ValidateLocation()
    //if nil != err {
    //    fmt.Println(err.Error())
    //    return
    //}

}
