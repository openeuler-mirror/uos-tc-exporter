// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package metrics_bak

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewQdiscCbq 测试 NewQdiscCbq 构造函数
func TestNewQdiscCbq(t *testing.T) {
	qdisc := NewQdiscCbq()

	require.NotNil(t, qdisc)
	assert.NotNil(t, qdisc.qdiscCbqAvgIdle)
	assert.NotNil(t, qdisc.qdiscCbqBorrows)
	assert.NotNil(t, qdisc.qdiscCbqOveractions)
	assert.NotNil(t, qdisc.qdiscCbqUnderTime)
}

// TestQdiscCbq_ID 测试 ID 方法
func TestQdiscCbq_ID(t *testing.T) {
	qdisc := NewQdiscCbq()

	id := qdisc.ID()
	assert.Equal(t, "qdisc_cbq", id)
}

// TestNewQdiscCbqAvgIdle 测试 newQdiscCbqAvgIdle 构造函数
func TestNewQdiscCbqAvgIdle(t *testing.T) {
	avgIdle := newQdiscCbqAvgIdle()

	require.NotNil(t, avgIdle)
	require.NotNil(t, avgIdle.baseMetrics)
	assert.Equal(t, "qdisc_cbq_bavg_idle", avgIdle.ID())
}

// TestQdiscCbqAvgIdle_Collect 测试 qdiscCbqAvgIdle 的 Collect 方法
func TestQdiscCbqAvgIdle_Collect(t *testing.T) {
	avgIdle := newQdiscCbqAvgIdle()
	ch := make(chan prometheus.Metric, 1)

	avgIdle.Collect(ch, 1000.0, []string{"test-ns", "eth0", "cbq"})

	select {
	case metric := <-ch:
		desc := metric.Desc()
		assert.Contains(t, desc.String(), "qdisc_cbq_bavg_idle")
	default:
		t.Fatal("Expected metric to be collected")
	}
}

// TestQdiscCbqBorrows_Collect 测试 qdiscCbqBorrows 的 Collect 方法
func TestQdiscCbqBorrows_Collect(t *testing.T) {
	borrows := &qdiscCbqBorrows{
		baseMetrics: NewMetrics(
			"qdisc_cbq_borrows",
			"CBQ borrows xstat",
			[]string{"namespace", "device", "kind"}),
	}
	ch := make(chan prometheus.Metric, 1)

	borrows.Collect(ch, 50.0, []string{"test-ns", "eth0", "cbq"})

	select {
	case metric := <-ch:
		desc := metric.Desc()
		assert.Contains(t, desc.String(), "qdisc_cbq_borrows")
	default:
		t.Fatal("Expected metric to be collected")
	}
}

// TestQdiscCbqOveractions_Collect 测试 qdiscCbqOveractions 的 Collect 方法
func TestQdiscCbqOveractions_Collect(t *testing.T) {
	overactions := &qdiscCbqOveractions{
		baseMetrics: NewMetrics(
			"qdisc_cbq_overactions",
			"CBQ overactions xstat",
			[]string{"namespace", "device", "kind"}),
	}
	ch := make(chan prometheus.Metric, 1)

	overactions.Collect(ch, 10.0, []string{"test-ns", "eth0", "cbq"})

	select {
	case metric := <-ch:
		desc := metric.Desc()
		assert.Contains(t, desc.String(), "qdisc_cbq_overactions")
	default:
		t.Fatal("Expected metric to be collected")
	}
}

// TestQdiscCbqUnderTime_Collect 测试 qdiscCbqUnderTime 的 Collect 方法
func TestQdiscCbqUnderTime_Collect(t *testing.T) {
	underTime := &qdiscCbqUnderTime{
		baseMetrics: NewMetrics(
			"qdisc_cbq_undertime",
			"CBQ undertime xstat",
			[]string{"namespace", "device", "kind"}),
	}
	ch := make(chan prometheus.Metric, 1)

	underTime.Collect(ch, 200.0, []string{"test-ns", "eth0", "cbq"})

	select {
	case metric := <-ch:
		desc := metric.Desc()
		assert.Contains(t, desc.String(), "qdisc_cbq_undertime")
	default:
		t.Fatal("Expected metric to be collected")
	}
}

// TestQdiscCbq_Collect_EmptyChannel 测试 Collect 方法在空通道上的行为
func TestQdiscCbq_Collect_EmptyChannel(t *testing.T) {
	qdisc := NewQdiscCbq()
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

// TestQdiscCbq_Collect_WithBufferedChannel 测试 Collect 方法在有缓冲通道上的行为
func TestQdiscCbq_Collect_WithBufferedChannel(t *testing.T) {
	qdisc := NewQdiscCbq()
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

// TestQdiscCbq_StructFields 测试结构体字段的初始化
func TestQdiscCbq_StructFields(t *testing.T) {
	qdisc := NewQdiscCbq()

	// 验证所有字段都被正确初始化
	assert.NotNil(t, qdisc.qdiscCbqAvgIdle)
	assert.NotNil(t, qdisc.qdiscCbqBorrows)
	assert.NotNil(t, qdisc.qdiscCbqOveractions)
	assert.NotNil(t, qdisc.qdiscCbqUnderTime)

	// 验证字段类型
	assert.IsType(t, qdiscCbqAvgIdle{}, qdisc.qdiscCbqAvgIdle)
	assert.IsType(t, qdiscCbqBorrows{}, qdisc.qdiscCbqBorrows)
	assert.IsType(t, qdiscCbqOveractions{}, qdisc.qdiscCbqOveractions)
	assert.IsType(t, qdiscCbqUnderTime{}, qdisc.qdiscCbqUnderTime)
}

// TestQdiscCbq_ConcurrentAccess 测试并发访问
func TestQdiscCbq_ConcurrentAccess(t *testing.T) {
	qdisc := NewQdiscCbq()
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

// TestQdiscCbq_MetricValues 测试不同指标值的收集
func TestQdiscCbq_MetricValues(t *testing.T) {
	testCases := []struct {
		name     string
		value    float64
		labels   []string
		expected string
	}{
		{
			name:     "zero value",
			value:    0.0,
			labels:   []string{"ns1", "eth0", "cbq"},
			expected: "qdisc_cbq_bavg_idle",
		},
		{
			name:     "positive value",
			value:    1000.0,
			labels:   []string{"ns1", "eth0", "cbq"},
			expected: "qdisc_cbq_bavg_idle",
		},
		{
			name:     "negative value",
			value:    -100.0,
			labels:   []string{"ns1", "eth0", "cbq"},
			expected: "qdisc_cbq_bavg_idle",
		},
		{
			name:     "large value",
			value:    1e10,
			labels:   []string{"ns1", "eth0", "cbq"},
			expected: "qdisc_cbq_bavg_idle",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			avgIdle := newQdiscCbqAvgIdle()
			ch := make(chan prometheus.Metric, 1)

			avgIdle.Collect(ch, tc.value, tc.labels)

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

// TestQdiscCbq_LabelVariations 测试不同标签组合
func TestQdiscCbq_LabelVariations(t *testing.T) {
	testCases := []struct {
		name   string
		labels []string
	}{
		{
			name:   "standard labels",
			labels: []string{"default", "eth0", "cbq"},
		},
		{
			name:   "custom namespace",
			labels: []string{"custom-ns", "eth1", "cbq"},
		},
		{
			name:   "different device",
			labels: []string{"default", "wlan0", "cbq"},
		},
		{
			name:   "empty namespace",
			labels: []string{"", "eth0", "cbq"},
		},
		{
			name:   "empty device",
			labels: []string{"default", "", "cbq"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			avgIdle := newQdiscCbqAvgIdle()
			ch := make(chan prometheus.Metric, 1)

			avgIdle.Collect(ch, 100.0, tc.labels)

			select {
			case metric := <-ch:
				desc := metric.Desc()
				assert.Contains(t, desc.String(), "qdisc_cbq_bavg_idle")
			default:
				t.Fatal("Expected metric to be collected")
			}
		})
	}
}

// TestQdiscCbq_EdgeCases 测试边界情况
func TestQdiscCbq_EdgeCases(t *testing.T) {
	t.Run("zero value", func(t *testing.T) {
		avgIdle := newQdiscCbqAvgIdle()
		ch := make(chan prometheus.Metric, 1)

		// 测试零值
		avgIdle.Collect(ch, 0.0, []string{"ns", "eth0", "cbq"})

		select {
		case metric := <-ch:
			desc := metric.Desc()
			assert.Contains(t, desc.String(), "qdisc_cbq_bavg_idle")
		default:
			t.Fatal("Expected metric to be collected")
		}
	})

	t.Run("negative value", func(t *testing.T) {
		avgIdle := newQdiscCbqAvgIdle()
		ch := make(chan prometheus.Metric, 1)

		// 测试负值
		avgIdle.Collect(ch, -100.0, []string{"ns", "eth0", "cbq"})

		select {
		case metric := <-ch:
			desc := metric.Desc()
			assert.Contains(t, desc.String(), "qdisc_cbq_bavg_idle")
		default:
			t.Fatal("Expected metric to be collected")
		}
	})

	t.Run("large value", func(t *testing.T) {
		avgIdle := newQdiscCbqAvgIdle()
		ch := make(chan prometheus.Metric, 1)

		// 测试大值
		avgIdle.Collect(ch, 1e10, []string{"ns", "eth0", "cbq"})

		select {
		case metric := <-ch:
			desc := metric.Desc()
			assert.Contains(t, desc.String(), "qdisc_cbq_bavg_idle")
		default:
			t.Fatal("Expected metric to be collected")
		}
	})
}
