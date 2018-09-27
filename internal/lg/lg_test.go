package lg

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogLevelString(t *testing.T) {
	output := fmt.Sprintf("%s", DEBUG)
	assert.Equal(t, "DEBUG", output)
}

func TestParseLogLevel(t *testing.T) {
	res, _ := ParseLogLevel("DEBUG", false)
	assert.Equal(t, DEBUG, res)
	res, _ = ParseLogLevel("INFO", false)
	assert.Equal(t, INFO, res)
	// verbose 设置为true时，都返回DEBUG
	res, _ = ParseLogLevel("INFO", true)
	assert.Equal(t, DEBUG, res)
}

type mockLogger struct {
	// 用来记录是不是调用了log方法
	Count int
}

func (l *mockLogger) Output(maxdepth int, s string) error {
	l.Count++
	return nil
}

func TestLogfSuccuse(t *testing.T) {
	logger := new(mockLogger)
	// 如果当前日志等级大于配置日志等级，则可以输出
	for i := 0; i < 5; i++ {
		Logf(logger, INFO, ERROR, "test")
	}
	assert.Equal(t, 5, logger.Count)
}

func TestLogfFail(t *testing.T) {
	logger := new(mockLogger)
	// 如果当前日志等级小于配置日志等级，则不可以输出
	for i := 0; i < 5; i++ {
		Logf(logger, INFO, DEBUG, "test")
	}
	assert.Equal(t, 0, logger.Count)
}
