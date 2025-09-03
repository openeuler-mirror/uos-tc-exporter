// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewQdiscChoke 测试 NewQdiscChoke 构造函数
func TestNewQdiscChoke(t *testing.T) {
	qdisc := NewQdiscChoke()

	require.NotNil(t, qdisc)
	assert.NotNil(t, qdisc.qdiscChokeEarly)
	assert.NotNil(t, qdisc.qdiscChokeMarked)
	assert.NotNil(t, qdisc.qdiscChokeMatched)
	assert.NotNil(t, qdisc.qdiscChokeOther)
	assert.NotNil(t, qdisc.qdiscChokePdrop)
}

// TestQdiscChoke_ID 测试 ID 方法
func TestQdiscChoke_ID(t *testing.T) {
	qdisc := NewQdiscChoke()

	id := qdisc.ID()
	assert.Equal(t, "qdisc_choke", id)
}

// TestNewQdiscChokeEarly 测试 newQdiscChokeEarly 构造函数
func TestNewQdiscChokeEarly(t *testing.T) {
	early := newQdiscChokeEarly()

	require.NotNil(t, early)
	require.NotNil(t, early.baseMetrics)
	assert.Equal(t, "qdisc_choke_early", early.ID())
}

// TestNewQdiscChokeMarked 测试 newQdiscChokeMarked 构造函数
func TestNewQdiscChokeMarked(t *testing.T) {
	marked := newQdiscChokeMarked()

	require.NotNil(t, marked)
	require.NotNil(t, marked.baseMetrics)
	assert.Equal(t, "qdisc_choke_marked", marked.ID())
}

// TestNewQdiscChokeMatched 测试 newQdiscChokeMatched 构造函数
func TestNewQdiscChokeMatched(t *testing.T) {
	matched := newQdiscChokeMatched()

	require.NotNil(t, matched)
	require.NotNil(t, matched.baseMetrics)
	assert.Equal(t, "qdisc_choke_matched", matched.ID())
}

// TestNewQdiscChokeOther 测试 newQdiscChokeOther 构造函数
func TestNewQdiscChokeOther(t *testing.T) {
	other := newQdiscChokeOther()

	require.NotNil(t, other)
	require.NotNil(t, other.baseMetrics)
	assert.Equal(t, "qdisc_choke_other", other.ID())
}

// TestNewQdiscChokePdrop 测试 newQdiscChokePdrop 构造函数
func TestNewQdiscChokePdrop(t *testing.T) {
	pdrop := newQdiscChokePdrop()

	require.NotNil(t, pdrop)
	require.NotNil(t, pdrop.baseMetrics)
	assert.Equal(t, "qdisc_choke_pdrop", pdrop.ID())
}

// TestQdiscChokeEarly_Collect 测试 qdiscChokeEarly 的 Collect 方法
func TestQdiscChokeEarly_Collect(t *testing.T) {
	early := newQdiscChokeEarly()
	ch := make(chan prometheus.Metric, 1)

	early.Collect(ch, 100.0, []string{"test-ns", "eth0", "choke"})

	select {
	case metric := <-ch:
		desc := metric.Desc()
		assert.Contains(t, desc.String(), "qdisc_choke_early")
	default:
		t.Fatal("Expected metric to be collected")
	}
}

// TestQdiscChokeMarked_Collect 测试 qdiscChokeMarked 的 Collect 方法
func TestQdiscChokeMarked_Collect(t *testing.T) {
	marked := newQdiscChokeMarked()
	ch := make(chan prometheus.Metric, 1)

	marked.Collect(ch, 50.0, []string{"test-ns", "eth0", "choke"})

	select {
	case metric := <-ch:
		desc := metric.Desc()
		assert.Contains(t, desc.String(), "qdisc_choke_marked")
	default:
		t.Fatal("Expected metric to be collected")
	}
}

// TestQdiscChokeMatched_Collect 测试 qdiscChokeMatched 的 Collect 方法
func TestQdiscChokeMatched_Collect(t *testing.T) {
	matched := newQdiscChokeMatched()
	ch := make(chan prometheus.Metric, 1)

	matched.Collect(ch, 25.0, []string{"test-ns", "eth0", "choke"})

	select {
	case metric := <-ch:
		desc := metric.Desc()
		assert.Contains(t, desc.String(), "qdisc_choke_matched")
	default:
		t.Fatal("Expected metric to be collected")
	}
}

// TestQdiscChokeOther_Collect 测试 qdiscChokeOther 的 Collect 方法
func TestQdiscChokeOther_Collect(t *testing.T) {
	other := newQdiscChokeOther()
	ch := make(chan prometheus.Metric, 1)

	other.Collect(ch, 10.0, []string{"test-ns", "eth0", "choke"})

	select {
	case metric := <-ch:
		desc := metric.Desc()
		assert.Contains(t, desc.String(), "qdisc_choke_other")
	default:
		t.Fatal("Expected metric to be collected")
	}
}

// TestQdiscChokePdrop_Collect 测试 qdiscChokePdrop 的 Collect 方法
func TestQdiscChokePdrop_Collect(t *testing.T) {
	pdrop := newQdiscChokePdrop()
	ch := make(chan prometheus.Metric, 1)

	pdrop.Collect(ch, 5.0, []string{"test-ns", "eth0", "choke"})

	select {
	case metric := <-ch:
		desc := metric.Desc()
		assert.Contains(t, desc.String(), "qdisc_choke_pdrop")
	default:
		t.Fatal("Expected metric to be collected")
	}
}

// TestQdiscChoke_Collect_EmptyChannel 测试 Collect 方法在空通道上的行为
func TestQdiscChoke_Collect_EmptyChannel(t *testing.T) {
	qdisc := NewQdiscChoke()
	ch := make(chan prometheus.Metric) // 无缓冲通道

	// 这个测试主要验证 Collect 方法不会因为通道问题而崩溃
	// 在实际环境中，tc 包的函数调用可能会失败，但方法本身应该能正常处理
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Collect method panicked: %v", r)
		}
	}()

	// 由于我们无法轻易模拟 tc 包的行为，这里主要测试方法不会崩溃
	qdisc.Collect(ch)
}

// TestQdiscChoke_Collect_WithBufferedChannel 测试 Collect 方法在有缓冲通道上的行为
func TestQdiscChoke_Collect_WithBufferedChannel(t *testing.T) {
	qdisc := NewQdiscChoke()
	ch := make(chan prometheus.Metric, 10) // 有缓冲通道

	// 这个测试主要验证 Collect 方法不会因为通道问题而崩溃
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Collect method panicked: %v", r)
		}
	}()

	// 由于我们无法轻易模拟 tc 包的行为，这里主要测试方法不会崩溃
	qdisc.Collect(ch)
}

// TestQdiscChoke_StructFields 测试结构体字段的初始化
func TestQdiscChoke_StructFields(t *testing.T) {
	qdisc := NewQdiscChoke()

	// 验证所有字段都被正确初始化
	assert.NotNil(t, qdisc.qdiscChokeEarly)
	assert.NotNil(t, qdisc.qdiscChokeMarked)
	assert.NotNil(t, qdisc.qdiscChokeMatched)
	assert.NotNil(t, qdisc.qdiscChokeOther)
	assert.NotNil(t, qdisc.qdiscChokePdrop)

	// 验证字段类型
	assert.IsType(t, qdiscChokeEarly{}, qdisc.qdiscChokeEarly)
	assert.IsType(t, qdiscChokeMarked{}, qdisc.qdiscChokeMarked)
	assert.IsType(t, qdiscChokeMatched{}, qdisc.qdiscChokeMatched)
	assert.IsType(t, qdiscChokeOther{}, qdisc.qdiscChokeOther)
	assert.IsType(t, qdiscChokePdrop{}, qdisc.qdiscChokePdrop)
}

// TestQdiscChoke_ConcurrentAccess 测试并发访问
func TestQdiscChoke_ConcurrentAccess(t *testing.T) {
	qdisc := NewQdiscChoke()
	ch := make(chan prometheus.Metric, 10)

	// 启动多个 goroutine 并发调用 Collect
	done := make(chan bool, 3)

	for i := 0; i < 3; i++ {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Collect method panicked in goroutine: %v", r)
				}
				done <- true
			}()
			qdisc.Collect(ch)
		}()
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 3; i++ {
		<-done
	}
}

// TestQdiscChoke_MetricValues 测试不同指标值的收集
func TestQdiscChoke_MetricValues(t *testing.T) {
	testCases := []struct {
		name     string
		value    float64
		labels   []string
		expected string
	}{
		{
			name:     "zero value",
			value:    0.0,
			labels:   []string{"ns1", "eth0", "choke"},
			expected: "qdisc_choke_early",
		},
		{
			name:     "positive value",
			value:    1000.0,
			labels:   []string{"ns1", "eth0", "choke"},
			expected: "qdisc_choke_early",
		},
		{
			name:     "negative value",
			value:    -100.0,
			labels:   []string{"ns1", "eth0", "choke"},
			expected: "qdisc_choke_early",
		},
		{
			name:     "large value",
			value:    1e10,
			labels:   []string{"ns1", "eth0", "choke"},
			expected: "qdisc_choke_early",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			early := newQdiscChokeEarly()
			ch := make(chan prometheus.Metric, 1)

			early.Collect(ch, tc.value, tc.labels)

			select {
			case metric := <-ch:
				desc := metric.Desc()
				assert.Contains(t, desc.String(), tc.expected)
			default:
				t.Fatal("Expected metric to be collected")
			}
		})
	}
}

// TestQdiscChoke_LabelVariations 测试不同标签组合
func TestQdiscChoke_LabelVariations(t *testing.T) {
	testCases := []struct {
		name   string
		labels []string
	}{
		{
			name:   "standard labels",
			labels: []string{"default", "eth0", "choke"},
		},
		{
			name:   "custom namespace",
			labels: []string{"custom-ns", "eth1", "choke"},
		},
		{
			name:   "different device",
			labels: []string{"default", "wlan0", "choke"},
		},
		{
			name:   "empty namespace",
			labels: []string{"", "eth0", "choke"},
		},
		{
			name:   "empty device",
			labels: []string{"default", "", "choke"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			early := newQdiscChokeEarly()
			ch := make(chan prometheus.Metric, 1)

			early.Collect(ch, 100.0, tc.labels)

			select {
			case metric := <-ch:
				desc := metric.Desc()
				assert.Contains(t, desc.String(), "qdisc_choke_early")
			default:
				t.Fatal("Expected metric to be collected")
			}
		})
	}
}

// TestQdiscChoke_EdgeCases 测试边界情况
func TestQdiscChoke_EdgeCases(t *testing.T) {
	t.Run("zero value", func(t *testing.T) {
		early := newQdiscChokeEarly()
		ch := make(chan prometheus.Metric, 1)

		// 测试零值
		early.Collect(ch, 0.0, []string{"ns", "eth0", "choke"})

		select {
		case metric := <-ch:
			desc := metric.Desc()
			assert.Contains(t, desc.String(), "qdisc_choke_early")
		default:
			t.Fatal("Expected metric to be collected")
		}
	})

	t.Run("negative value", func(t *testing.T) {
		early := newQdiscChokeEarly()
		ch := make(chan prometheus.Metric, 1)

		// 测试负值
		early.Collect(ch, -100.0, []string{"ns", "eth0", "choke"})

		select {
		case metric := <-ch:
			desc := metric.Desc()
			assert.Contains(t, desc.String(), "qdisc_choke_early")
		default:
			t.Fatal("Expected metric to be collected")
		}
	})

	t.Run("large value", func(t *testing.T) {
		early := newQdiscChokeEarly()
		ch := make(chan prometheus.Metric, 1)

		// 测试大值
		early.Collect(ch, 1e10, []string{"ns", "eth0", "choke"})

		select {
		case metric := <-ch:
			desc := metric.Desc()
			assert.Contains(t, desc.String(), "qdisc_choke_early")
		default:
			t.Fatal("Expected metric to be collected")
		}
	})
}

// TestQdiscChoke_AllMetricTypes 测试所有指标类型的收集
func TestQdiscChoke_AllMetricTypes(t *testing.T) {
	metricTypes := []struct {
		name      string
		value     float64
		expected  string
		collector func() interface {
			Collect(ch chan<- prometheus.Metric, value float64, labels []string)
		}
	}{
		{
			name:     "early",
			value:    100.0,
			expected: "qdisc_choke_early",
			collector: func() interface {
				Collect(ch chan<- prometheus.Metric, value float64, labels []string)
			} {
				return newQdiscChokeEarly()
			},
		},
		{
			name:     "marked",
			value:    50.0,
			expected: "qdisc_choke_marked",
			collector: func() interface {
				Collect(ch chan<- prometheus.Metric, value float64, labels []string)
			} {
				return newQdiscChokeMarked()
			},
		},
		{
			name:     "matched",
			value:    25.0,
			expected: "qdisc_choke_matched",
			collector: func() interface {
				Collect(ch chan<- prometheus.Metric, value float64, labels []string)
			} {
				return newQdiscChokeMatched()
			},
		},
		{
			name:     "other",
			value:    10.0,
			expected: "qdisc_choke_other",
			collector: func() interface {
				Collect(ch chan<- prometheus.Metric, value float64, labels []string)
			} {
				return newQdiscChokeOther()
			},
		},
		{
			name:     "pdrop",
			value:    5.0,
			expected: "qdisc_choke_pdrop",
			collector: func() interface {
				Collect(ch chan<- prometheus.Metric, value float64, labels []string)
			} {
				return newQdiscChokePdrop()
			},
		},
	}

	for _, mt := range metricTypes {
		t.Run(mt.name, func(t *testing.T) {
			collector := mt.collector()
			ch := make(chan prometheus.Metric, 1)

			collector.Collect(ch, mt.value, []string{"ns", "eth0", "choke"})

			select {
			case metric := <-ch:
				desc := metric.Desc()
				assert.Contains(t, desc.String(), mt.expected)
			default:
				t.Fatal("Expected metric to be collected")
			}
		})
	}
}

// TestQdiscChoke_MultipleValues 测试多个值的同时收集
func TestQdiscChoke_MultipleValues(t *testing.T) {
	ch := make(chan prometheus.Metric, 5)

	// 收集所有5种指标类型
	early := newQdiscChokeEarly()
	marked := newQdiscChokeMarked()
	matched := newQdiscChokeMatched()
	other := newQdiscChokeOther()
	pdrop := newQdiscChokePdrop()

	early.Collect(ch, 100.0, []string{"ns", "eth0", "choke"})
	marked.Collect(ch, 50.0, []string{"ns", "eth0", "choke"})
	matched.Collect(ch, 25.0, []string{"ns", "eth0", "choke"})
	other.Collect(ch, 10.0, []string{"ns", "eth0", "choke"})
	pdrop.Collect(ch, 5.0, []string{"ns", "eth0", "choke"})

	// 验证收集到了5个指标
	metrics := make([]prometheus.Metric, 0, 5)
	for i := 0; i < 5; i++ {
		select {
		case metric := <-ch:
			metrics = append(metrics, metric)
		default:
			t.Fatalf("Expected metric %d to be collected", i)
		}
	}

	assert.Len(t, metrics, 5)

	// 验证每个指标都有正确的描述符
	expectedNames := []string{
		"qdisc_choke_early",
		"qdisc_choke_marked",
		"qdisc_choke_matched",
		"qdisc_choke_other",
		"qdisc_choke_pdrop",
	}

	for i, metric := range metrics {
		desc := metric.Desc()
		assert.Contains(t, desc.String(), expectedNames[i])
	}
}

// TestQdiscChoke_Performance 测试性能相关场景
func TestQdiscChoke_Performance(t *testing.T) {
	early := newQdiscChokeEarly()
	ch := make(chan prometheus.Metric, 1000)

	// 测试大量指标收集的性能
	start := time.Now()
	for i := 0; i < 1000; i++ {
		early.Collect(ch, float64(i), []string{"ns", "eth0", "choke"})
	}
	duration := time.Since(start)

	// 验证所有指标都被收集
	metrics := make([]prometheus.Metric, 0, 1000)
	for i := 0; i < 1000; i++ {
		select {
		case metric := <-ch:
			metrics = append(metrics, metric)
		default:
			t.Fatalf("Expected metric %d to be collected", i)
		}
	}

	assert.Len(t, metrics, 1000)
	assert.True(t, duration < time.Second, "Performance test took too long: %v", duration)
}
