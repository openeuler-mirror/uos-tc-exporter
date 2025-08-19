// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package exporter

import (
	"strings"
	"testing"
	"time"

	"gitee.com/openeuler/uos-tc-exporter/pkg/logger"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: Config{
				Logging: logger.Config{
					Level:   "info",
					LogPath: "/var/log/test.log",
					MaxSize: "10MB",
					MaxAge:  time.Hour * 24,
				},
				Address:     "127.0.0.1",
				Port:        9062,
				MetricsPath: "/metrics",
			},
			wantErr: false,
		},
		{
			name: "invalid port too low",
			config: Config{
				Logging: logger.Config{
					Level:   "info",
					LogPath: "/var/log/test.log",
					MaxSize: "10MB",
					MaxAge:  time.Hour * 24,
				},
				Address:     "127.0.0.1",
				Port:        0,
				MetricsPath: "/metrics",
			},
			wantErr: true,
			errMsg:  "port validation failed",
		},
		{
			name: "invalid port too high",
			config: Config{
				Logging: logger.Config{
					Level:   "info",
					LogPath: "/var/log/test.log",
					MaxSize: "10MB",
					MaxAge:  time.Hour * 24,
				},
				Address:     "127.0.0.1",
				Port:        70000,
				MetricsPath: "/metrics",
			},
			wantErr: true,
			errMsg:  "port validation failed",
		},
		{
			name: "invalid address empty",
			config: Config{
				Logging: logger.Config{
					Level:   "info",
					LogPath: "/var/log/test.log",
					MaxSize: "10MB",
					MaxAge:  time.Hour * 24,
				},
				Address:     "",
				Port:        9062,
				MetricsPath: "/metrics",
			},
			wantErr: true,
			errMsg:  "address validation failed",
		},
		{
			name: "invalid metrics path no leading slash",
			config: Config{
				Logging: logger.Config{
					Level:   "info",
					LogPath: "/var/log/test.log",
					MaxSize: "10MB",
					MaxAge:  time.Hour * 24,
				},
				Address:     "127.0.0.1",
				Port:        9062,
				MetricsPath: "metrics",
			},
			wantErr: true,
			errMsg:  "metrics path validation failed",
		},
		{
			name: "invalid log level",
			config: Config{
				Logging: logger.Config{
					Level:   "invalid",
					LogPath: "/var/log/test.log",
					MaxSize: "10MB",
					MaxAge:  time.Hour * 24,
				},
				Address:     "127.0.0.1",
				Port:        9062,
				MetricsPath: "/metrics",
			},
			wantErr: true,
			errMsg:  "logging validation failed",
		},
		{
			name: "invalid log path with special chars",
			config: Config{
				Logging: logger.Config{
					Level:   "info",
					LogPath: "/var/log/test<>.log",
					MaxSize: "10MB",
					MaxAge:  time.Hour * 24,
				},
				Address:     "127.0.0.1",
				Port:        9062,
				MetricsPath: "/metrics",
			},
			wantErr: true,
			errMsg:  "logging validation failed",
		},
		{
			name: "invalid metrics path with special chars",
			config: Config{
				Logging: logger.Config{
					Level:   "info",
					LogPath: "/var/log/test.log",
					MaxSize: "10MB",
					MaxAge:  time.Hour * 24,
				},
				Address:     "127.0.0.1",
				Port:        9062,
				MetricsPath: "/metrics<>",
			},
			wantErr: true,
			errMsg:  "metrics path validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Config.Validate() error message = %v, should contain %v", err, tt.errMsg)
			}
		})
	}
}

func TestConfig_validateAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		{"valid ipv4", "127.0.0.1", false},
		{"valid ipv6", "::1", false},
		{"valid ipv6 expanded", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", false},
		{"valid domain", "example.com", false},
		{"valid domain with subdomain", "sub.example.com", false},
		{"valid interface eth0", "eth0", false},
		{"valid interface wlan0", "wlan0", false},
		{"valid interface ens33", "ens33", false},
		{"valid interface enp0s3", "enp0s3", false},
		{"valid interface docker0", "docker0", false},
		{"valid interface br0", "br0", false},
		{"valid interface lo", "lo", false},
		{"valid interface veth", "veth123456", false},
		{"empty address", "", true},
		{"invalid format", "invalid-address", true},
		{"invalid interface name", "invalid_interface", true},
		{"domain too long", strings.Repeat("a", 254), true},
		{"domain label too long", strings.Repeat("a", 64) + ".com", true},
		{"domain with invalid chars", "ex@mple.com", true},
		{"domain starting with dash", "-example.com", true},
		{"domain ending with dash", "example-.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{Address: tt.address}
			err := config.validateAddress()
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_validatePort(t *testing.T) {
	tests := []struct {
		name    string
		port    int
		wantErr bool
	}{
		{"valid port 8080", 8080, false},
		{"valid port 9062", 9062, false},
		{"valid port 65535", 65535, false},
		{"privileged port 80", 80, false},   // should warn but not error
		{"privileged port 443", 443, false}, // should warn but not error
		{"privileged port 22", 22, false},   // should warn but not error
		{"port too low 0", 0, true},
		{"port too low -1", -1, true},
		{"port too high 65536", 65536, true},
		{"port too high 70000", 70000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{Port: tt.port}
			err := config.validatePort()
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePort() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_validateMetricsPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid path /metrics", "/metrics", false},
		{"valid path /api/metrics", "/api/metrics", false},
		{"valid path /prometheus/metrics", "/prometheus/metrics", false},
		{"valid path /", "/", false},
		{"empty path", "", true},
		{"no leading slash", "metrics", true},
		{"with special chars <", "/metrics<", true},
		{"with special chars >", "/metrics>", true},
		{"with special chars :", "/metrics:", true},
		{"with special chars \"", "/metrics\"", true},
		{"with special chars |", "/metrics|", true},
		{"with special chars ?", "/metrics?", true},
		{"with special chars *", "/metrics*", true},
		{"with multiple special chars", "/metrics<>:\"|?*", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{MetricsPath: tt.path}
			err := config.validateMetricsPath()
			if (err != nil) != tt.wantErr {
				t.Errorf("validateMetricsPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_validateLogging(t *testing.T) {
	tests := []struct {
		name    string
		logging logger.Config
		wantErr bool
	}{
		{
			name: "valid logging config",
			logging: logger.Config{
				Level:   "info",
				LogPath: "/var/log/test.log",
				MaxSize: "10MB",
				MaxAge:  time.Hour * 24,
			},
			wantErr: false,
		},
		{
			name: "valid debug level",
			logging: logger.Config{
				Level:   "debug",
				LogPath: "/var/log/test.log",
				MaxSize: "10MB",
				MaxAge:  time.Hour * 24,
			},
			wantErr: false,
		},
		{
			name: "valid warn level",
			logging: logger.Config{
				Level:   "warn",
				LogPath: "/var/log/test.log",
				MaxSize: "10MB",
				MaxAge:  time.Hour * 24,
			},
			wantErr: false,
		},
		{
			name: "valid error level",
			logging: logger.Config{
				Level:   "error",
				LogPath: "/var/log/test.log",
				MaxSize: "10MB",
				MaxAge:  time.Hour * 24,
			},
			wantErr: false,
		},
		{
			name: "case insensitive level",
			logging: logger.Config{
				Level:   "INFO",
				LogPath: "/var/log/test.log",
				MaxSize: "10MB",
				MaxAge:  time.Hour * 24,
			},
			wantErr: false,
		},
		{
			name: "invalid log level",
			logging: logger.Config{
				Level:   "invalid",
				LogPath: "/var/log/test.log",
				MaxSize: "10MB",
				MaxAge:  time.Hour * 24,
			},
			wantErr: true,
		},
		{
			name: "log path with special chars",
			logging: logger.Config{
				Level:   "info",
				LogPath: "/var/log/test<>.log",
				MaxSize: "10MB",
				MaxAge:  time.Hour * 24,
			},
			wantErr: true,
		},
		{
			name: "max size too long",
			logging: logger.Config{
				Level:   "info",
				LogPath: "/var/log/test.log",
				MaxSize: strings.Repeat("a", 21),
				MaxAge:  time.Hour * 24,
			},
			wantErr: true,
		},
		{
			name: "empty level (should be valid)",
			logging: logger.Config{
				Level:   "",
				LogPath: "/var/log/test.log",
				MaxSize: "10MB",
				MaxAge:  time.Hour * 24,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{Logging: tt.logging}
			err := config.validateLogging()
			if (err != nil) != tt.wantErr {
				t.Errorf("validateLogging() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_GetBindAddress(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		port     int
		expected string
	}{
		{"localhost with port 8080", "127.0.0.1", 8080, "127.0.0.1:8080"},
		{"localhost with port 9062", "127.0.0.1", 9062, "127.0.0.1:9062"},
		{"domain with port 80", "example.com", 80, "example.com:80"},
		{"ipv6 with port 443", "::1", 443, "::1:443"},
		{"interface with port 3000", "eth0", 3000, "eth0:3000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Address: tt.address,
				Port:    tt.port,
			}
			if got := config.GetBindAddress(); got != tt.expected {
				t.Errorf("GetBindAddress() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestConfig_IsLocalhost(t *testing.T) {
	tests := []struct {
		name    string
		address string
		want    bool
	}{
		{"localhost ipv4", "127.0.0.1", true},
		{"localhost name", "localhost", true},
		{"localhost ipv6", "::1", true},
		{"public ip", "8.8.8.8", false},
		{"domain", "example.com", false},
		{"interface", "eth0", false},
		{"private ip", "192.168.1.1", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{Address: tt.address}
			if got := config.IsLocalhost(); got != tt.want {
				t.Errorf("IsLocalhost() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_IsPublic(t *testing.T) {
	tests := []struct {
		name    string
		address string
		want    bool
	}{
		{"localhost ipv4", "127.0.0.1", false},
		{"localhost name", "localhost", false},
		{"localhost ipv6", "::1", false},
		{"public ip", "8.8.8.8", true},
		{"public ip 2", "1.1.1.1", true},
		{"private ip", "192.168.1.1", false},
		{"private ip 2", "10.0.0.1", false},
		{"private ip 3", "172.16.0.1", false},
		{"domain", "example.com", false}, // not an IP, so false
		{"interface", "eth0", false},     // not an IP, so false
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{Address: tt.address}
			if got := config.IsPublic(); got != tt.want {
				t.Errorf("IsPublic() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_isValidDomain(t *testing.T) {
	tests := []struct {
		name   string
		domain string
		want   bool
	}{
		{"valid domain", "example.com", true},
		{"valid domain with subdomain", "sub.example.com", true},
		{"valid domain with multiple subdomains", "a.b.c.example.com", true},
		{"valid domain with numbers", "example123.com", true},
		{"valid domain with hyphens", "my-example.com", true},
		{"valid domain with hyphens in middle", "my-example-site.com", true},
		{"valid domain with numbers and hyphens", "my-example123.com", true},
		{"valid domain with single letter", "a.com", true},
		{"valid domain with single letter subdomain", "a.b.com", true},
		{"valid domain with max length label", strings.Repeat("a", 63) + ".com", true},
		{"valid domain with max total length", "a." + strings.Repeat("a", 250), false}, // 总长度超过253
		{"empty domain", "", false},
		{"domain too long", strings.Repeat("a", 254), false},
		{"domain label too long", strings.Repeat("a", 64) + ".com", false},
		{"domain with invalid chars @", "ex@mple.com", false},
		{"domain with invalid chars !", "ex!mple.com", false},
		{"domain with invalid chars #", "ex#mple.com", false},
		{"domain with invalid chars $", "ex$mple.com", false},
		{"domain with invalid chars %", "ex%mple.com", false},
		{"domain with invalid chars ^", "ex^mple.com", false},
		{"domain with invalid chars &", "ex&mple.com", false},
		{"domain with invalid chars *", "ex*mple.com", false},
		{"domain with invalid chars (", "ex(mple.com", false},
		{"domain with invalid chars )", "ex)mple.com", false},
		{"domain with invalid chars +", "ex+mple.com", false},
		{"domain with invalid chars =", "ex=mple.com", false},
		{"domain with invalid chars [", "ex[mple.com", false},
		{"domain with invalid chars ]", "ex]mple.com", false},
		{"domain with invalid chars {", "ex{mple.com", false},
		{"domain with invalid chars }", "ex}mple.com", false},
		{"domain with invalid chars |", "ex|mple.com", false},
		{"domain with invalid chars \\", "ex\\mple.com", false},
		{"domain with invalid chars /", "ex/mple.com", false},
		{"domain with invalid chars ;", "ex;mple.com", false},
		{"domain with invalid chars :", "ex:mple.com", false},
		{"domain with invalid chars '", "ex'mple.com", false},
		{"domain with invalid chars \"", "ex\"mple.com", false},
		{"domain with invalid chars ,", "ex,mple.com", false},
		{"domain with invalid chars .", "ex.mple.com", true}, // 点号在域名中是合法的
		{"domain with invalid chars <", "ex<mple.com", false},
		{"domain with invalid chars >", "ex>mple.com", false},
		{"domain with invalid chars ?", "ex?mple.com", false},
		{"domain starting with dash", "-example.com", false},
		{"domain ending with dash", "example-.com", false},
		{"domain label starting with dash", "ex.-ample.com", false},
		{"domain label ending with dash", "ex.ample-.com", false},
		{"domain with consecutive dots", "ex..ample.com", false},
		{"domain with dot at end", "example.com.", false},
		{"domain with dot at start", ".example.com", false},
		{"domain with empty label", "ex..ample.com", false},
		{"domain without dot", "example", false},
		{"domain with only dots", "...", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{}
			if got := config.isValidDomain(tt.domain); got != tt.want {
				t.Errorf("isValidDomain(%q) = %v, want %v", tt.domain, got, tt.want)
			}
		})
	}
}

func TestConfig_isValidInterface(t *testing.T) {
	tests := []struct {
		name  string
		iface string
		want  bool
	}{
		{"valid eth0", "eth0", true},
		{"valid eth1", "eth1", true},
		{"valid eth10", "eth10", true},
		{"valid wlan0", "wlan0", true},
		{"valid wlan1", "wlan1", true},
		{"valid ens33", "ens33", true},
		{"valid ens34", "ens34", true},
		{"valid enp0s3", "enp0s3", true},
		{"valid enp0s8", "enp0s8", true},
		{"valid docker0", "docker0", true},
		{"valid docker1", "docker1", true},
		{"valid br0", "br0", true},
		{"valid br1", "br1", true},
		{"valid lo", "lo", true},
		{"valid veth123456", "veth123456", true},
		{"valid vethabcdef", "vethabcdef", true},
		{"valid vethABCDEF", "vethABCDEF", false}, // 大写字母在接口名称中通常不被支持
		{"empty interface", "", false},
		{"interface too long", strings.Repeat("a", 16), false},
		{"interface with underscore", "eth_0", false},
		{"interface with invalid chars", "eth@0", false},
		{"interface with spaces", "eth 0", false},
		{"interface with dots", "eth.0", false},
		{"interface with special chars", "eth#0", false},
		{"interface starting with number", "0eth", false},
		{"interface with mixed case", "Eth0", false},
		{"interface with consecutive numbers", "eth00", true}, // valid pattern
		{"interface with many numbers", "eth999", true},       // valid pattern
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{}
			if got := config.isValidInterface(tt.iface); got != tt.want {
				t.Errorf("isValidInterface(%q) = %v, want %v", tt.iface, got, tt.want)
			}
		})
	}
}

// 基准测试
func BenchmarkConfig_Validate(b *testing.B) {
	config := Config{
		Logging: logger.Config{
			Level:   "info",
			LogPath: "/var/log/test.log",
			MaxSize: "10MB",
			MaxAge:  time.Hour * 24,
		},
		Address:     "127.0.0.1",
		Port:        9062,
		MetricsPath: "/metrics",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.Validate()
	}
}

func BenchmarkConfig_validateAddress(b *testing.B) {
	config := &Config{Address: "127.0.0.1"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.validateAddress()
	}
}

func BenchmarkConfig_validatePort(b *testing.B) {
	config := &Config{Port: 8080}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.validatePort()
	}
}
