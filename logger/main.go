package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"time"
)

type customLogger struct {
	info io.Writer
	warn io.Writer
	err  io.Writer
}

type LoggerWriter interface {
	Info(string)
	Warning(string)
	Error(string)
	Message(message string) string
}

func NewLogger() LoggerWriter {
	infoLogs, err := os.OpenFile("./logs/info.log", os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		log.Fatal(err)
	}
	errorLogs, err := os.OpenFile("./logs/error.log", os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		log.Fatal(err)
	}
	warningLogs, err := os.OpenFile("./logs/warning.log", os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		log.Fatal(err)
	}

	return &customLogger{
		info: infoLogs,
		warn: warningLogs,
		err:  errorLogs,
	}
}

var (
	yellowBg = string([]byte{27, 91, 57, 48, 59, 52, 51, 109})
	redBg    = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	blueBg   = string([]byte{27, 91, 56, 55, 59, 52, 52, 109})
	reset    = string([]byte{27, 91, 48, 109})
)

func itoa(i int, wid int) string {
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}

	b[bp] = byte('0' + i)
	return string(b[bp:])
}

func formatFileRow(t time.Time, file string, line int) string {
	var fileName string
	for i := len(file) - 1; i >= 0; i-- {
		if file[i] == '/' {
			fileName = file[i+1:]
			break
		}
	}

	year, month, day := t.Date()
	hour, min, sec := t.Clock()
	timestamp := fmt.Sprintf("%d/%s/%s %s:%s:%s", year, itoa(int(month), 2), itoa(day, 2), itoa(hour, 2), itoa(min, 2), itoa(sec, 2))
	return fmt.Sprintf("%s %s:%s:", timestamp, fileName, itoa(line, -1))
}

func (l *customLogger) Message(message string) string {
	now := time.Now()

	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "???"
		line = 0
	}

	return fmt.Sprintf("%s %s\n", formatFileRow(now, file, line), message)
}

func (l *customLogger) Info(message string) {
	os.Stdout.WriteString(blueBg + " INFO: " + reset + " " + message)
	l.info.Write([]byte("INFO: " + message))
}

func (l *customLogger) Warning(message string) {
	os.Stdout.WriteString(yellowBg + " WARNING: " + reset + " " + message)
	l.warn.Write([]byte("WARNING: " + message))
}

func (l *customLogger) Error(message string) {
	os.Stderr.WriteString(redBg + " ERROR: " + reset + " " + message)
	l.err.Write([]byte("ERROR: " + message))
}
