package app

import (
	"context"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"syscall"
	"time"
)


type LifeCycle interface {
	Start(context.Context) error
	Stop(context.Context) error
}

type Hook struct {
	Start func(context.Context) error
	Stop  func(context.Context) error
}

type options struct {
	startTimeOut time.Duration
	stopTimeOut  time.Duration
	sigs         []os.Signal
	siFn         func(*App, os.Signal)
}

type Option func(o *options)

type App struct {
	opts         options // app的相关属性
	start_before []Hook	// 这里是中间件的相关的Hook
	hooks        []Hook // 这里是所有的服务相关serHook
	cancel       func() // 整个服务的关闭
}

// 设置超时时间
func SetStartTimeOut(t time.Duration) Option {
	return func(o *options) {
		o.startTimeOut = t
	}
}

func New(opts ...Option) *App {
	options := options{
		startTimeOut: time.Second * 30,
		stopTimeOut:  time.Second * 30,
		sigs: []os.Signal{
			syscall.SIGTERM,
			syscall.SIGQUIT,
			syscall.SIGINT,
		},
		siFn: func(app *App, signal os.Signal) {
			switch signal {
			case syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT:
				app.Stop()
			default:
			}
		},
	}
	for _, o := range opts {
		o(&options)
	}
	return &App{
		opts: options,
	}
}

func (a *App) Run() error {
	var ctx context.Context
	ctx, a.cancel = context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)
	// startbefore 在启动前加载的服务应该是独立的
	err := a.runStartBefore(ctx)
	if err != nil {
		return errors.Wrap(err, "app :bad ending")
	}

	// 循环所有的中间件，ctx的关闭事件，如果ctx取消，就释放掉所有的中间件
	for _, hook := range a.start_before {
		hook := hook
		if hook.Stop != nil {
			g.Go(func() error {
				<-ctx.Done()
				stopCtx, cancel := context.WithTimeout(ctx, a.opts.stopTimeOut)
				defer cancel()
				return hook.Stop(stopCtx)
			})
		}
	}
	//  循环监控所有注册的服务，启动和ctx的关闭
	for _, hook := range a.hooks {
		hook := hook
		if hook.Stop != nil {
			g.Go(func() error {
				<-ctx.Done()
				stopCtx, cancel := context.WithTimeout(ctx, a.opts.stopTimeOut)
				defer cancel()
				return hook.Stop(stopCtx)
			})
		}
		if hook.Start != nil {
			g.Go(func() error {
				starCtx, cancel := context.WithTimeout(ctx, a.opts.startTimeOut)
				defer cancel()
				return hook.Start(starCtx)
			})
		}
	}
	// 如果没有信号，就关闭它
	if len(a.opts.sigs) == 0 {
		return g.Wait()
	}
	c := make(chan os.Signal, len(a.opts.sigs))
	signal.Notify(c, a.opts.sigs...)
	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case sig := <-c:
				if a.opts.siFn != nil {
					a.opts.siFn(a, sig)
				}
			}

		}
	})

	return g.Wait()
}

// 启动独立中间件：mysql ，redis ，mq ping 等
func (a *App) runStartBefore(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	for _, hook := range a.start_before {
		hook := hook
		if hook.Start != nil {
			g.Go(func() error {
				starCtx, cancel := context.WithTimeout(ctx, a.opts.startTimeOut)
				defer cancel()
				return hook.Start(starCtx)
			})
		}
	}
	return g.Wait()
}


// 启动真实服务 the real server ,include http,service
func (a *App) runHooks(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	for _, hook := range a.start_before {
		hook := hook
		if hook.Stop != nil {
			g.Go(func() error {
				<-ctx.Done()
				stopCtx, cancel := context.WithTimeout(ctx, a.opts.stopTimeOut)
				defer cancel()
				return hook.Stop(stopCtx)
			})
		}
		if hook.Start != nil {
			g.Go(func() error {
				starCtx, cancel := context.WithTimeout(ctx, a.opts.startTimeOut)
				defer cancel()
				return hook.Start(starCtx)
			})
		}
	}
	return g.Wait()
}

func (a *App) Stop() {
	if a.cancel != nil {
		a.cancel()
	}
}

// 需要将服务进行管理，具体服务分为mysql，redis等 中间件，
// 这些服务是独立的，应该作为startbefore来进行管理操作，
// 具体的其他服务应该用lifecycle进行管理
func (a *App) AppendStartBefore(lc LifeCycle) {
	a.start_before = append(a.start_before, Hook{
		Start: func(ctx context.Context) error {
			return lc.Start(ctx)
		},
		Stop: func(ctx context.Context) error {
			return lc.Stop(ctx)
		},
	})
}

func (a *App) Append(lc LifeCycle) {
	a.hooks = append(a.hooks, Hook{
		Start: func(ctx context.Context) error {
			return lc.Start(ctx)
		},
		Stop: func(ctx context.Context) error {
			return lc.Stop(ctx)
		},
	})
}

func (a *App) AppendHook(hook Hook) {
	a.hooks = append(a.hooks, hook)
}
