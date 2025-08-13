// SPDX-FileCopyrightText: 2025 UnionTech Software Technology Co., Ltd.
// SPDX-License-Identifier: MIT

package utils

import (
	"fmt"
	"net/url"
)

func ValidateURI(uri string) error {
	parsedURL, err := url.ParseRequestURI(uri)
	if err != nil {
		return fmt.Errorf("invalid URI format: %w", err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("unsupported scheme: %s", parsedURL.Scheme)
	}
	return nil
}
