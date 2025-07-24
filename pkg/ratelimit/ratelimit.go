// Package ratelimit 提供了基于令牌桶算法的限流器实现
//
// 限流器通过控制令牌的发放速率来限制操作的频率，适用于API限流、
// 资源访问控制等场景。支持并发安全的令牌获取和释放。
//
// 使用示例：
//
//	limiter, err := NewRateLimiter(time.Second, 10) // 每秒最多10个请求
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer limiter.Stop()
//
//	if err := limiter.Get(); err != nil {
//	    // 处理限流情况
//	}
package ratelimit

import (
	"context"
	"errors"
	"sync"
	"time"
)

// 预定义的错误类型
var (
	// ErrRateLimited 当令牌不可用时返回此错误
	ErrRateLimited = errors.New("rate limited: no tokens available")
	// ErrRateLimitSize 当令牌桶大小无效时返回此错误
	ErrRateLimitSize = errors.New("rate limit error: bucket size must be greater than zero")
	// ErrRateLimitTime 当限流时间间隔无效时返回此错误
	ErrRateLimitTime = errors.New("rate limit error: time interval must be greater than zero")
	// ErrLimiterClosed 当限流器已关闭时返回此错误
	ErrLimiterClosed = errors.New("rate limiter is closed")
)

// RateLimiter 是一个基于令牌桶算法的限流器
//
// 它通过定期向令牌桶中添加令牌来控制操作频率。
// 每次调用Get()方法时会尝试从桶中取出一个令牌，
// 如果桶为空则返回限流错误。
type RateLimiter struct {
	// tokens 令牌通道，用作令牌桶
	tokens chan struct{}
	// limit 令牌补充的时间间隔
	limit time.Duration
	// ticker 定时器，用于定期补充令牌
	ticker *time.Ticker
	// ctx 上下文，用于控制goroutine生命周期
	ctx context.Context
	// cancel 取消函数，用于停止令牌补充
	cancel context.CancelFunc
	// mu 读写锁，保护closed字段
	mu sync.RWMutex
	// closed 标记限流器是否已关闭
	closed bool
	// wg 等待组，确保goroutine正确退出
	wg sync.WaitGroup
}

// NewRateLimiter 创建一个新的限流器
//
// 参数：
//   - limit: 令牌补充的时间间隔，例如time.Second表示每秒补充一个令牌
//   - bucketSize: 令牌桶的容量，即最多可以存储的令牌数量
//
// 返回：
//   - *RateLimiter: 新创建的限流器实例
//   - error: 如果参数无效则返回错误
//
// 注意：需要调用Stop()方法来正确关闭限流器，释放资源
func NewRateLimiter(limit time.Duration, bucketSize int) (*RateLimiter, error) {
	if bucketSize <= 0 {
		return nil, ErrRateLimitSize
	}
	if limit <= 0 {
		return nil, ErrRateLimitTime
	}

	ctx, cancel := context.WithCancel(context.Background())

	rl := &RateLimiter{
		tokens: make(chan struct{}, bucketSize),
		limit:  limit,
		ticker: time.NewTicker(limit),
		ctx:    ctx,
		cancel: cancel,
		closed: false,
	}

	// 初始化时填满令牌桶
	for i := 0; i < bucketSize; i++ {
		rl.tokens <- struct{}{}
	}

	// 启动令牌补充goroutine
	rl.wg.Add(1)
	go rl.startRefreshTokens()

	return rl, nil
}

// startRefreshTokens 定期向令牌桶中补充令牌
// 这是一个内部方法，在单独的goroutine中运行
func (rl *RateLimiter) startRefreshTokens() {
	defer rl.wg.Done()

	for {
		select {
		case <-rl.ctx.Done():
			// 收到停止信号，退出goroutine
			return
		case <-rl.ticker.C:
			// 尝试添加一个令牌到桶中
			// 使用非阻塞的select确保当桶满时不会阻塞
			select {
			case rl.tokens <- struct{}{}:
				// 成功添加令牌
			default:
				// 桶已满，跳过这次补充
			}
		}
	}
}

// Get 尝试获取一个令牌
//
// 返回：
//   - nil: 成功获取令牌
//   - ErrRateLimited: 没有可用令牌，请求被限流
//   - ErrLimiterClosed: 限流器已关闭
//
// 此方法是并发安全的，可以从多个goroutine同时调用
func (rl *RateLimiter) Get() error {
	// 检查限流器是否已关闭
	rl.mu.RLock()
	if rl.closed {
		rl.mu.RUnlock()
		return ErrLimiterClosed
	}
	rl.mu.RUnlock()

	// 尝试从令牌桶中获取令牌（非阻塞）
	select {
	case _, ok := <-rl.tokens:
		if !ok {
			// 通道已关闭
			return ErrLimiterClosed
		}
		return nil
	default:
		// 没有可用令牌
		return ErrRateLimited
	}
}

// TryGetWithTimeout 在指定超时时间内尝试获取令牌
//
// 参数：
//   - timeout: 等待令牌的最大时间
//
// 返回：
//   - nil: 成功获取令牌
//   - ErrRateLimited: 超时未获取到令牌
//   - ErrLimiterClosed: 限流器已关闭
func (rl *RateLimiter) TryGetWithTimeout(timeout time.Duration) error {
	// 检查限流器是否已关闭
	rl.mu.RLock()
	if rl.closed {
		rl.mu.RUnlock()
		return ErrLimiterClosed
	}
	rl.mu.RUnlock()

	// 创建超时定时器
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case _, ok := <-rl.tokens:
		if !ok {
			return ErrLimiterClosed
		}
		return nil
	case <-timer.C:
		return ErrRateLimited
	case <-rl.ctx.Done():
		return ErrLimiterClosed
	}
}

// Stop 停止限流器并释放所有资源
//
// 此方法会：
// 1. 停止令牌补充
// 2. 关闭令牌通道
// 3. 等待所有goroutine退出
// 4. 释放相关资源
//
// 调用Stop后，限流器将无法再使用，所有Get操作都会返回ErrLimiterClosed
// 此方法是幂等的，多次调用是安全的
func (rl *RateLimiter) Stop() {
	rl.mu.Lock()
	if rl.closed {
		// 已经关闭，直接返回
		rl.mu.Unlock()
		return
	}
	rl.closed = true
	rl.mu.Unlock()

	// 停止定时器
	if rl.ticker != nil {
		rl.ticker.Stop()
	}

	// 取消上下文，通知goroutine退出
	if rl.cancel != nil {
		rl.cancel()
	}

	// 等待goroutine退出
	rl.wg.Wait()

	// 关闭令牌通道
	if rl.tokens != nil {
		close(rl.tokens)
	}
}

// IsClosed 检查限流器是否已关闭
//
// 返回：
//   - true: 限流器已关闭
//   - false: 限流器仍在运行
func (rl *RateLimiter) IsClosed() bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return rl.closed
}

// GetLimit 获取当前的限流时间间隔
//
// 返回：
//   - time.Duration: 令牌补充的时间间隔
func (rl *RateLimiter) GetLimit() time.Duration {
	return rl.limit
}

// GetBucketSize 获取令牌桶的容量
//
// 返回：
//   - int: 令牌桶的最大容量
func (rl *RateLimiter) GetBucketSize() int {
	return cap(rl.tokens)
}

// GetAvailableTokens 获取当前可用令牌数量
//
// 返回：
//   - int: 当前令牌桶中的令牌数量
//
// 注意：这个值只是一个近似值，因为在并发环境下
// 令牌数量可能在调用之间发生变化
func (rl *RateLimiter) GetAvailableTokens() int {
	rl.mu.RLock()
	if rl.closed {
		rl.mu.RUnlock()
		return 0
	}
	rl.mu.RUnlock()

	return len(rl.tokens)
}
