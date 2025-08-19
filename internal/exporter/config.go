// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package exporter

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
	"time"

	"gitee.com/openeuler/uos-tc-exporter/pkg/logger"
	"gitee.com/openeuler/uos-tc-exporter/pkg/utils"
	"github.com/alecthomas/kingpin"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Logging     logger.Config `yaml:"log"`
	Address     string        `yaml:"address"`
	Port        int           `yaml:"port"`
	MetricsPath string        `yaml:"metricsPath"`
}

var (
	Configfile    *string
	DefaultConfig = Config{
		Logging: logger.Config{
			Level:   "debug",
			LogPath: "/var/log/tc-exporter.log",
			MaxSize: "10MB",
			MaxAge:  time.Hour * 24 * 7},
		Address:     "127.0.0.1",
		Port:        9062,
		MetricsPath: "/metrics",
	}
)

func init() {
	kingpin.HelpFlag.Short('h')
	Configfile = kingpin.Flag("config", "Configuration file").
		Short('c').
		Default("/etc/uos-exporter/tc-exporter.yaml").
		String()
}
func Unpack(config interface{}) error {
	if !utils.FileExists(*Configfile) {
		logrus.Errorf("%s file not found", *Configfile)
		logrus.Debug("Use default config")
	} else {
		file, err := os.Open(*Configfile)
		if err != nil {
			logrus.Error("Failed to open config file: ", err)
			return err
		}
		err = yaml.NewDecoder(file).Decode(config)
		if err != nil {
			return err
		}
	}
	return nil
}

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	var errors []string

	// 验证地址格式
	if err := c.validateAddress(); err != nil {
		errors = append(errors, fmt.Sprintf("address validation failed: %v", err))
	}

	// 验证端口范围
	if err := c.validatePort(); err != nil {
		errors = append(errors, fmt.Sprintf("port validation failed: %v", err))
	}

	// 验证指标路径
	if err := c.validateMetricsPath(); err != nil {
		errors = append(errors, fmt.Sprintf("metrics path validation failed: %v", err))
	}

	// 验证日志配置
	if err := c.validateLogging(); err != nil {
		errors = append(errors, fmt.Sprintf("logging validation failed: %v", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}

// validateAddress 验证地址格式
func (c *Config) validateAddress() error {
	if c.Address == "" {
		return fmt.Errorf("address cannot be empty")
	}

	// 检查是否为有效的IP地址
	if net.ParseIP(c.Address) != nil {
		return nil
	}

	// 检查是否为有效的域名
	if c.isValidDomain(c.Address) {
		return nil
	}

	// 检查是否为有效的网络接口名称
	if c.isValidInterface(c.Address) {
		return nil
	}

	return fmt.Errorf("invalid address format: %s (must be valid IP, domain, or interface name)", c.Address)
}

// validatePort 验证端口范围
func (c *Config) validatePort() error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", c.Port)
	}

	// 检查是否为特权端口（需要root权限）
	if c.Port < 1024 {
		logrus.Warnf("Port %d is a privileged port (< 1024), ensure the application has sufficient permissions", c.Port)
	}

	return nil
}

// validateMetricsPath 验证指标路径
func (c *Config) validateMetricsPath() error {
	if c.MetricsPath == "" {
		return fmt.Errorf("metrics path cannot be empty")
	}

	// 检查路径是否以 / 开头
	if !strings.HasPrefix(c.MetricsPath, "/") {
		return fmt.Errorf("metrics path must start with '/', got: %s", c.MetricsPath)
	}

	// 检查路径是否包含非法字符
	invalidChars := regexp.MustCompile(`[<>:"|?*]`)
	if invalidChars.MatchString(c.MetricsPath) {
		return fmt.Errorf("metrics path contains invalid characters: %s", c.MetricsPath)
	}

	return nil
}

// isValidDomain 检查是否为有效的域名
func (c *Config) isValidDomain(domain string) bool {
	// 简单的域名验证规则
	if len(domain) > 253 {
		return false
	}

	// 检查是否包含至少一个点号（顶级域名）
	if !strings.Contains(domain, ".") {
		return false
	}

	// 检查每个标签的长度
	labels := strings.Split(domain, ".")
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return false
		}
		// 标签只能包含字母、数字和连字符
		if !regexp.MustCompile(`^[a-zA-Z0-9-]+$`).MatchString(label) {
			return false
		}
		// 标签不能以连字符开头或结尾
		if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return false
		}
	}

	return true
}

// isValidInterface 检查是否为有效的网络接口名称
func (c *Config) isValidInterface(iface string) bool {
	// 网络接口名称通常遵循特定规则
	if len(iface) == 0 || len(iface) > 15 {
		return false
	}

	// 接口名称只能包含字母、数字、连字符和下划线
	if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(iface) {
		return false
	}

	// 检查是否为特殊接口名称
	specialInterfaces := []string{"lo", "eth0", "eth1", "wlan0", "docker0", "br0"}
	for _, special := range specialInterfaces {
		if iface == special {
			return true
		}
	}

	// 更严格的验证：必须是常见的接口命名模式
	// 例如：eth0, wlan0, ens33, enp0s3 等
	validPatterns := []*regexp.Regexp{
		regexp.MustCompile(`^eth\d+$`),        // eth0, eth1
		regexp.MustCompile(`^wlan\d+$`),       // wlan0, wlan1
		regexp.MustCompile(`^ens\d+$`),        // ens33, ens34
		regexp.MustCompile(`^enp\d+s\d+$`),    // enp0s3, enp0s8
		regexp.MustCompile(`^docker\d+$`),     // docker0, docker1
		regexp.MustCompile(`^br\d+$`),         // br0, br1
		regexp.MustCompile(`^lo$`),            // loopback
		regexp.MustCompile(`^veth[a-f0-9]+$`), // veth123456
	}

	for _, pattern := range validPatterns {
		if pattern.MatchString(iface) {
			return true
		}
	}

	return false
}

// GetBindAddress 获取完整的绑定地址
func (c *Config) GetBindAddress() string {
	return fmt.Sprintf("%s:%d", c.Address, c.Port)
}

// IsLocalhost 检查是否为本地地址
func (c *Config) IsLocalhost() bool {
	return c.Address == "127.0.0.1" || c.Address == "localhost" || c.Address == "::1"
}

// IsPublic 检查是否为公网地址
func (c *Config) IsPublic() bool {
	if c.IsLocalhost() {
		return false
	}

	ip := net.ParseIP(c.Address)
	if ip == nil {
		return false
	}

	return true
}

// validateLogging 验证日志配置
func (c *Config) validateLogging() error {
	// 验证日志级别
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}

	if c.Logging.Level != "" && !validLevels[strings.ToLower(c.Logging.Level)] {
		return fmt.Errorf("invalid log level: %s, supported levels are: debug, info, warn, error", c.Logging.Level)
	}

	// 验证日志路径
	if c.Logging.LogPath != "" {
		// 检查路径是否包含非法字符
		invalidChars := regexp.MustCompile(`[<>:"|?*]`)
		if invalidChars.MatchString(c.Logging.LogPath) {
			return fmt.Errorf("log path contains invalid characters: %s", c.Logging.LogPath)
		}
	}

	// 验证最大文件大小
	if c.Logging.MaxSize != "" {
		// 这里可以添加更复杂的文件大小验证逻辑
		if len(c.Logging.MaxSize) > 20 {
			return fmt.Errorf("log max size string too long: %s", c.Logging.MaxSize)
		}
	}

	return nil
}
