/**
  @author: $(USER)
  @data:$(DATE)
  @note:
**/
package jeager

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"xorm.io/builder"
	"xorm.io/xorm/contexts"
)

//jeager分布式追踪的xorm日志相关的钩子函数
type TracingHook struct {
	before func(c *contexts.ContextHook) (context.Context, error)
	after  func(c *contexts.ContextHook) error
}

func (h *TracingHook) BeforeProcess(c *contexts.ContextHook) (context.Context, error) {
	return h.before(c)
}

func (h *TracingHook) AfterProcess(c *contexts.ContextHook) error {
	return h.after(c)
}

// 让编译器知道这个是xorm的Hook，防止编译器无法检查到异常
var _ contexts.Hook = &TracingHook{}

type xormHookSpan struct{}

var xormHookSpanKey = &xormHookSpan{}

func before(c *contexts.ContextHook) (context.Context, error) {

	span, _ := opentracing.StartSpanFromContext(c.Ctx, "xorm sql execute")
	c.Ctx = context.WithValue(c.Ctx, xormHookSpanKey, span)
	return c.Ctx, nil
}

func after(c *contexts.ContextHook) error {
	// 自己实现opentracing的SpanFromContext方法，断言将interface{}转换成opentracing的span
	sp, ok := c.Ctx.Value(xormHookSpanKey).(opentracing.Span)
	if !ok {
		// 没有则说明没有span
		return nil
	}
	defer sp.Finish()

	// 记录需要的内容
	if c.Err != nil {
		sp.LogFields(log.Object("errors", c.Err))
	}

	// 使用xorm的builder将查询语句和参数结合，方便后期调试
	sql, _ := builder.ConvertToBoundSQL(c.SQL, c.Args)

	// 记录
	sp.LogFields(log.String("SQL", sql))
	sp.LogFields(log.Object("args", c.Args))
	sp.SetTag("execute_time", c.ExecuteTime)

	return nil
}

func NewTracingHook() *TracingHook {
	return &TracingHook{
		before: before,
		after:  after,
	}
}
