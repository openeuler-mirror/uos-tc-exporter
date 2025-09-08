// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package metrics_bak

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewQdiscCodel 测试 NewQdiscCodel 构造函数
func TestNewQdiscCodel(t *testing.T) {
	qdisc := NewQdiscCodel()

	require.NotNil(t, qdisc)
	assert.NotNil(t, qdisc.qdiscCodelCeMark)
	assert.NotNil(t, qdisc.qdiscCodelCount)
	assert.NotNil(t, qdisc.qdiscCodelDropNext)
	assert.NotNil(t, qdisc.qdiscCodelDropOverlimit)
	assert.NotNil(t, qdisc.qdiscCodelDropping)
	assert.NotNil(t, qdisc.qdiscCodelEcnMark)
	assert.NotNil(t, qdisc.qdiscCodelLdelay)
	assert.NotNil(t, qdisc.qdiscCodelMaxPacket)
}

// TestQdiscCodel_ID 测试 ID 方法
func TestQdiscCodel_ID(t *testing.T) {
	qdisc := NewQdiscCodel()

	id := qdisc.ID()
	assert.Equal(t, "qdisc_codel", id)
}

// TestNewQdiscCodelCeMark 测试 newQdiscCodelCeMark 构造函数
func TestNewQdiscCodelCeMark(t *testing.T) {
	ceMark := newQdiscCodelCeMark()

	require.NotNil(t, ceMark)
	require.NotNil(t, ceMark.baseMetrics)
	assert.Equal(t, "qdisc_codel_ce_mark", ceMark.ID())
}

// TestNewQdiscCodelCount 测试 newQdiscCodelCount 构造函数
func TestNewQdiscCodelCount(t *testing.T) {
	count := newQdiscCodelCount()

	require.NotNil(t, count)
	require.NotNil(t, count.baseMetrics)
	assert.Equal(t, "qdisc_codel_count", count.ID())
}

// TestNewQdiscCodelDropNext 测试 newQdiscCodelDropNext 构造函数
func TestNewQdiscCodelDropNext(t *testing.T) {
	dropNext := newQdiscCodelDropNext()

	require.NotNil(t, dropNext)
	require.NotNil(t, dropNext.baseMetrics)
	assert.Equal(t, "qdisc_codel_drop_next", dropNext.ID())
}

// TestNewQdiscCodelDropOverlimit 测试 newQdiscCodelDropOverlimit 构造函数
func TestNewQdiscCodelDropOverlimit(t *testing.T) {
	dropOverlimit := newQdiscCodelDropOverlimit()

	require.NotNil(t, dropOverlimit)
	require.NotNil(t, dropOverlimit.baseMetrics)
	assert.Equal(t, "qdisc_codel_drop_overlimit", dropOverlimit.ID())
}

// TestNewQdiscCodelDropping 测试 newQdiscCodelDropping 构造函数
func TestNewQdiscCodelDropping(t *testing.T) {
	dropping := newQdiscCodelDropping()

	require.NotNil(t, dropping)
	require.NotNil(t, dropping.baseMetrics)
	assert.Equal(t, "qdisc_codel_dropping", dropping.ID())
}

// TestNewQdiscCodelEcnMark 测试 newQdiscCodelEcnMark 构造函数
func TestNewQdiscCodelEcnMark(t *testing.T) {
	ecnMark := newQdiscCodelEcnMark()

	require.NotNil(t, ecnMark)
	require.NotNil(t, ecnMark.baseMetrics)
	assert.Equal(t, "qdisc_codel_ecn_mark", ecnMark.ID())
}

// TestNewQdiscCodelLdelay 测试 newQdiscCodelLdelay 构造函数
func TestNewQdiscCodelLdelay(t *testing.T) {
	ldelay := newQdiscCodelLdelay()

	require.NotNil(t, ldelay)
	require.NotNil(t, ldelay.baseMetrics)
	assert.Equal(t, "qdisc_codel_ldelay", ldelay.ID())
}

// TestNewQdiscCodelMaxPacket 测试 newQdiscCodelMaxPacket 构造函数
func TestNewQdiscCodelMaxPacket(t *testing.T) {
	maxPacket := newQdiscCodelMaxPacket()

	require.NotNil(t, maxPacket)
	require.NotNil(t, maxPacket.baseMetrics)
	assert.Equal(t, "qdisc_codel_max_packet", maxPacket.ID())
}

// TestQdiscCodelCeMark_Collect 测试 qdiscCodelCeMark 的 Collect 方法
func TestQdiscCodelCeMark_Collect(t *testing.T) {
	ceMark := newQdiscCodelCeMark()
	ch := make(chan prometheus.Metric, 1)

	ceMark.Collect(ch, 100.0, []string{"test-ns", "eth0", "codel"})

	select {
	case metric := <-ch:
		desc := metric.Desc()
		assert.Contains(t, desc.String(), "qdisc_codel_ce_mark")
	default:
		t.Fatal("Expected metric to be collected")
	}
}

// TestQdiscCodelCount_Collect 测试 qdiscCodelCount 的 Collect 方法
func TestQdiscCodelCount_Collect(t *testing.T) {
	count := newQdiscCodelCount()
	ch := make(chan prometheus.Metric, 1)

	count.Collect(ch, 50.0, []string{"test-ns", "eth0", "codel"})

	select {
	case metric := <-ch:
		desc := metric.Desc()
		assert.Contains(t, desc.String(), "qdisc_codel_count")
	default:
		t.Fatal("Expected metric to be collected")
	}
}

// TestQdiscCodelDropNext_Collect 测试 qdiscCodelDropNext 的 Collect 方法
func TestQdiscCodelDropNext_Collect(t *testing.T) {
	dropNext := newQdiscCodelDropNext()
	ch := make(chan prometheus.Metric, 1)

	dropNext.Collect(ch, 25.0, []string{"test-ns", "eth0", "codel"})

	select {
	case metric := <-ch:
		desc := metric.Desc()
		assert.Contains(t, desc.String(), "qdisc_codel_drop_next")
	default:
		t.Fatal("Expected metric to be collected")
	}
}

// TestQdiscCodelDropOverlimit_Collect 测试 qdiscCodelDropOverlimit 的 Collect 方法
func TestQdiscCodelDropOverlimit_Collect(t *testing.T) {
	dropOverlimit := newQdiscCodelDropOverlimit()
	ch := make(chan prometheus.Metric, 1)

	dropOverlimit.Collect(ch, 10.0, []string{"test-ns", "eth0", "codel"})

	select {
	case metric := <-ch:
		desc := metric.Desc()
		assert.Contains(t, desc.String(), "qdisc_codel_drop_overlimit")
	default:
		t.Fatal("Expected metric to be collected")
	}
}

// TestQdiscCodelDropping_Collect 测试 qdiscCodelDropping 的 Collect 方法
func TestQdiscCodelDropping_Collect(t *testing.T) {
	dropping := newQdiscCodelDropping()
	ch := make(chan prometheus.Metric, 1)

	dropping.Collect(ch, 5.0, []string{"test-ns", "eth0", "codel"})

	select {
	case metric := <-ch:
		desc := metric.Desc()
		assert.Contains(t, desc.String(), "qdisc_codel_dropping")
	default:
		t.Fatal("Expected metric to be collected")
	}
}

// TestQdiscCodelEcnMark_Collect 测试 qdiscCodelEcnMark 的 Collect 方法
func TestQdiscCodelEcnMark_Collect(t *testing.T) {
	ecnMark := newQdiscCodelEcnMark()
	ch := make(chan prometheus.Metric, 1)

	ecnMark.Collect(ch, 15.0, []string{"test-ns", "eth0", "codel"})

	select {
	case metric := <-ch:
		desc := metric.Desc()
		assert.Contains(t, desc.String(), "qdisc_codel_ecn_mark")
	default:
		t.Fatal("Expected metric to be collected")
	}
}

// TestQdiscCodelLdelay_Collect 测试 qdiscCodelLdelay 的 Collect 方法
func TestQdiscCodelLdelay_Collect(t *testing.T) {
	ldelay := newQdiscCodelLdelay()
	ch := make(chan prometheus.Metric, 1)

	ldelay.Collect(ch, 200.0, []string{"test-ns", "eth0", "codel"})

	select {
	case metric := <-ch:
		desc := metric.Desc()
		assert.Contains(t, desc.String(), "qdisc_codel_ldelay")
	default:
		t.Fatal("Expected metric to be collected")
	}
}

// TestQdiscCodelMaxPacket_Collect 测试 qdiscCodelMaxPacket 的 Collect 方法
func TestQdiscCodelMaxPacket_Collect(t *testing.T) {
	maxPacket := newQdiscCodelMaxPacket()
	ch := make(chan prometheus.Metric, 1)

	maxPacket.Collect(ch, 1500.0, []string{"test-ns", "eth0", "codel"})

	select {
	case metric := <-ch:
		desc := metric.Desc()
		assert.Contains(t, desc.String(), "qdisc_codel_max_packet")
	default:
		t.Fatal("Expected metric to be collected")
	}
}

// TestQdiscCodel_Collect_EmptyChannel 测试 Collect 方法在空通道上的行为
func TestQdiscCodel_Collect_EmptyChannel(t *testing.T) {
	qdisc := NewQdiscCodel()
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

// TestQdiscCodel_Collect_WithBufferedChannel 测试 Collect 方法在有缓冲通道上的行为
func TestQdiscCodel_Collect_WithBufferedChannel(t *testing.T) {
	qdisc := NewQdiscCodel()
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

// TestQdiscCodel_StructFields 测试结构体字段的初始化
func TestQdiscCodel_StructFields(t *testing.T) {
	qdisc := NewQdiscCodel()

	// 验证所有字段都被正确初始化
	assert.NotNil(t, qdisc.qdiscCodelCeMark)
	assert.NotNil(t, qdisc.qdiscCodelCount)
	assert.NotNil(t, qdisc.qdiscCodelDropNext)
	assert.NotNil(t, qdisc.qdiscCodelDropOverlimit)
	assert.NotNil(t, qdisc.qdiscCodelDropping)
	assert.NotNil(t, qdisc.qdiscCodelEcnMark)
	assert.NotNil(t, qdisc.qdiscCodelLdelay)
	assert.NotNil(t, qdisc.qdiscCodelMaxPacket)

	// 验证字段类型
	assert.IsType(t, qdiscCodelCeMark{}, qdisc.qdiscCodelCeMark)
	assert.IsType(t, qdiscCodelCount{}, qdisc.qdiscCodelCount)
	assert.IsType(t, qdiscCodelDropNext{}, qdisc.qdiscCodelDropNext)
	assert.IsType(t, qdiscCodelDropOverlimit{}, qdisc.qdiscCodelDropOverlimit)
	assert.IsType(t, qdiscCodelDropping{}, qdisc.qdiscCodelDropping)
	assert.IsType(t, qdiscCodelEcnMark{}, qdisc.qdiscCodelEcnMark)
	assert.IsType(t, qdiscCodelLdelay{}, qdisc.qdiscCodelLdelay)
	assert.IsType(t, qdiscCodelMaxPacket{}, qdisc.qdiscCodelMaxPacket)
}

// TestQdiscCodel_ConcurrentAccess 测试并发访问
func TestQdiscCodel_ConcurrentAccess(t *testing.T) {
	qdisc := NewQdiscCodel()
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

// TestQdiscCodel_MetricValues 测试不同指标值的收集
func TestQdiscCodel_MetricValues(t *testing.T) {
	testCases := []struct {
		name     string
		value    float64
		labels   []string
		expected string
	}{
		{
			name:     "zero value",
			value:    0.0,
			labels:   []string{"ns1", "eth0", "codel"},
			expected: "qdisc_codel_ce_mark",
		},
		{
			name:     "positive value",
			value:    1000.0,
			labels:   []string{"ns1", "eth0", "codel"},
			expected: "qdisc_codel_ce_mark",
		},
		{
			name:     "negative value",
			value:    -100.0,
			labels:   []string{"ns1", "eth0", "codel"},
			expected: "qdisc_codel_ce_mark",
		},
		{
			name:     "large value",
			value:    1e10,
			labels:   []string{"ns1", "eth0", "codel"},
			expected: "qdisc_codel_ce_mark",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ceMark := newQdiscCodelCeMark()
			ch := make(chan prometheus.Metric, 1)

			ceMark.Collect(ch, tc.value, tc.labels)

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

// TestQdiscCodel_LabelVariations 测试不同标签组合
func TestQdiscCodel_LabelVariations(t *testing.T) {
	testCases := []struct {
		name   string
		labels []string
	}{
		{
			name:   "standard labels",
			labels: []string{"default", "eth0", "codel"},
		},
		{
			name:   "custom namespace",
			labels: []string{"custom-ns", "eth1", "codel"},
		},
		{
			name:   "different device",
			labels: []string{"default", "wlan0", "codel"},
		},
		{
			name:   "empty namespace",
			labels: []string{"", "eth0", "codel"},
		},
		{
			name:   "empty device",
			labels: []string{"default", "", "codel"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ceMark := newQdiscCodelCeMark()
			ch := make(chan prometheus.Metric, 1)

			ceMark.Collect(ch, 100.0, tc.labels)

			select {
			case metric := <-ch:
				desc := metric.Desc()
				assert.Contains(t, desc.String(), "qdisc_codel_ce_mark")
			default:
				t.Fatal("Expected metric to be collected")
			}
		})
	}
}

// TestQdiscCodel_EdgeCases 测试边界情况
func TestQdiscCodel_EdgeCases(t *testing.T) {
	t.Run("zero value", func(t *testing.T) {
		ceMark := newQdiscCodelCeMark()
		ch := make(chan prometheus.Metric, 1)

		// 测试零值
		ceMark.Collect(ch, 0.0, []string{"ns", "eth0", "codel"})

		select {
		case metric := <-ch:
			desc := metric.Desc()
			assert.Contains(t, desc.String(), "qdisc_codel_ce_mark")
		default:
			t.Fatal("Expected metric to be collected")
		}
	})

	t.Run("negative value", func(t *testing.T) {
		ceMark := newQdiscCodelCeMark()
		ch := make(chan prometheus.Metric, 1)

		// 测试负值
		ceMark.Collect(ch, -100.0, []string{"ns", "eth0", "codel"})

		select {
		case metric := <-ch:
			desc := metric.Desc()
			assert.Contains(t, desc.String(), "qdisc_codel_ce_mark")
		default:
			t.Fatal("Expected metric to be collected")
		}
	})

	t.Run("large value", func(t *testing.T) {
		ceMark := newQdiscCodelCeMark()
		ch := make(chan prometheus.Metric, 1)

		// 测试大值
		ceMark.Collect(ch, 1e10, []string{"ns", "eth0", "codel"})

		select {
		case metric := <-ch:
			desc := metric.Desc()
			assert.Contains(t, desc.String(), "qdisc_codel_ce_mark")
		default:
			t.Fatal("Expected metric to be collected")
		}
	})
}

// TestQdiscCodel_AllMetricTypes 测试所有指标类型的收集
func TestQdiscCodel_AllMetricTypes(t *testing.T) {
	metricTypes := []struct {
		name      string
		value     float64
		expected  string
		collector func() interface {
			Collect(ch chan<- prometheus.Metric, value float64, labels []string)
		}
	}{
		{
			name:     "ce_mark",
			value:    100.0,
			expected: "qdisc_codel_ce_mark",
			collector: func() interface {
				Collect(ch chan<- prometheus.Metric, value float64, labels []string)
			} {
				return newQdiscCodelCeMark()
			},
		},
		{
			name:     "count",
			value:    50.0,
			expected: "qdisc_codel_count",
			collector: func() interface {
				Collect(ch chan<- prometheus.Metric, value float64, labels []string)
			} {
				return newQdiscCodelCount()
			},
		},
		{
			name:     "drop_next",
			value:    25.0,
			expected: "qdisc_codel_drop_next",
			collector: func() interface {
				Collect(ch chan<- prometheus.Metric, value float64, labels []string)
			} {
				return newQdiscCodelDropNext()
			},
		},
		{
			name:     "drop_overlimit",
			value:    10.0,
			expected: "qdisc_codel_drop_overlimit",
			collector: func() interface {
				Collect(ch chan<- prometheus.Metric, value float64, labels []string)
			} {
				return newQdiscCodelDropOverlimit()
			},
		},
		{
			name:     "dropping",
			value:    5.0,
			expected: "qdisc_codel_dropping",
			collector: func() interface {
				Collect(ch chan<- prometheus.Metric, value float64, labels []string)
			} {
				return newQdiscCodelDropping()
			},
		},
		{
			name:     "ecn_mark",
			value:    15.0,
			expected: "qdisc_codel_ecn_mark",
			collector: func() interface {
				Collect(ch chan<- prometheus.Metric, value float64, labels []string)
			} {
				return newQdiscCodelEcnMark()
			},
		},
		{
			name:     "ldelay",
			value:    200.0,
			expected: "qdisc_codel_ldelay",
			collector: func() interface {
				Collect(ch chan<- prometheus.Metric, value float64, labels []string)
			} {
				return newQdiscCodelLdelay()
			},
		},
		{
			name:     "max_packet",
			value:    1500.0,
			expected: "qdisc_codel_max_packet",
			collector: func() interface {
				Collect(ch chan<- prometheus.Metric, value float64, labels []string)
			} {
				return newQdiscCodelMaxPacket()
			},
		},
	}

	for _, mt := range metricTypes {
		t.Run(mt.name, func(t *testing.T) {
			collector := mt.collector()
			ch := make(chan prometheus.Metric, 1)

			collector.Collect(ch, mt.value, []string{"ns", "eth0", "codel"})

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

// TestQdiscCodel_MultipleValues 测试多个值的同时收集
func TestQdiscCodel_MultipleValues(t *testing.T) {
	ch := make(chan prometheus.Metric, 8)

	// 收集所有8种指标类型
	ceMark := newQdiscCodelCeMark()
	count := newQdiscCodelCount()
	dropNext := newQdiscCodelDropNext()
	dropOverlimit := newQdiscCodelDropOverlimit()
	dropping := newQdiscCodelDropping()
	ecnMark := newQdiscCodelEcnMark()
	ldelay := newQdiscCodelLdelay()
	maxPacket := newQdiscCodelMaxPacket()

	ceMark.Collect(ch, 100.0, []string{"ns", "eth0", "codel"})
	count.Collect(ch, 50.0, []string{"ns", "eth0", "codel"})
	dropNext.Collect(ch, 25.0, []string{"ns", "eth0", "codel"})
	dropOverlimit.Collect(ch, 10.0, []string{"ns", "eth0", "codel"})
	dropping.Collect(ch, 5.0, []string{"ns", "eth0", "codel"})
	ecnMark.Collect(ch, 15.0, []string{"ns", "eth0", "codel"})
	ldelay.Collect(ch, 200.0, []string{"ns", "eth0", "codel"})
	maxPacket.Collect(ch, 1500.0, []string{"ns", "eth0", "codel"})

	// 验证收集到了8个指标
	metrics := make([]prometheus.Metric, 0, 8)
	for i := 0; i < 8; i++ {
		select {
		case metric := <-ch:
			metrics = append(metrics, metric)
		default:
			t.Fatalf("Expected metric %d to be collected", i)
		}
	}

	assert.Len(t, metrics, 8)

	// 验证每个指标都有正确的描述符
	expectedNames := []string{
		"qdisc_codel_ce_mark",
		"qdisc_codel_count",
		"qdisc_codel_drop_next",
		"qdisc_codel_drop_overlimit",
		"qdisc_codel_dropping",
		"qdisc_codel_ecn_mark",
		"qdisc_codel_ldelay",
		"qdisc_codel_max_packet",
	}

	for i, metric := range metrics {
		desc := metric.Desc()
		assert.Contains(t, desc.String(), expectedNames[i])
	}
}

// TestQdiscCodel_Performance 测试性能相关场景
func TestQdiscCodel_Performance(t *testing.T) {
	ceMark := newQdiscCodelCeMark()
	ch := make(chan prometheus.Metric, 1000)

	// 测试大量指标收集的性能
	start := time.Now()
	for i := 0; i < 1000; i++ {
		ceMark.Collect(ch, float64(i), []string{"ns", "eth0", "codel"})
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

// TestQdiscCodel_ConstructorConsistency 测试构造函数的一致性
func TestQdiscCodel_ConstructorConsistency(t *testing.T) {
	// 测试多次调用构造函数返回的对象是否一致
	qdisc1 := NewQdiscCodel()
	qdisc2 := NewQdiscCodel()

	// 验证两个对象都有相同的结构
	assert.Equal(t, qdisc1.ID(), qdisc2.ID())
	assert.NotNil(t, qdisc1.qdiscCodelCeMark)
	assert.NotNil(t, qdisc2.qdiscCodelCeMark)
	assert.NotNil(t, qdisc1.qdiscCodelCount)
	assert.NotNil(t, qdisc2.qdiscCodelCount)
	assert.NotNil(t, qdisc1.qdiscCodelDropNext)
	assert.NotNil(t, qdisc2.qdiscCodelDropNext)
	assert.NotNil(t, qdisc1.qdiscCodelDropOverlimit)
	assert.NotNil(t, qdisc2.qdiscCodelDropOverlimit)
	assert.NotNil(t, qdisc1.qdiscCodelDropping)
	assert.NotNil(t, qdisc2.qdiscCodelDropping)
	assert.NotNil(t, qdisc1.qdiscCodelEcnMark)
	assert.NotNil(t, qdisc2.qdiscCodelEcnMark)
	assert.NotNil(t, qdisc1.qdiscCodelLdelay)
	assert.NotNil(t, qdisc2.qdiscCodelLdelay)
	assert.NotNil(t, qdisc1.qdiscCodelMaxPacket)
	assert.NotNil(t, qdisc2.qdiscCodelMaxPacket)
}

// TestQdiscCodel_MetricDescriptions 测试指标描述的正确性
func TestQdiscCodel_MetricDescriptions(t *testing.T) {
	metricTests := []struct {
		name        string
		constructor func() interface{ ID() string }
		expectedID  string
	}{
		{"ce_mark", func() interface{ ID() string } { return newQdiscCodelCeMark() }, "qdisc_codel_ce_mark"},
		{"count", func() interface{ ID() string } { return newQdiscCodelCount() }, "qdisc_codel_count"},
		{"drop_next", func() interface{ ID() string } { return newQdiscCodelDropNext() }, "qdisc_codel_drop_next"},
		{"drop_overlimit", func() interface{ ID() string } { return newQdiscCodelDropOverlimit() }, "qdisc_codel_drop_overlimit"},
		{"dropping", func() interface{ ID() string } { return newQdiscCodelDropping() }, "qdisc_codel_dropping"},
		{"ecn_mark", func() interface{ ID() string } { return newQdiscCodelEcnMark() }, "qdisc_codel_ecn_mark"},
		{"ldelay", func() interface{ ID() string } { return newQdiscCodelLdelay() }, "qdisc_codel_ldelay"},
		{"max_packet", func() interface{ ID() string } { return newQdiscCodelMaxPacket() }, "qdisc_codel_max_packet"},
	}

	for _, mt := range metricTests {
		t.Run(mt.name, func(t *testing.T) {
			metric := mt.constructor()
			assert.Equal(t, mt.expectedID, metric.ID())
		})
	}
}

// TestQdiscCodel_ChannelHandling 测试通道处理的各种情况
func TestQdiscCodel_ChannelHandling(t *testing.T) {
	t.Run("closed channel", func(t *testing.T) {
		ceMark := newQdiscCodelCeMark()
		ch := make(chan prometheus.Metric, 1)
		close(ch)

		// 测试向已关闭的通道发送数据时的行为
		defer func() {
			if r := recover(); r != nil {
				// 预期会 panic，因为向已关闭的通道发送数据会 panic
				assert.Contains(t, r.(error).Error(), "send on closed channel")
			}
		}()

		ceMark.Collect(ch, 100.0, []string{"ns", "eth0", "codel"})
	})

	t.Run("buffered channel", func(t *testing.T) {
		ceMark := newQdiscCodelCeMark()
		ch := make(chan prometheus.Metric, 5)

		// 测试向有缓冲的通道发送多个指标
		for i := 0; i < 5; i++ {
			ceMark.Collect(ch, float64(i), []string{"ns", "eth0", "codel"})
		}

		// 验证所有指标都被发送
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
	})
}

// TestQdiscCodel_LabelValidation 测试标签验证
func TestQdiscCodel_LabelValidation(t *testing.T) {
	ceMark := newQdiscCodelCeMark()
	ch := make(chan prometheus.Metric, 1)

	// 测试不同长度的标签数组
	testCases := []struct {
		name        string
		labels      []string
		shouldPanic bool
	}{
		{
			name:        "correct labels",
			labels:      []string{"ns", "eth0", "codel"},
			shouldPanic: false,
		},
		{
			name:        "too few labels",
			labels:      []string{"ns", "eth0"},
			shouldPanic: true,
		},
		{
			name:        "too many labels",
			labels:      []string{"ns", "eth0", "codel", "extra"},
			shouldPanic: true,
		},
		{
			name:        "empty labels",
			labels:      []string{},
			shouldPanic: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				defer func() {
					if r := recover(); r != nil {
						// 预期会 panic
						assert.NotNil(t, r)
					} else {
						t.Error("Expected panic but didn't get one")
					}
				}()
			}

			ceMark.Collect(ch, 100.0, tc.labels)

			if !tc.shouldPanic {
				select {
				case metric := <-ch:
					desc := metric.Desc()
					assert.Contains(t, desc.String(), "qdisc_codel_ce_mark")
				default:
					t.Fatal("Expected metric to be collected")
				}
			}
		})
	}
}
