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
	otypePtr := flag.Int("type", 1, "操作类型")
	oldPtr := flag.String("old", "", "被替换字符串")
	newPtr := flag.String("new", "", "替换的字符串")
	flag.Parse()

	location := *locationPtr
	prefix := *prefixPtr
	otype := *otypePtr
	olds := *oldPtr
	news := *newPtr

	if "" == prefix {
		fmt.Println("必须指定需转换目录前缀，格式: -prefix=XXX。")
		return
	}
	if "" == location {
		msg := "要转换的目录非法，请核对-location参数值。"
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
	if 3 == otype && "" == olds {
		fmt.Println("必须指定被替换的字符串, 格式: -old=XXX。")
		return
	}
	fmt.Println("要转换的目录:", location)
	if "" == prefix {
		prefix = location
	}
	c, err := convertor.NewConvertor(location, prefix)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	switch otype {
	case 1:
		c.ConvertorPath()
	case 2:
		c.RemoveCMake()
	case 3:
		c.ReplaceContent(olds, news)
	}
}
