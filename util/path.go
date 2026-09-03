package util

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hosgf/element/logger"
)

// MatchPath 判断 urlPath 是否命中任一模式；支持末尾 * 前缀匹配。
func MatchPath(urlPath string, patterns []string) bool {
	if len(patterns) == 0 {
		return false
	}
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if urlPath == p {
			return true
		}
		if strings.HasSuffix(p, "*") {
			prefix := strings.TrimSuffix(p, "*")
			if strings.HasPrefix(urlPath, prefix) {
				return true
			}
		}
	}
	return false
}

// JoinPath 拼接路径，将多个片段用 / 连接。
func JoinPath(base string, parts ...string) string {
	for _, p := range parts {
		if p != "" {
			base += "/" + p
		}
	}
	return base
}

func AppDir() string {
	homePath := GetHomePath()
	if len(homePath) < 1 {
		return homePath
	}
	dir := exeDir()
	tmpDir, _ := filepath.EvalSymlinks(os.TempDir())
	if strings.Contains(dir, tmpDir) {
		return callerDir()
	}
	return dir
}

func exeDir() string {
	exePath, err := os.Executable()
	if err != nil {
		logger.Log().Error(context.Background(), err)
	}
	dir, _ := filepath.EvalSymlinks(filepath.Dir(exePath))
	return dir
}

func callerDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	return path.Dir(filename)
}
