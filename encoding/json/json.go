/**
  @author: $(USER)
  @data:$(DATE)
  @note:
**/
package json

import (
	"encoding/json"
	"os"
)

//目前go的json在性能上有一定欠缺(反射机制导致的),如果后期想动态选择json解析引擎，可添加至这个包

func LoadConfig(filePath string, dst interface{}) error {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	err = decoder.Decode(dst)
	if err != nil {
		return err
	}
	return nil
}

func SaveConfig(filePath string, dst interface{}) error {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.Encode(dst)
	return nil
}
