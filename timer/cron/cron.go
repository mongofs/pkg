/**
  @author: $(USER)
  @data:$(DATE)
  @note:
**/
package cron

import (
	robfig_cron "github.com/robfig/cron"
)

//开源库cron的封装  后期可以自定义一些方法
type MyCron struct {
	*robfig_cron.Cron
}

func NewCron() *MyCron {
	return &MyCron{Cron: robfig_cron.New()}
}

func (c *MyCron) AddFunc(spec string, f func()) error {
	return c.Cron.AddFunc(spec, f)
}

func (c *MyCron) Start() {
	if c.Cron != nil {
		c.Cron.Start()
	}
}

func (c *MyCron) Stop() {
	if c.Cron != nil {
		c.Cron.Stop()
	}
}
