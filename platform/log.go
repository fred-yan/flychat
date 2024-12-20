package platform

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"time"
)

type Hook struct {
	writer   *os.File
	logPath  string
	fileName string
	fileDate string
}

func (MyHook *Hook) Levels() []logrus.Level {
	return logrus.AllLevels
}
func (MyHook *Hook) Fire(entry *logrus.Entry) error {
	timer := time.Now().Format("2006-01-02")
	line, _ := entry.String()
	//需要切换日志文件
	if MyHook.fileDate != timer {
		MyHook.fileDate = timer
		MyHook.writer.Close()
		filepath := fmt.Sprintf("%s/%s", MyHook.logPath, MyHook.fileDate)
		err := os.MkdirAll(filepath, os.ModePerm)
		if err != nil {
			logrus.Error(err)
			return err
		}
		filename := fmt.Sprintf("%s/%s.log", filepath, MyHook.fileName)
		MyHook.writer, _ = os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	}
	MyHook.writer.Write([]byte(line))
	return nil
}

type LogFormatter struct {
}

func (m *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	timestamp := entry.Time.Format("2006-01-02 15:04:05.000")
	var newLog string
	newLog = fmt.Sprintf("[%s] [%s] %s\n", timestamp, entry.Level, entry.Message)

	b.WriteString(newLog)
	return b.Bytes(), nil
}

func InitFile(logPath string, fileName string) {
	logrus.SetFormatter(&LogFormatter{})
	timer := time.Now().Format("2006-01-02")
	filepath := fmt.Sprintf("%s", logPath)
	err := os.MkdirAll(filepath, os.ModePerm)
	if err != nil {
		logrus.Error(err)
		return
	}
	filename := fmt.Sprintf("%s/%s-%s.log", filepath, timer, fileName)
	writer, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		logrus.Error(err)
		return
	}
	logrus.AddHook(&Hook{
		writer:   writer,
		logPath:  logPath,
		fileName: fileName,
		fileDate: timer,
	})
}

func InitAppLogger(logPath string, fileName string) *logrus.Logger {
	logger := logrus.New()

	timer := time.Now().Format("2006-01-02")
	filepath := fmt.Sprintf("%s", logPath)
	err := os.MkdirAll(filepath, os.ModePerm)
	if err != nil {
		logrus.Error(err)
		return nil
	}
	filename := fmt.Sprintf("%s/%s-%s.log", filepath, timer, fileName)
	logFile, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		logrus.Error(err)
		return nil
	}
	logger.SetFormatter(&LogFormatter{})
	logger.SetOutput(io.MultiWriter(logFile, os.Stderr))
	return logger
}

var Logger = InitAppLogger("./log", "flychat")
