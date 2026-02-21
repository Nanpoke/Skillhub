package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CopyFile 复制文件（使用流式处理支持大文件）
func CopyFile(src, dst string) error {
	// 打开源文件
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// 获取源文件权限
	info, err := srcFile.Stat()
	if err != nil {
		return err
	}

	// 创建目标文件
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// 使用 io.Copy 流式复制（内存友好）
	_, err = io.Copy(dstFile, srcFile)
	return err
}

// CopyDir 递归复制目录
func CopyDir(src, dst string, skipGit bool) error {
	// 获取源目录信息
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	// 创建目标目录
	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		return err
	}

	// 读取源目录
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		// 跳过 .git 目录（可选）
		if skipGit && entry.Name() == ".git" {
			continue
		}

		if entry.IsDir() {
			if err := CopyDir(srcPath, dstPath, skipGit); err != nil {
				return err
			}
		} else {
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// UnzipFile 安全解压 ZIP 文件
// 使用 SanitizeZipPath 防止路径遍历攻击
func UnzipFile(zipPath, destDir string) error {
	// 打开 ZIP 文件
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	// 创建目标目录
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	// 解压文件
	for _, f := range r.File {
		// 使用安全路径验证
		fpath, err := SanitizeZipPath(destDir, f.Name)
		if err != nil {
			return fmt.Errorf("unsafe path in zip: %w", err)
		}

		// 创建目录
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, f.Mode()); err != nil {
				return err
			}
			continue
		}

		// 创建父目录
		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return err
		}

		// 解压文件
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		// 使用 io.Copy 流式解压
		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// RemoveDir 安全删除目录
func RemoveDir(path string) error {
	return os.RemoveAll(path)
}

// EnsureDir 确保目录存在
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}
