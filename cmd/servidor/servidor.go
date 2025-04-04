package main

import (
	"os"
	"recarga-inteligente/internal/logger"
	"recarga-inteligente/internal/store"
	"recarga-inteligente/internal/tcpIP"
)

func main() {
	logger := logger.NewLogger(os.Stdout)
	connectionStore := store.NewConnectionStore()

	//Inicia o servidor TCP na porta 5000
	erro := tcpIP.StartServerTCP(":5000", connectionStore, logger)
	if erro != nil {
		logger.Erro("Erro ao iniciar servidor TCP em StartServerTCP")
		return
	}
}
