package main

import (
	"os"

	"rgt-server/config"
	"rgt-server/log"
	"rgt-server/util"
)

func createLoggerFile(logFilePathName string, level log.LogLevel) log.Logger {
	util.TruncateFile(logFilePathName, 15*1024*1024, 10*1024*1024)
	logPathname := util.RelativePathToAbsolute(logFilePathName)
	file, err := os.OpenFile(logPathname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Error("Error creating log file:", logPathname, "Error:", err)
		return nil
	}
	logger := log.NewLogger(level, file)
	logger.SetFormatterf(log.TimestampLogPrintf)
	logger.SetFormatter(log.TimestampLogPrintln)
	return logger
}

func configLoggerFile(logger log.Logger, conf *config.ServerConfig) {
	logger.SetLevel(conf.ServerLogLevel().Get())
	setLoggerFilePathname(logger, conf.ServerLogPathName().Get())
	conf.ServerLogLevel().SetHook(func(val log.LogLevel) {
		logger.SetLevel(val)
	})
	conf.ServerLogPathName().SetHook(func(filePathname string) {
		setLoggerFilePathname(logger, filePathname)
	})
}

func setLoggerFilePathname(logger log.Logger, newLogPathname string) {
	var currentLogInfo os.FileInfo
	newLogInfo, errNew := os.Stat(newLogPathname)
	if errNew != nil {
		log.Errorf("invalid path name for log: %s", newLogPathname)
		return
	}
	if currentFile, ok := logger.GetOutput().(*os.File); ok {
		currentLogInfo, _ = currentFile.Stat()
	}
	if currentLogInfo == nil || !os.SameFile(currentLogInfo, newLogInfo) {
		newFile, err := os.OpenFile(newLogPathname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Error("Error creating log file:", newLogPathname, "Error:", err)
			return
		}
		closeLoggerFile(logger)
		logger.SetOutput(newFile)
	}
}

func closeLoggerFile(logger log.Logger) {
	if file, ok := logger.GetOutput().(*os.File); ok {
		file.Close()
	}
}
