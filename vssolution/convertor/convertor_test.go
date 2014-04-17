package convertor

import (
    "strings"
    "testing"
)

// 测试工厂方法
func TestNewConvertor(t *testing.T) {
    location := "/tmp"
    _, err := NewConvertor(location, location)
    if err != nil {
        t.Error("目录[", location, "]存在，应该报错")
    }

    location = "/temp1"
    _, err = NewConvertor(location, location)
    if err == nil {
        t.Error("目录[", location, "]不存在，应该报错")
    }

    location = "/bin/bash"
    _, err = NewConvertor(location, location)
    if err == nil {
        t.Error("[", location, "]是个文件，应该报错")
    }
}

// 测试清除CMake内容
func TestRemoveCMakeContent(t *testing.T) {
    oldContent := `
  <ItemGroup>
    <CustomBuild Include="..\..\core\CMakeLists.txt">
    </CustomBuild>
  </ItemGroup>
    `
    result := string(removeCMakeContent([]byte(oldContent)))
    result = strings.TrimSpace(result)
    if "" != result {
        t.Error("期望[", "", "],实际[", result, "]")
    }
}

// 测试替换内容
func TestConvertPathContent(t *testing.T) {
    rootPath := `\a\b`
    currentPath := "/a/b/e/f"
    oldContent := `sss /a/b/s1/s2 /a/b/c/d;xxxxx/a/b/s1/s2<\a\b\s1\s2 zzz`
    expected := `sss ../../s1/s2 ../../c/d;xxxxx../../s1/s2<../../s1/s2 zzz`
    actual, err := convertPathContent(rootPath, currentPath, oldContent)
    if err != nil {
        t.Error(err.Error())
    }
    if expected != actual {
        t.Error("linux期望[", expected, "],实际[", actual, "]")
    }
    rootPath = `D:\test`
    currentPath = "D:/test/e/f"
    oldContent = `ssD:\test\s1\s2 D:/test/c/d;D:\test\s1\s2<zzz`
    expected = `ss../../s1/s2 ../../c/d;../../s1/s2<zzz`
    actual, err = convertPathContent(rootPath, currentPath, oldContent)
    if err != nil {
        t.Error(err.Error())
    }
    if expected != actual {
        t.Error("window期望[", expected, "],实际[", actual, "]")
    }
}
