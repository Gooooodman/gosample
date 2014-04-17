package convertor

import (
    "bytes"
    "errors"
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "regexp"
)

type info struct {
    location string // 要转换的根目录
}

// Convertor 一个VS解决方案下项目中绝对路径转换相对路径的转换器
type Convertor struct {
    *info
}

// 工厂方法
func NewConvertor(location string) (*Convertor, error) {
    file, err := os.Open(location)
    msg := "参数[location]值必须是一个存在的文件夹。"
    if os.IsNotExist(err) {
        return nil, errors.New(msg)
    }
    defer file.Close()

    fi, err := file.Stat()
    if err != nil {
        return nil, errors.New(msg)
    }

    if !fi.IsDir() {
        return nil, errors.New(msg)
    }
    return &Convertor{&info{location}}, nil
}

// 返回location值
func (c *Convertor) Location() string {
    return c.location
}

// 递归删除目录下所有项目文件中CMake相关内容
func (c *Convertor) RemoveCMake() error {
    return removeDirCMake(c.location)
}

// 递归删除目录下所有文件中的CMake相关内容
func removeDirCMake(location string) error {
    fmt.Println("处理路径[" + location + "]")
    f, err := os.Open(location)
    if err != nil {
        return errors.New("打开目录[" + location + "]时发生错误[" + err.Error() + "]\n")
    }
    var buffer bytes.Buffer
    list, err := f.Readdir(-1)
    f.Close()
    if err != nil {
        buffer.WriteString("浏览目录[" + location + "]时发生错误[" + err.Error() + "]\n")
    }
    for _, fileInfo := range list {
        err = nil
        if fileInfo.IsDir() {
            err = removeDirCMake(filepath.Join(location, fileInfo.Name()))
        } else if ".vcxproj" == filepath.Ext(fileInfo.Name()) {
            err = removeOneCMake(filepath.Join(location, fileInfo.Name()))
        }
        if err != nil {
            buffer.WriteString(err.Error())
        }
    }
    if buffer.Len() > 0 {
        return errors.New(buffer.String())
    }
    return nil
}

// 删除一个文件的Cmake相关内容
func removeOneCMake(fname string) error {
    fmt.Println("处理文件[" + fname + "]")
    oldContent, err := ioutil.ReadFile(fname)
    if err != nil {
        return err
    }
    //    fmt.Println(string(oldContent))
    re := regexp.MustCompile(`<ItemGroup>\s*<CustomBuild Include=\".*CMakeLists.txt\">[\s|\S]*</CustomBuild>\s*</ItemGroup>`)
    newContent := re.ReplaceAll(oldContent, []byte(""))
    //    fmt.Println(string(newContent))
    err = ioutil.WriteFile(fname, newContent, os.ModePerm)
    if err != nil {
        return err
    }
    return nil
}
