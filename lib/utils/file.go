package utils

import (
	"path/filepath"
	"os"
	"fmt"
	"io"
	"io/ioutil"
)

// checkDirPath 会检查目录路径。
func CheckDirPath(dirPath string) (absDirPath string, err error) {
	if dirPath == "" {
		err = fmt.Errorf("invalid dir path: %s", dirPath)
		return
	}
	if filepath.IsAbs(dirPath) {
		absDirPath = dirPath
	} else {
		absDirPath, err = filepath.Abs(dirPath)
		if err != nil {
			return
		}
	}
	var dir *os.File
	dir, err = os.Open(absDirPath)
	if err != nil && !os.IsNotExist(err) {
		return
	}
	if dir == nil {
		err = os.MkdirAll(absDirPath, 0700)
		if err != nil && !os.IsExist(err) {
			return
		}
	} else {
		var fileInfo os.FileInfo
		fileInfo, err = dir.Stat()
		if err != nil {
			return
		}
		if !fileInfo.IsDir() {
			err = fmt.Errorf("not directory: %s", absDirPath)
			return
		}
	}
	return
}


func SaveFile(dirPath, fileName string, b []byte) (err error) {
	// 检查和准备数据。
	var absDirPath string
	if absDirPath, err = CheckDirPath(dirPath); err != nil {
		return err
	}
	// 创建图片文件。
	filePath := filepath.Join(absDirPath, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("couldn't create file: %s (path: %s)", err, filePath)
	}
	defer file.Close()
	_, err = file.Write(b)
	return err
}

func AppendFile(dirPath, fileName string, b []byte) (err error) {
	// 检查和准备数据。
	var absDirPath string
	if absDirPath, err = CheckDirPath(dirPath); err != nil {
		return err
	}
	// 创建图片文件。
	filePath := filepath.Join(absDirPath, fileName)
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0700)
	if err != nil {
		return fmt.Errorf("couldn't open file: %s (path: %s)", err, filePath)
	}
	defer file.Close()
	_, err = file.Write(b)
	return err
}

func SaveFileByReader(dirPath, fileName string, reader io.Reader) (err error) {
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return
	}
	return SaveFile(dirPath, fileName, b)
}

func AppendFileByReader(dirPath, fileName string, reader io.Reader) (err error) {
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return
	}
	return AppendFile(dirPath, fileName, b)
}
