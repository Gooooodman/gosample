package main

import (
    "flag"
    "fmt"
    "github.com/fengbaoxp/gosample/vssolution/convertor"
)

func main() {
    location := flag.String("location", "", "需转换的VS解决方案根目录")
    flag.Parse()

    if "" == *location {
        fmt.Println("参数[location]必须设置，并且是合法的目录。")
        return
    }

    c, err := convertor.NewConvertor(*location)
    if err != nil {
        fmt.Println(err.Error())
        return
    }
    c.RemoveCMake()
}
