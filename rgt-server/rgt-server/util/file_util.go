package util

import (
	"bufio"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"rgt-server/log"
	"time"
)

func IsFile(pathName string) bool {
	fileInfo, err := os.Stat(pathName)
	if err != nil {
		return false
	}
	if fileInfo.Mode().IsDir() {
		return false
	}
	return true
}

func FileExists(filePathName string) bool {
	_, err := os.Stat(filePathName)
	return err == nil
}

func FileSize(filePathName string) int64 {
	fi, err := os.Stat(filePathName)
	if err != nil {
		return 0
	}
	return fi.Size()
}

func RelativePathToAbsolute(pathName string) string {
	if !filepath.IsAbs(pathName) {
		curDir, err := os.Getwd()
		if err == nil {
			return filepath.Join(curDir, pathName)
		} else {
			log.Error(err)
		}
	}
	return pathName
}

func RemoveFiles(filePath string, fileMask string, daysRetation int) {
	dir := os.DirFS(filePath)
	files, err := fs.Glob(dir, fileMask)
	if err != nil {
		log.Errorf("Error cleaning app log files: %v", err)
		return
	}
	if len(files) == 0 {
		return
	}
	log.Infof("%d files found to remove. path='%s' mask='%s'", len(files), filePath, fileMask)
	filesRemoved := 0
	for _, file := range files {
		if removeFile(path.Join(filePath, file), daysRetation) {
			filesRemoved++
		}
	}
	log.Infof("%d files removed. path='%s' mask='%s'", filesRemoved, filePath, fileMask)
}

func removeFile(filePathName string, daysRetation int) bool {
	fileStat, fileErr := os.Stat(filePathName)
	if fileErr == nil {
		if time.Since(fileStat.ModTime()).Hours()/24 > float64(daysRetation) {
			errRemove := os.Remove(filePathName)
			if errRemove == nil {
				log.Debugf("file_util.RemoveFiles(): File '%s' removed.", filePathName)
				return true
			} else {
				log.Debugf("file_util.RemoveFiles(): Error removing file '%s': %v", filePathName, errRemove)
			}
		}
	} else {
		log.Debugf("file_util.RemoveFiles(): Error getting file info for file '%s': %v", filePathName, fileErr)
	}
	return false
}

func TruncateFile(filePathName string, maxFileSize int64, newFileSize int64) {
	var in *os.File
	var out *os.File
	var err error

	tmpFileName := filePathName + ".tmp"

	if newFileSize > maxFileSize || FileSize(filePathName) < maxFileSize {
		return
	}
	defer func() {
		if in != nil {
			in.Close()
		}
		if out != nil {
			out.Close()
		}
		if FileExists(tmpFileName) {
			os.Remove(filePathName)
			os.Rename(tmpFileName, filePathName)
		}
	}()
	in, err = os.OpenFile(filePathName, os.O_RDONLY, 0666)
	if err != nil {
		log.Error("Error opening file to truncate. File: ", filePathName, ". Error: ", err)
		return
	}
	_, err = in.Seek(-1*newFileSize, io.SeekEnd)
	if err != nil {
		log.Error("Error reading initial position to truncate. File: ", filePathName, ". Error: ", err)
		return
	}
	out, err = os.Create(tmpFileName)
	if err != nil {
		log.Error("Error creating temp file to truncate. File: ", filePathName, ". Error: ", err)
		return
	}
	scanner := bufio.NewScanner(in)
	scanner.Scan()
	for scanner.Scan() {
		_, err = out.Write(scanner.Bytes())
		if err != nil {
			log.Error("Error writing to temp file to truncate. File: ", filePathName, ". Error: ", err)
			return
		}
		_, err = out.WriteString("\n")
		if err != nil {
			log.Error("Error writing to temp file to truncate. File: ", filePathName, ". Error: ", err)
			return
		}
	}
}
