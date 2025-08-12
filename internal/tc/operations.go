// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

// Package tc 提供了 Linux Traffic Control (TC) 的操作接口
package tc

import (
	"fmt"

	"github.com/florianl/go-tc"
	"github.com/jsimonetti/rtnetlink"
	"github.com/mdlayher/netlink"
	"golang.org/x/sys/unix"
)

// Handle 表示一个 TC 句柄
type Handle struct {
	Major uint32
	Minor uint32
}

// NewHandle 创建一个新的 TC 句柄
func NewHandle(major, minor uint32) Handle {
	return Handle{
		Major: major,
		Minor: minor,
	}
}

// ToUint32 将句柄转换为 uint32 格式
func (h Handle) ToUint32() uint32 {
	return (h.Major << 16) | h.Minor
}

// String 返回句柄的字符串表示
func (h Handle) String() string {
	return fmt.Sprintf("%d:%d", h.Major, h.Minor)
}

// ParseHandle 从 uint32 解析句柄
func ParseHandle(handle uint32) Handle {
	return NewHandle((handle&0xffff0000)>>16, handle&0x0000ffff)
}

// FormatHandle 格式化 TC 句柄为字符串
func FormatHandle(handle uint32) string {
	return ParseHandle(handle).String()
}

// ConnectionManager 管理网络连接
type ConnectionManager struct {
	namespace string
}

// NewConnectionManager 创建一个新的连接管理器
func NewConnectionManager(namespace string) *ConnectionManager {
	return &ConnectionManager{
		namespace: namespace,
	}
}

// getNetlinkConfig 获取网络命名空间配置
func (cm *ConnectionManager) getNetlinkConfig() (*netlink.Config, error) {
	if cm.namespace == DefaultNetNS {
		return nil, nil
	}

	ns := NewNetworkNamespace(cm.namespace)
	if !ns.Exists() {
		return nil, fmt.Errorf("network namespace does not exist: %s", cm.namespace)
	}

	file, err := ns.GetFileDescriptor()
	if err != nil {
		return nil, fmt.Errorf("failed to open network namespace: %w", err)
	}

	return &netlink.Config{NetNS: int(file.Fd())}, nil
}

// GetNetlinkConn 获取 rtnetlink 连接
func (cm *ConnectionManager) GetNetlinkConn() (*rtnetlink.Conn, error) {
	config, err := cm.getNetlinkConfig()
	if err != nil {
		return nil, err
	}

	conn, err := rtnetlink.Dial(config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial rtnetlink: %w", err)
	}

	return conn, nil
}

// GetTcConn 获取 TC 连接
func (cm *ConnectionManager) GetTcConn() (*tc.Tc, error) {
	config, err := cm.getNetlinkConfig()
	if err != nil {
		return nil, err
	}

	var tcConfig *tc.Config
	if config != nil {
		tcConfig = &tc.Config{NetNS: config.NetNS}
	}

	sock, err := tc.Open(tcConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open TC connection: %w", err)
	}

	return sock, nil
}

// TcObjectCollector 收集 TC 对象的收集器
type TcObjectCollector struct {
	connManager *ConnectionManager
}

// NewTcObjectCollector 创建一个新的 TC 对象收集器
func NewTcObjectCollector(namespace string) *TcObjectCollector {
	return &TcObjectCollector{
		connManager: NewConnectionManager(namespace),
	}
}

// collectObjects 收集 TC 对象的通用方法
func (tcoc *TcObjectCollector) collectObjects(
	devID uint32,
	collectFunc func(*tc.Tc) ([]tc.Object, error),
) ([]tc.Object, error) {
	// 获取 TC 连接
	sock, err := tcoc.connManager.GetTcConn()
	if err != nil {
		return nil, err
	}
	defer sock.Close()

	// 收集所有对象
	objects, err := collectFunc(sock)
	if err != nil {
		return nil, err
	}

	// 按接口索引过滤对象
	var result []tc.Object
	for _, obj := range objects {
		if obj.Ifindex == devID {
			result = append(result, obj)
		}
	}

	return result, nil
}

// GetQdiscs 获取指定接口的所有 qdisc
func (tcoc *TcObjectCollector) GetQdiscs(devID uint32) ([]tc.Object, error) {
	return tcoc.collectObjects(devID, func(sock *tc.Tc) ([]tc.Object, error) {
		return sock.Qdisc().Get()
	})
}

// GetClasses 获取指定接口的所有 class
func (tcoc *TcObjectCollector) GetClasses(devID uint32) ([]tc.Object, error) {
	return tcoc.collectObjects(devID, func(sock *tc.Tc) ([]tc.Object, error) {
		return sock.Class().Get(&tc.Msg{
			Family:  unix.AF_UNSPEC,
			Info:    0,
			Handle:  tc.HandleRoot,
			Ifindex: devID,
		})
	})
}

// GetFilters 获取指定接口的所有 filter
func (tcoc *TcObjectCollector) GetFilters(devID uint32) ([]tc.Object, error) {
	return tcoc.collectObjects(devID, func(sock *tc.Tc) ([]tc.Object, error) {
		return sock.Filter().Get(&tc.Msg{
			Family:  unix.AF_UNSPEC,
			Info:    0,
			Handle:  tc.HandleRoot,
			Ifindex: devID,
		})
	})
}
