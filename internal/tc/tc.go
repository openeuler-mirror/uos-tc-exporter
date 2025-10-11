// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

// Package tc 提供了 Linux Traffic Control (TC) 的操作接口
package tc

import (
	"github.com/florianl/go-tc"
	"github.com/jsimonetti/rtnetlink"
)

// GetNetlinkConn 获取指定网络命名空间的 rtnetlink 连接
//
// 参数：
//   - ns: 网络命名空间名称
//
// 返回：
//   - *rtnetlink.Conn: 网络连接
//   - error: 如果连接失败则返回错误
func GetNetlinkConn(ns string) (*rtnetlink.Conn, error) {
	cm := NewConnectionManager(ns)
	return cm.GetNetlinkConn()
}

// GetTcConn 获取指定网络命名空间的 TC 连接
//
// 参数：
//   - ns: 网络命名空间名称
//
// 返回：
//   - *tc.Tc: TC 连接
//   - error: 如果连接失败则返回错误
func GetTcConn(ns string) (*tc.Tc, error) {
	cm := NewConnectionManager(ns)
	return cm.GetTcConn()
}

// GetQdiscs 获取指定接口在指定网络命名空间中的所有 qdisc
//
// 参数：
//   - devID: 网络接口索引
//   - ns: 网络命名空间名称
//
// 返回：
//   - []tc.Object: qdisc 对象列表
//   - error: 如果获取失败则返回错误
func GetQdiscs(devID uint32, ns string) ([]tc.Object, error) {
	collector := NewTcObjectCollector(ns)
	return collector.GetQdiscs(devID)
}

// GetClasses 获取指定接口在指定网络命名空间中的所有 class
//
// 参数：
//   - devID: 网络接口索引
//   - ns: 网络命名空间名称
//
// 返回：
//   - []tc.Object: class 对象列表
//   - error: 如果获取失败则返回错误
func GetClasses(devID uint32, ns string) ([]tc.Object, error) {
	collector := NewTcObjectCollector(ns)
	return collector.GetClasses(devID)
}

// GetFilters 获取指定接口在指定网络命名空间中的所有 filter
//
// 参数：
//   - devID: 网络接口索引
//   - ns: 网络命名空间名称
//
// 返回：
//   - []tc.Object: filter 对象列表
//   - error: 如果获取失败则返回错误
func GetFilters(devID uint32, ns string) ([]tc.Object, error) {
	collector := NewTcObjectCollector(ns)
	return collector.GetFilters(devID)
}

// GetNetNameSpaceList 获取系统中所有网络命名空间的名称列表
//
// 返回：
//   - []string: 网络命名空间名称列表
//   - error: 如果获取失败则返回错误
//
// 注意：此函数保持向后兼容性，建议使用 GetNetworkNamespaceNames
func GetNetNameSpaceList() ([]string, error) {
	return GetNetworkNamespaceNames()
}

// GetInterfaceInNetNS 获取指定网络命名空间中的所有网络接口
//
// 参数：
//   - ns: 网络命名空间名称
//
// 返回：
//   - []rtnetlink.LinkMessage: 网络接口列表（排除回环接口）
//   - error: 如果获取失败则返回错误
//
// 注意：此函数保持向后兼容性，建议使用 GetInterfacesInNamespace
func GetInterfaceInNetNS(ns string) ([]rtnetlink.LinkMessage, error) {
	return GetInterfacesInNamespace(ns)
}
