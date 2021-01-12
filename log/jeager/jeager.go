/**
  @author: $(USER)
  @data:$(DATE)
  @note:
**/
package jeager

import (
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"io"
)

//分布式追踪系统
func InitJeager(serviceName string, addr string, port int) (closer io.Closer, err error) {
	host := fmt.Sprintf("%d:%d", addr, port)
	// 根据配置初始化Tracer 返回Closer
	tracer, closer, err := (&config.Configuration{
		ServiceName: serviceName,
		Disabled:    false,
		Sampler: &config.SamplerConfig{
			Type: jaeger.SamplerTypeConst,
			// param的值在0到1之间，设置为1则将所有的Operation输出到Reporter
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			// 请注意，如果logSpans没开启就一定没有结果
			LogSpans: true,
			// 请一定要注意，LocalAgentHostPort填错了就一定没有结果
			LocalAgentHostPort: host,
		},
	}).NewTracer()
	if err != nil {
		return
	}

	// 设置全局Tracer - 如果不设置将会导致上下文无法生成正确的Span
	opentracing.SetGlobalTracer(tracer)
	return
}
