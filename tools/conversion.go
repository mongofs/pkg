/**
  @author: $(USER)
  @data:$(DATE)
  @note:
**/
package tools

import (
	"strconv"
	"time"
)

var (
	toBeCharge   = "2006-01-02T15:04:05+08:00"
	toBeChargeUT = "2006-01-02T15:04:05+00:00"
	toBeCharge2  = "2006-01-02 15:04:05"
	timeLayout   = "2006-01-02"
	timeLayoutDt = "2006-01-02T15:04:05+08:00"
	loc, _       = time.LoadLocation("Asia/Shanghai") //设置时区
)

//1 go不支持隐式类型转换 这里提供一些封装函数
//2 标准库转换不够简洁，如string转int  有两个参数，使得代码不够简洁，这里一个封装，但必须确保数据是可预知的不出错误的值

func StringToInt(s string) int {
	value, _ := strconv.Atoi(s)
	return value
}

func StringToInt64(s string) int64 {
	value, _ := strconv.ParseInt(s, 10, 64)
	return value
}

func StringToFloat64(s string) float64 {
	value, _ := strconv.ParseFloat(s, 10)
	return value
}

func StringToBool(s string) bool {
	value, _ := strconv.ParseBool(s)
	return value
}

//  yyyy-mm-dd  hh:mm:ss 格式的时间转换为TTL时间
func StringTimeToInt(str string) int {
	te, _ := time.ParseInLocation(toBeCharge2, str, loc)
	return int(te.Unix())
}

//TTL时间转Time YYYY-MM--HH
func TTLToYMD(t int64) time.Time {
	// go语言固定日期模版
	timeLayout := "2006-01-02"
	// time.Unix的第二个参数传递0或10结果一样，因为都不大于1e9
	timeStr := time.Unix(t, 0).Format(timeLayout)
	st, _ := time.Parse(timeLayout, timeStr) //string转time
	return st
}

//TTL时间转Time YYYY-MM--HH MM:SS
func TTLToDateTime(t int64) time.Time {
	// go语言固定日期模版
	timeLayout := "2006-01-02 12:24"
	// time.Unix的第二个参数传递0或10结果一样，因为都不大于1e9
	timeStr := time.Unix(t, 0).Format(timeLayout)
	st, _ := time.Parse(timeLayout, timeStr) //string转time
	return st
}

//string转上海时间Time
func StringToTimeYMD(t string) time.Time {
	timeLayout := "2006-01-02"
	loc, _ := time.LoadLocation("Asia/Shanghai")
	ymd, _ := time.ParseInLocation(timeLayout, t, loc)
	return ymd
}
