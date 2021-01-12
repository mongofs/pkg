/**
  @author: $(USER)
  @data:$(DATE)
  @note:
**/
package tools

import (
	"net/url"
	"strings"
)

func HttpSplicing(baseUrl string, param map[string]string) (string, error) {
	reqUrl := baseUrl
	if !strings.Contains(reqUrl, "?") {
		reqUrl = reqUrl + "?"
	}

	var tem = ""
	for k, v := range param {
		tem = tem + k + "=" + v + "&"
	}
	reqUrl += tem
	_, err := url.Parse(reqUrl)
	if err != nil {
		return "", err
	}
	return reqUrl, nil
}
