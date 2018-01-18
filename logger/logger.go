package logger

import "fmt"

var Logger = &LevelLogger{}

type LevelLogger struct {
	Level int
}

func (l *LevelLogger) Log(lvl int, a ...interface{}) {
	if lvl <= l.Level {
		fmt.Print(a...)
	}
}

func (l *LevelLogger) Logf(lvl int, format string, a ...interface{}) {
	if lvl <= l.Level {
		fmt.Printf(format, a...)
	}
}

func (l *LevelLogger) Logln(lvl int, a ...interface{}) {
	if lvl <= l.Level {
		fmt.Println(a...)
	}
}
