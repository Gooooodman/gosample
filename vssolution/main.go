package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fengbaoxp/gosample/vssolution/convertor"
)

func main() {
	prefixPtr := flag.String("prefix", "", "需转换目录前缀")
	locationPtr := flag.String("location", "", "需转换的VS解决方案根目录")
	flag.Parse()
	location := *locationPtr
	prefix := *prefixPtr

	if "" == location {
		msg := "无法获取要转换的目录，请通过参数[location]指定"
		exef, err := exec.LookPath(os.Args[0])
		if err != nil {
			fmt.Println(msg)
			return
		}
		abs, err := filepath.Abs(exef)
		if err != nil {
			fmt.Println(msg)
			return
		}
		location = filepath.Dir(abs)
	}
	fmt.Println("需要转换的VS解决方案根目录:", location)
	if "" == prefix {
		prefix = location
	}
	c, err := convertor.NewConvertor(location, prefix)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	c.RemoveCMake()
	c.ConvertorPath()
}
