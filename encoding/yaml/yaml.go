/**
  @author: $(USER)
  @data:$(DATE)
  @note:
**/
package yaml

import (
	"github.com/spf13/viper"
)

func Load(path, name, filetype string, obj interface{}) (err error) {
	yamlDecode := viper.New()
	yamlDecode.AddConfigPath(path)
	yamlDecode.SetConfigName(name)
	yamlDecode.SetConfigType(filetype)
	err = yamlDecode.ReadInConfig()
	if err != nil {
		return
	}
	env := yamlDecode.GetString("env")
	err = yamlDecode.UnmarshalKey(env, obj)
	if err != nil {
		return
	}
	return
}
