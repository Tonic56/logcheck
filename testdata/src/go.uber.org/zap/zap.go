// Package zap is a minimal stub of go.uber.org/zap used only in analysistest
// testdata. It exposes the Logger and SugaredLogger types with the logging
// methods that logcheck must detect.
package zap

// Logger is the non-sugar zap logger stub.
type Logger struct{}

func (l *Logger) Info(msg string, fields ...interface{})   {}
func (l *Logger) Warn(msg string, fields ...interface{})   {}
func (l *Logger) Error(msg string, fields ...interface{})  {}
func (l *Logger) Debug(msg string, fields ...interface{})  {}
func (l *Logger) Fatal(msg string, fields ...interface{})  {}
func (l *Logger) Panic(msg string, fields ...interface{})  {}
func (l *Logger) DPanic(msg string, fields ...interface{}) {}

// SugaredLogger is the sugared zap logger stub.
type SugaredLogger struct{}

func (s *SugaredLogger) Infof(template string, args ...interface{})  {}
func (s *SugaredLogger) Warnf(template string, args ...interface{})  {}
func (s *SugaredLogger) Errorf(template string, args ...interface{}) {}
func (s *SugaredLogger) Debugf(template string, args ...interface{}) {}
func (s *SugaredLogger) Infow(msg string, keysAndValues ...interface{})  {}
func (s *SugaredLogger) Warnw(msg string, keysAndValues ...interface{})  {}
func (s *SugaredLogger) Errorw(msg string, keysAndValues ...interface{}) {}
func (s *SugaredLogger) Debugw(msg string, keysAndValues ...interface{}) {}
