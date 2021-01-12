/**
  @author: $(USER)
  @data:$(DATE)
  @note:
**/
package json

import (
	"fmt"
	"os"
	"strconv"
	"testing"
)

var testFilePath = "test.test"

type testJson struct {
	Age  int    `json:"age"`
	Name string `json:"name"`
}

//json文件创建
func TestCreateJsonFile(t *testing.T) {
	data := []testJson{}
	for i := 0; i < 100; i++ {
		data = append(data, testJson{
			Age:  i,
			Name: strconv.Itoa(i),
		})
	}

	err := SaveConfig(testFilePath, &data)

	if err != nil {
		fmt.Println("Create JsonFile Fail:", err)
	} else {
		fmt.Println("Create JsonFile OK")
	}
}

//json文件导入
func TestLoadJsonFile(t *testing.T) {
	data := []testJson{}

	err := LoadConfig(testFilePath, &data)
	if err != nil {
		fmt.Println("Load JsonFile Fail:", err)
	} else {
		fmt.Println("Load JsonFile OK")
	}
}

//删除文件
func TestDelJsonFile(t *testing.T) {
	err := os.Remove(testFilePath)

	if err != nil {
		fmt.Println("Remove JsonFile Fail:", err)
	} else {
		fmt.Println("Remove JsonFile OK")
	}
}
