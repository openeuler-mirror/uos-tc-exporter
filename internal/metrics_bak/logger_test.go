// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package metrics_bak

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetMetricsLogger(t *testing.T) {
	// 重置全局logger以确保测试独立性
	globalLogger = nil
	once = sync.Once{}

	// 测试单例模式
	logger1 := GetMetricsLogger()
	logger2 := GetMetricsLogger()

	assert.NotNil(t, logger1)
	assert.Equal(t, logger1, logger2)
	assert.Equal(t, 5*time.Second, logger1.logInterval)
}

func TestMetricsLogger_LogCollectionStart(t *testing.T) {
	logger := &MetricsLogger{
		lastLogTime:  make(map[string]time.Time),
		logInterval:  100 * time.Millisecond, // 使用较短的间隔便于测试
		metricsCount: make(map[string]int),
		errorCount:   make(map[string]int),
	}

	// 测试首次调用
	logger.LogCollectionStart("test_collector")

	assert.Equal(t, 0, logger.metricsCount["test_collector"])
	assert.Equal(t, 0, logger.errorCount["test_collector"])

	// 测试短时间内重复调用（应该被抑制）
	logger.LogCollectionStart("test_collector")

	// 等待间隔时间后再次调用
	time.Sleep(150 * time.Millisecond)
	logger.LogCollectionStart("test_collector")

	// 验证计数器被重置
	assert.Equal(t, 0, logger.metricsCount["test_collector"])
}

func TestMetricsLogger_LogCollectionComplete(t *testing.T) {
	logger := &MetricsLogger{
		lastLogTime:  make(map[string]time.Time),
		logInterval:  100 * time.Millisecond,
		metricsCount: make(map[string]int),
		errorCount:   make(map[string]int),
	}

	// 测试无指标时的情况
	logger.LogCollectionComplete("test_collector")

	// 设置指标计数后测试
	logger.metricsCount["test_collector"] = 42
	logger.LogCollectionComplete("test_collector")

	// 验证计数保持不变
	assert.Equal(t, 42, logger.metricsCount["test_collector"])
}

func TestMetricsLogger_LogError(t *testing.T) {
	logger := &MetricsLogger{
		lastLogTime:  make(map[string]time.Time),
		logInterval:  100 * time.Millisecond,
		metricsCount: make(map[string]int),
		errorCount:   make(map[string]int),
	}

	// 测试首次错误记录
	err := assert.AnError
	logger.LogError("test_collector", "test_operation", err)

	assert.Equal(t, 1, logger.errorCount["test_collector"])

	// 测试短时间内重复错误（应该被抑制）
	logger.LogError("test_collector", "test_operation", err)
	assert.Equal(t, 1, logger.errorCount["test_collector"])

	// 等待间隔时间后再次记录错误
	time.Sleep(150 * time.Millisecond)
	logger.LogError("test_collector", "test_operation", err)
	assert.Equal(t, 2, logger.errorCount["test_collector"])

	// 测试不同操作的不同错误
	logger.LogError("test_collector", "different_operation", err)
	assert.Equal(t, 1, logger.errorCount["test_collector"]) // 不同操作，错误计数独立
}

func TestMetricsLogger_LogNoData(t *testing.T) {
	logger := &MetricsLogger{
		lastLogTime:  make(map[string]time.Time),
		logInterval:  100 * time.Millisecond,
		metricsCount: make(map[string]int),
		errorCount:   make(map[string]int),
	}

	// 测试首次无数据记录
	logger.LogNoData("test_collector", "no_namespace")

	// 测试短时间内重复记录（应该被抑制）
	logger.LogNoData("test_collector", "no_namespace")

	// 等待间隔时间后再次记录
	time.Sleep(150 * time.Millisecond)
	logger.LogNoData("test_collector", "no_namespace")

	// 测试不同原因的不同记录
	logger.LogNoData("test_collector", "no_interfaces")
}

func TestMetricsLogger_IncrementMetricsCount(t *testing.T) {
	logger := &MetricsLogger{
		lastLogTime:  make(map[string]time.Time),
		logInterval:  100 * time.Millisecond,
		metricsCount: make(map[string]int),
		errorCount:   make(map[string]int),
	}

	// 测试增加不存在的收集器计数（应该无效果）
	logger.IncrementMetricsCount("nonexistent", 5)

	// 初始化收集器后测试
	logger.metricsCount["test_collector"] = 10
	logger.IncrementMetricsCount("test_collector", 5)
	assert.Equal(t, 15, logger.metricsCount["test_collector"])

	// 测试增加负数
	logger.IncrementMetricsCount("test_collector", -3)
	assert.Equal(t, 12, logger.metricsCount["test_collector"])
}

func TestMetricsLogger_GetStats(t *testing.T) {
	logger := &MetricsLogger{
		lastLogTime:  make(map[string]time.Time),
		logInterval:  100 * time.Millisecond,
		metricsCount: make(map[string]int),
		errorCount:   make(map[string]int),
	}

	// 测试空统计
	stats := logger.GetStats()
	assert.Empty(t, stats)

	// 设置一些数据后测试
	logger.metricsCount["collector1"] = 100
	logger.metricsCount["collector2"] = 200
	logger.errorCount["collector1"] = 5
	logger.errorCount["collector2"] = 10

	stats = logger.GetStats()
	expected := map[string]interface{}{
		"collector1_metrics": 100,
		"collector2_metrics": 200,
		"collector1_errors":  5,
		"collector2_errors":  10,
	}

	assert.Equal(t, expected, stats)
}

func TestMetricsLogger_SetLogInterval(t *testing.T) {
	logger := &MetricsLogger{
		lastLogTime:  make(map[string]time.Time),
		logInterval:  100 * time.Millisecond,
		metricsCount: make(map[string]int),
		errorCount:   make(map[string]int),
	}

	// 测试设置新的间隔
	newInterval := 200 * time.Millisecond
	logger.SetLogInterval(newInterval)
	assert.Equal(t, newInterval, logger.logInterval)

	// 测试设置零间隔
	logger.SetLogInterval(0)
	assert.Equal(t, time.Duration(0), logger.logInterval)
}

func TestMetricsLogger_Concurrency(t *testing.T) {
	logger := &MetricsLogger{
		lastLogTime:  make(map[string]time.Time),
		logInterval:  10 * time.Millisecond,
		metricsCount: make(map[string]int),
		errorCount:   make(map[string]int),
	}

	// 并发测试
	const numGoroutines = 100
	var wg sync.WaitGroup

	// 并发增加指标计数
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			logger.IncrementMetricsCount("concurrent_test", 1)
		}(i)
	}

	wg.Wait()

	// 验证最终计数
	assert.Equal(t, numGoroutines, logger.metricsCount["concurrent_test"])
}

func TestMetricsLogger_EdgeCases(t *testing.T) {
	logger := &MetricsLogger{
		lastLogTime:  make(map[string]time.Time),
		logInterval:  100 * time.Millisecond,
		metricsCount: make(map[string]int),
		errorCount:   make(map[string]int),
	}

	// 测试空字符串
	logger.LogCollectionStart("")
	logger.LogError("", "", assert.AnError)
	logger.LogNoData("", "")
	logger.IncrementMetricsCount("", 1)

	// 测试特殊字符
	logger.LogCollectionStart("test:collector")
	logger.LogError("test:collector", "test:operation", assert.AnError)

	// 测试极长字符串
	longString := string(make([]byte, 1000))
	logger.LogCollectionStart(longString)

	// 验证没有panic
	assert.True(t, true)
}

func TestMetricsLogger_Reset(t *testing.T) {
	// 测试重置全局logger
	globalLogger = nil
	once = sync.Once{}

	// 重新获取logger
	logger := GetMetricsLogger()
	assert.NotNil(t, logger)

	// 验证初始状态
	assert.Equal(t, 5*time.Second, logger.logInterval)
	assert.Empty(t, logger.metricsCount)
	assert.Empty(t, logger.errorCount)
}
