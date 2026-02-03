//go:build windows
// +build windows

package storage

import (
	"context"
	"fmt"
	"path/filepath"

	"golang.org/x/sys/windows"
)

// GetDiskSpace 获取磁盘空间信息 (Windows)
func (s *LocalStorage) GetDiskSpace(ctx context.Context) (total, used, free uint64, err error) {
	absPath, err := filepath.Abs(s.basePath)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get absolute path: %w", err)
	}

	pathPtr, err := windows.UTF16PtrFromString(absPath)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to convert path: %w", err)
	}

	var freeBytesAvailable, totalBytes, totalFreeBytes uint64

	err = windows.GetDiskFreeSpaceEx(
		pathPtr,
		&freeBytesAvailable,
		&totalBytes,
		&totalFreeBytes,
	)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get disk space: %w", err)
	}

	total = totalBytes
	free = freeBytesAvailable
	used = total - free

	return total, used, free, nil
}
