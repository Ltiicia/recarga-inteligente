package logger

import (
	"io"
	"log"
)

type Logger struct {
	output io.Writer
}

func NewLogger(output io.Writer) *Logger {
	return &Logger{output: output}
}

func (logger *Logger) Info(msg string) {
	log.SetOutput(logger.output)
	log.Println(msg)
}

func (logger *Logger) Erro(erro string) {
	log.SetOutput(logger.output)
	log.Println("ERRO: " + erro)
}
