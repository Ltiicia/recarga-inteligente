package tcpIP

import (
	"fmt"
	"net"
	"recarga-inteligente/internal/handler"
	"recarga-inteligente/internal/logger"
	"recarga-inteligente/internal/store"
)

func StartServerTCP(porta string, connectionStore *store.ConnectionStore, logger *logger.Logger) error {
	listener, erro := net.Listen("tcp", porta)

	if erro != nil {
		logger.Erro(fmt.Sprintf("Erro ao iniciar servidor: %v", erro))
		return nil
	}
	defer listener.Close()

	logger.Info(fmt.Sprintf("Servidor inicializado escutando na porta %s...", porta))

	//Aceita conexoes e trata cada uma em uma goroutine
	for {
		novaConexao, erro := listener.Accept()
		if erro != nil {
			logger.Erro(fmt.Sprintf("erro ao aceitar conexao: %v", erro))
			continue
		}

		go handler.HandleConnection(novaConexao, connectionStore, logger)
	}
}
