package convertor

import "testing"

// 测试工厂方法
func TestNewConvertor(t *testing.T) {
    location := "/tmp"
    _, err := NewConvertor(location)
    if err != nil {
        t.Error("目录[", location, "]存在，应该报错")
    }

    location = "/temp1"
    _, err = NewConvertor(location)
    if err == nil {
        t.Error("目录[", location, "]不存在，应该报错")
    }

    location = "/bin/bash"
    _, err = NewConvertor(location)
    if err == nil {
        t.Error("[", location, "]是个文件，应该报错")
    }
}

func TestRemoveCmake(t *testing.T) {
    temp, err := NewConvertor("/tmp")
    err = temp.RemoveCMake()
    if err != nil {
        t.Error(err.Error())
    }
}
