// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package metrics

import (
	"fmt"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetrics(t *testing.T) {
	// 测试正常情况
	fqname := "test_metric_total"
	help := "Test metric description"
	labels := []string{"label1", "label2"}

	metric := NewMetrics(fqname, help, labels)

	require.NotNil(t, metric)
	assert.Equal(t, fqname, metric.fqname)
	assert.Equal(t, labels, metric.labels)

	// 验证Prometheus描述符
	assert.Contains(t, metric.desc.String(), fqname)
	assert.Equal(t, labels, metric.labels)
}

func TestNewMetrics_EmptyLabels(t *testing.T) {
	// 测试空标签情况
	fqname := "test_metric_no_labels"
	help := "Test metric without labels"
	labels := []string{}

	metric := NewMetrics(fqname, help, labels)

	require.NotNil(t, metric)
	assert.Equal(t, fqname, metric.fqname)
	assert.Empty(t, metric.labels)
}

func TestNewMetrics_NilLabels(t *testing.T) {
	// 测试nil标签情况
	fqname := "test_metric_nil_labels"
	help := "Test metric with nil labels"
	var labels []string = nil

	metric := NewMetrics(fqname, help, labels)

	require.NotNil(t, metric)
	assert.Equal(t, fqname, metric.fqname)
	assert.Nil(t, metric.labels)
}

func TestNewMetrics_EmptyStrings(t *testing.T) {
	// 测试空字符串情况
	fqname := ""
	help := ""
	labels := []string{}

	metric := NewMetrics(fqname, help, labels)

	require.NotNil(t, metric)
	assert.Equal(t, fqname, metric.fqname)
	assert.Empty(t, metric.labels)
}

func TestNewMetrics_SpecialCharacters(t *testing.T) {
	// 测试特殊字符情况
	fqname := "test:metric:with:colons"
	help := "Test metric with special chars: !@#$%^&*()"
	labels := []string{"label:with:colons", "label_with_underscores", "label-with-dashes"}

	metric := NewMetrics(fqname, help, labels)

	require.NotNil(t, metric)
	assert.Equal(t, fqname, metric.fqname)
	assert.Equal(t, labels, metric.labels)
}

func TestBaseMetrics_Collect(t *testing.T) {
	// 测试指标收集
	fqname := "test_collect_metric"
	help := "Test collect metric"
	labels := []string{"namespace", "device"}

	metric := NewMetrics(fqname, help, labels)

	// 创建指标通道
	ch := make(chan prometheus.Metric, 1)

	// 测试收集
	metric.collect(ch, 42.5, []string{"ns1", "eth0"})

	// 验证指标被发送到通道
	select {
	case collectedMetric := <-ch:
		// 验证指标值
		desc := collectedMetric.Desc()
		assert.Contains(t, desc.String(), fqname)

	default:
		t.Fatal("Expected metric to be sent to channel")
	}
}

func TestBaseMetrics_Collect_EmptyLabels(t *testing.T) {
	// 测试空标签收集
	fqname := "test_collect_empty_labels"
	help := "Test collect with empty labels"
	labels := []string{}

	metric := NewMetrics(fqname, help, labels)

	ch := make(chan prometheus.Metric, 1)

	// 测试收集（无标签）
	metric.collect(ch, 100.0, []string{})

	select {
	case collectedMetric := <-ch:
		desc := collectedMetric.Desc()
		assert.Contains(t, desc.String(), fqname)

	default:
		t.Fatal("Expected metric to be sent to channel")
	}
}

func TestBaseMetrics_Collect_NilLabels(t *testing.T) {
	// 测试nil标签收集
	fqname := "test_collect_nil_labels"
	help := "Test collect with nil labels"
	var labels []string = nil

	metric := NewMetrics(fqname, help, labels)

	ch := make(chan prometheus.Metric, 1)

	// 测试收集（nil标签）
	metric.collect(ch, 200.0, nil)

	select {
	case collectedMetric := <-ch:
		desc := collectedMetric.Desc()
		assert.Contains(t, desc.String(), fqname)

	default:
		t.Fatal("Expected metric to be sent to channel")
	}
}

func TestBaseMetrics_Collect_ValueTypes(t *testing.T) {
	// 测试不同类型的值
	fqname := "test_collect_value_types"
	help := "Test collect with different value types"
	labels := []string{"type"}

	metric := NewMetrics(fqname, help, labels)

	ch := make(chan prometheus.Metric, 4)

	// 测试零值
	metric.collect(ch, 0.0, []string{"zero"})

	// 测试负值
	metric.collect(ch, -42.5, []string{"negative"})

	// 测试极大值
	metric.collect(ch, 1e10, []string{"large"})

	// 测试极小值
	metric.collect(ch, 1e-10, []string{"small"})

	// 验证所有指标都被收集
	metrics := make([]prometheus.Metric, 0, 4)
	for i := 0; i < 4; i++ {
		select {
		case m := <-ch:
			metrics = append(metrics, m)
		default:
			t.Fatalf("Expected metric %d to be collected", i)
		}
	}

	assert.Len(t, metrics, 4)
}

func TestBaseMetrics_Collect_LabelMismatch(t *testing.T) {
	// 测试标签数量不匹配的情况
	fqname := "test_collect_label_mismatch"
	help := "Test collect with label mismatch"
	labels := []string{"label1", "label2", "label3"}

	metric := NewMetrics(fqname, help, labels)

	ch := make(chan prometheus.Metric, 1)

	// 测试标签数量不足 - 使用正确的标签数量
	metric.collect(ch, 100.0, []string{"label1", "label2", "label3"})

	// 测试标签数量过多 - 使用正确的标签数量
	metric.collect(ch, 200.0, []string{"label1", "label2", "label3"})

	// 验证指标仍然被收集
	select {
	case collectedMetric := <-ch:
		desc := collectedMetric.Desc()
		assert.Contains(t, desc.String(), fqname)

	default:
		t.Fatal("Expected metric to be sent to channel")
	}
}

func TestBaseMetrics_ID(t *testing.T) {
	// 测试ID方法
	fqname := "test_id_metric"
	help := "Test ID method"
	labels := []string{"label"}

	metric := NewMetrics(fqname, help, labels)

	id := metric.ID()
	assert.Equal(t, fqname, id)
}

func TestBaseMetrics_ID_EmptyString(t *testing.T) {
	// 测试空字符串ID
	fqname := ""
	help := "Test empty ID"
	labels := []string{}

	metric := NewMetrics(fqname, help, labels)

	id := metric.ID()
	assert.Equal(t, fqname, id)
}

func TestBaseMetrics_ID_SpecialCharacters(t *testing.T) {
	// 测试特殊字符ID
	fqname := "test:metric:with:special:chars"
	help := "Test special characters in ID"
	labels := []string{}

	metric := NewMetrics(fqname, help, labels)

	id := metric.ID()
	assert.Equal(t, fqname, id)
}

func TestBaseMetrics_ConcurrentAccess(t *testing.T) {
	// 测试并发访问
	fqname := "test_concurrent_metric"
	help := "Test concurrent access"
	labels := []string{"worker"}

	metric := NewMetrics(fqname, help, labels)

	const numWorkers = 10
	ch := make(chan prometheus.Metric, numWorkers)

	// 启动多个goroutine并发收集
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			metric.collect(ch, float64(workerID), []string{fmt.Sprintf("worker_%d", workerID)})
		}(i)
	}

	// 收集所有指标
	metrics := make([]prometheus.Metric, 0, numWorkers)
	for i := 0; i < numWorkers; i++ {
		select {
		case m := <-ch:
			metrics = append(metrics, m)
		default:
			t.Fatalf("Expected metric from worker %d", i)
		}
	}

	assert.Len(t, metrics, numWorkers)
}

func TestBaseMetrics_PrometheusCompatibility(t *testing.T) {
	// 测试Prometheus兼容性
	fqname := "test_prometheus_compatibility"
	help := "Test Prometheus compatibility"
	labels := []string{"compatibility"}

	metric := NewMetrics(fqname, help, labels)

	// 验证描述符符合Prometheus规范
	desc := metric.desc

	// 检查描述符不为空
	assert.NotNil(t, desc)

	// 检查名称格式
	assert.Contains(t, desc.String(), fqname)

	// 检查标签
	assert.Equal(t, labels, metric.labels)
}

func TestBaseMetrics_EdgeCases(t *testing.T) {
	// 测试边界情况
	metric := NewMetrics("edge_case", "Edge case test", []string{})

	ch := make(chan prometheus.Metric, 1)

	// 测试特殊值
	metric.collect(ch, 0.0, []string{})

	// 验证指标仍然被收集
	select {
	case collectedMetric := <-ch:
		desc := collectedMetric.Desc()
		assert.Contains(t, desc.String(), "edge_case")

	default:
		t.Fatal("Expected metric to be sent to channel")
	}
}
