package test

import (
	"datapi/pkg/core"
)

type MemoryLogSaver struct {
	logs []string
}

func NewInMemoryLogSaver() *MemoryLogSaver {
	return &MemoryLogSaver{logs: []string{}}
}

func (mls *MemoryLogSaver) SaveLog(message core.AccessLog) error {
	mls.logs = append(mls.logs, message.String())
	return core.PrintLogToStdout(message)
	//return nil
}

func (mls *MemoryLogSaver) Last() string {
	return mls.logs[len(mls.logs)-1]
}
