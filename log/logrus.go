package log

import (
	"context"
	"fmt"
	nested "github.com/antonfisher/nested-logrus-formatter"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm/logger"
	"path/filepath"
	"runtime"
	"strings"
	"tgwp/utils"
	"time"
)

func InitLogger() {
	hook := NewLfsHook(utils.GetRootPath("log/logfiles/api.log"), nil, 10)
	logrus.AddHook(hook)

	logrus.SetFormatter(formatter(true))
	logrus.SetReportCaller(true)
	logrus.SetLevel(logrus.DebugLevel)

	logrus.Debugln("[INIT] init log success")

}

func formatter(isConsole bool) *nested.Formatter {
	fmtter := &nested.Formatter{
		HideKeys:        false,
		TimestampFormat: "2006-01-02 15:04:05",
		CallerFirst:     true,
		ShowFullLevel:   true,
		CustomCallerFormatter: func(frame *runtime.Frame) string {
			funcInfo := runtime.FuncForPC(frame.PC)
			if funcInfo == nil {
				return "error during runtime.FuncForPC"
			}
			fullPath, line := funcInfo.FileLine(frame.PC)
			fncSlice := strings.Split(funcInfo.Name(), ".")
			fncName := fncSlice[len(fncSlice)-1]
			return fmt.Sprintf("[%15v] [%15v] [%3v]", filepath.Base(fullPath), fncName, line)
		},
	}
	if isConsole {
		fmtter.NoColors = false
	} else {
		fmtter.NoColors = true
	}
	return fmtter
}

func NewLfsHook(logName string, logLevel *string, maxRemainCnt uint) logrus.Hook {
	writer, err := rotatelogs.New(
		logName+"_bak %Y-%m-%d %H",
		// WithLinkName为最新的日志建立软连接，以方便随着找到当前日志文件
		rotatelogs.WithLinkName(logName),

		// WithRotationTime设置日志分割的时间，这里设置为24小时分割一次
		rotatelogs.WithRotationTime(time.Hour*24),

		// WithMaxAge和WithRotationCount二者只能设置一个，
		// WithMaxAge设置文件清理前的最长保存时间，go
		// WithRotationCount设置文件清理前最多保存的个数。
		//rotatelogs.WithMaxAge(time.Hour*24),
		rotatelogs.WithRotationCount(maxRemainCnt),
	)

	if err != nil {
		logrus.Errorf("config local file system for logger error: %v", err)
	}

	lfsHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.DebugLevel: writer,
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: writer,
		logrus.FatalLevel: writer,
		logrus.PanicLevel: writer,
	}, formatter(true))

	return lfsHook
}

type MyLoggerStruct struct {
	Logger *logrus.Logger
}

/*
实现接口
*/
func (m *MyLoggerStruct) LogMode(loglevel logger.LogLevel) logger.Interface {
	m.Logger.Level = logrus.Level(logrus.Level(loglevel))
	return m
}
func (m *MyLoggerStruct) Info(ctx context.Context, msg string, data ...interface{}) {
	if m.Logger.Level >= logrus.InfoLevel {
		m.Logger.WithContext(ctx).Infof(msg, data...)
	}
}
func (m *MyLoggerStruct) Warn(ctx context.Context, msg string, data ...interface{}) {
	if m.Logger.Level >= logrus.WarnLevel {
		m.Logger.WithContext(ctx).Warnf(msg, data...)
	}
}
func (m *MyLoggerStruct) Error(ctx context.Context, msg string, data ...interface{}) {
	if m.Logger.Level >= logrus.ErrorLevel {
		m.Logger.WithContext(ctx).Errorf(msg, data...)
	}
}
func (m *MyLoggerStruct) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sql, rowsAffected := fc()
	elapsed := time.Since(begin)
	if err != nil {
		m.Logger.WithContext(ctx).WithError(err).WithFields(logrus.Fields{
			"rowsAffected": rowsAffected,
			"sql":          sql,
			"elapsed":      elapsed,
		}).Error("trace")
	} else {
		return
	}
}
