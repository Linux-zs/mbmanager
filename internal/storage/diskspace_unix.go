//go:build !windows
// +build !windows

package storage

import (
	"context"
	"fmt"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// GetDiskSpace 获取磁盘空间信息 (Unix/Linux)
func (s *LocalStorage) GetDiskSpace(ctx context.Context) (total, used, free uint64, err error) {
	absPath, err := filepath.Abs(s.basePath)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get absolute path: %w", err)
	}

	var stat unix.Statfs_t
	err = unix.Statfs(absPath, &stat)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get disk space: %w", err)
	}

	total = stat.Blocks * uint64(stat.Bsize)
	free = stat.Bavail * uint64(stat.Bsize)
	used = total - free

	return total, used, free, nil
}
