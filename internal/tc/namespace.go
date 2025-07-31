// Package tc 提供了 Linux Traffic Control (TC) 的操作接口
//
// 该包封装了 TC 相关的网络操作，包括网络命名空间管理、
// 接口查询、qdisc/class/filter 操作等功能。
package tc

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/jsimonetti/rtnetlink"
	"github.com/sirupsen/logrus"
)

const (
	// NetNSDir 网络命名空间目录路径
	NetNSDir = "/var/run/netns"
	// DefaultNetNS 默认网络命名空间名称
	DefaultNetNS = "default"
)

// NetworkNamespace 表示一个网络命名空间
type NetworkNamespace struct {
	Name string
	Path string
}

// NewNetworkNamespace 创建一个网络命名空间实例
func NewNetworkNamespace(name string) *NetworkNamespace {
	ns := &NetworkNamespace{
		Name: name,
	}

	if name == DefaultNetNS {
		ns.Path = ""
	} else {
		ns.Path = filepath.Join(NetNSDir, name)
	}

	return ns
}

// Exists 检查网络命名空间是否存在
func (ns *NetworkNamespace) Exists() bool {
	if ns.Name == DefaultNetNS {
		return true
	}

	_, err := os.Stat(ns.Path)
	return err == nil
}

// GetFileDescriptor 获取网络命名空间的文件描述符
func (ns *NetworkNamespace) GetFileDescriptor() (*os.File, error) {
	if ns.Name == DefaultNetNS {
		return nil, nil
	}

	file, err := os.Open(ns.Path)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// GetNetworkNamespaces 获取系统中所有可用的网络命名空间
//
// 返回：
//   - []*NetworkNamespace: 网络命名空间列表，包含默认命名空间
//   - error: 如果读取目录失败则返回错误
func GetNetworkNamespaces() ([]*NetworkNamespace, error) {
	// 尝试读取网络命名空间目录
	files, err := os.ReadDir(NetNSDir)
	if err != nil {
		logrus.Debugf("Failed to read network namespace directory: %v", err)

		// 如果目录不存在，只返回默认命名空间
		if errors.Is(err, os.ErrNotExist) {
			logrus.Debug("Network namespace directory does not exist, returning default namespace only")
			return []*NetworkNamespace{NewNetworkNamespace(DefaultNetNS)}, nil
		}

		return nil, err
	}

	// 收集所有命名空间
	namespaces := make([]*NetworkNamespace, 0, len(files)+1)

	// 添加用户创建的命名空间
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		nsName := file.Name()
		ns := NewNetworkNamespace(nsName)

		// 验证命名空间是否有效
		if ns.Exists() {
			namespaces = append(namespaces, ns)
			logrus.Debugf("Found network namespace: %s", nsName)
		} else {
			logrus.Warnf("Network namespace file exists but is invalid: %s", nsName)
		}
	}

	// 添加默认命名空间
	defaultNS := NewNetworkNamespace(DefaultNetNS)
	namespaces = append(namespaces, defaultNS)
	logrus.Debugf("Added default network namespace: %s", DefaultNetNS)

	logrus.Debugf("Total network namespaces found: %d", len(namespaces))
	return namespaces, nil
}

// GetNetworkNamespaceNames 获取所有网络命名空间的名称列表
//
// 返回：
//   - []string: 网络命名空间名称列表
//   - error: 如果获取失败则返回错误
func GetNetworkNamespaceNames() ([]string, error) {
	namespaces, err := GetNetworkNamespaces()
	if err != nil {
		return nil, err
	}

	names := make([]string, len(namespaces))
	for i, ns := range namespaces {
		names[i] = ns.Name
	}

	return names, nil
}

// GetInterfacesInNamespace 获取指定网络命名空间中的所有网络接口
//
// 参数：
//   - nsName: 网络命名空间名称
//
// 返回：
//   - []rtnetlink.LinkMessage: 网络接口列表（排除回环接口）
//   - error: 如果获取失败则返回错误
func GetInterfacesInNamespace(nsName string) ([]rtnetlink.LinkMessage, error) {
	// 获取网络连接
	conn, err := GetNetlinkConn(nsName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// 获取所有网络接口
	links, err := conn.Link.List()
	if err != nil {
		return nil, err
	}

	// 过滤掉回环接口（通常是第一个接口）
	var interfaces []rtnetlink.LinkMessage
	for i, link := range links {
		// 跳过回环接口（index 1 通常是 lo）
		if i == 0 && link.Index == 1 {
			logrus.Debug("Skipping loopback interface")
			continue
		}

		interfaces = append(interfaces, link)
		logrus.Debugf("Found interface in namespace %s: %s (index: %d)",
			nsName, link.Attributes.Name, link.Index)
	}

	logrus.Debugf("Found %d interfaces in namespace %s", len(interfaces), nsName)
	return interfaces, nil
}

// ValidateNamespace 验证网络命名空间是否有效
//
// 参数：
//   - nsName: 网络命名空间名称
//
// 返回：
//   - bool: 如果命名空间有效则返回 true
func ValidateNamespace(nsName string) bool {
	ns := NewNetworkNamespace(nsName)
	return ns.Exists()
}
