// 简单的增强log库
package lg

import (
	"fmt"
	"strings"
)

type LogLevel int

const (
	DEBUG = LogLevel(1)
	INFO  = LogLevel(2)
	WARN  = LogLevel(3)
	ERROR = LogLevel(4)
	FATAL = LogLevel(5)
)

type AppLogFunc func(lvl LogLevel, f string, args ...interface{})

type Logger interface {
	Output(maxdepth int, s string) error
}

func (l LogLevel) String() string {
	// 判等的时候会做类型转换
	switch l {
	case 1:
		return "DEBUG"
	case 2:
		return "INFO"
	case 3:
		return "WARN"
	case 4:
		return "ERROR"
	case 5:
		return "FATAL"
	default:
		panic("不合法的日志等级")
	}
}

func ParseLogLevel(levelstr string, verbose bool) (LogLevel, error) {
	lvl := INFO

	// 解析日志等级
	switch strings.ToLower(levelstr) {
	case "debug":
		lvl = DEBUG
	case "info":
		lvl = INFO
	case "warn":
		lvl = WARN
	case "error":
		lvl = ERROR
	case "fatal":
		lvl = FATAL
	default:
		return lvl, fmt.Errorf("不合法的日志等级 '%s'", levelstr)
	}
	if verbose {
		lvl = DEBUG
	}
	return lvl, nil
}

func Logf(logger Logger, cfgLevel LogLevel, msgLevel LogLevel, f string, args ...interface{}) {
	// 判断下日志等级，如果当前日志等级msgLevel小于配置日志等级cfgLevel，则不打印到日志上
	if cfgLevel > msgLevel {
		return
	}
	logger.Output(3, fmt.Sprintf(msgLevel.String()+": "+f, args...))
}
