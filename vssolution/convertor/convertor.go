package convertor

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type info struct {
	location string // 要转换的解决方案目录
	prefix   string // 转换内容依据的根目录
}

// Convertor 一个VS解决方案下项目中绝对路径转换相对路径的转换器
type Convertor struct {
	*info
}

// 工厂方法
func NewConvertor(location, prefix string) (*Convertor, error) {
	err := validateDir(location, "location")
	if err != nil {
		return nil, err
	}
	err = validateDir(prefix, "prefix")
	if err != nil {
		return nil, err
	}
	return &Convertor{&info{location, prefix}}, nil
}

func validateDir(dir, dirtype string) error {
	msg := fmt.Sprintf("参数[%s]值必须是一个存在的文件夹。", dirtype)
	file, err := os.Open(dir)
	if os.IsNotExist(err) {
		return errors.New(msg)
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return errors.New(msg)
	}

	if !fi.IsDir() {
		return errors.New(msg)
	}
	return nil
}

// 递归删除目录下所有项目文件中CMake相关内容
func (c *Convertor) RemoveCMake() error {
	return walkPath(c.location, removeCMake)
}

func (c *Convertor) ConvertorPath() error {
	return walkPath(c.location, c.convertPath)
}

func (c *Convertor) ReplaceContent() error {
	return walkPath(c.location, replace)
}

//文件处理函数原型
type processFileFunc func(filename string) error

// 递归目录
func walkPath(location string, processFileFn processFileFunc) error {
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
			err = walkPath(filepath.Join(location, fileInfo.Name()), processFileFn)
		} else {
			ext := filepath.Ext(fileInfo.Name())
			if ".vcxproj" == ext || ".filters" == ext {
				err = processFileFn(filepath.Join(location, fileInfo.Name()))
			}
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
func removeCMake(fname string) error {
	fmt.Println("CMake处理文件[" + fname + "]")
	source, err := ioutil.ReadFile(fname)
	if err != nil {
		return err
	}
	target := removeCMakeContent(source)
	err = ioutil.WriteFile(fname, target, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

// 替换文本内容
func removeCMakeContent(source []byte) []byte {
	//    fmt.Println(string(source))
	re := regexp.MustCompile(`<ItemGroup>\s*<CustomBuild Include=\".*CMakeLists.txt\">[\s|\S]*</CustomBuild>\s*</ItemGroup>`)
	target := re.ReplaceAll(source, []byte(""))
	//    fmt.Println(string(target))
	return target
}

//绝对路径转换为相对路径
func (c *Convertor) convertPath(filename string) error {
	fmt.Println("Convert处理文件[" + filename + "]")
	source, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	currentPath := filepath.Dir(filename)
	target, err := convertPathContent(c.prefix, currentPath, string(source))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filename, []byte(target), os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

//替换文本内容
func convertPathContent(rootPath, currentPath string, source string) (string, error) {
	currentPath = strings.Replace(currentPath, `\`, `/`, -1)
	newRootPath := strings.Replace(rootPath, `\`, `/`, -1)
	rootPath = strings.Replace(rootPath, `\`, `\\`, -1)
	sreg := "((" + rootPath + "|" + newRootPath + ")[^< ;]*)[< ;]"
	//    fmt.Println(sreg)
	re := regexp.MustCompile(sreg)
	var buffer bytes.Buffer
	target := re.ReplaceAllStringFunc(source, func(source string) string {
		source = strings.Replace(source, `\`, `/`, -1)
		//fmt.Println("匹配内容>", source)
		result, err := filepath.Rel(currentPath, source)
		result = strings.Replace(result, `\`, `/`, -1)
		//fmt.Println("替换后内容>", result)
		if err != nil {
			buffer.WriteString("转换[" + currentPath + "]时发生错误[" + err.Error() + "]\n")
			return source
		}
		return result
	})
	if buffer.Len() > 0 {
		return target, errors.New(buffer.String())
	}
	return target, nil
}

func replace(filename string) error {
	fmt.Println("replace处理文件[" + filename + "]")
	source, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	target := strings.Replace(string(source), "soci_core_3_2", "soci_core", -1)
	target = strings.Replace(target, "soci_mysql_3_2", "soci_mysql", -1)
	target = strings.Replace(target, "soci_sqlite3_3_2", "soci_sqlite3", -1)
	err = ioutil.WriteFile(filename, []byte(target), os.ModePerm)
	if err != nil {
		return err
	}
	return nil

}
