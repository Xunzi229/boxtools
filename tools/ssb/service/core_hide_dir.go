package service

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

// 是否存在
func existDir(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			log.Println(err)
			return false
		}
	}
	return true
}

func HideDir(dir string) {
	switch runtime.GOOS {
	case "windows":
		if err := hideWindowDir(dir); err != nil {
			log.Println(err)
		}
	default:
		if err := hideUnixDir(dir); err != nil {
			log.Println(err)
		}
	}
}

func HideFile(doc string) {
	switch runtime.GOOS {
	case "windows":
		if err := hideWindowDir(doc); err != nil {
			log.Println(err)
		}
	default:
		if err := hideUnixDir(doc); err != nil {
			log.Println(err)
		}
	}
}

// windows
func hideWindowDir(pathName string) error {
	fileName, err := syscall.UTF16PtrFromString(pathName)
	if err != nil {
		return err
	}
	return syscall.SetFileAttributes(fileName, syscall.FILE_ATTRIBUTE_HIDDEN)
}

// unix
func hideUnixDir(pathName string) error {
	if strings.HasPrefix(filepath.Base(pathName), ".") {
		return nil
	}
	return os.Rename(pathName, "."+pathName)
}