package main

import (
	"fmt"
	"net"
	"os"
	"recarga-inteligente/cmd/veiculo/manageVeiculo"
	"recarga-inteligente/internal/logger"
	"recarga-inteligente/internal/tcpIP"
	"time"
)

func limparBuffer(conexao net.Conn, logger *logger.Logger) {
	conexao.SetReadDeadline(time.Now().Add(1 * time.Second))
	buf := make([]byte, 1024)
	n, _ := conexao.Read(buf)
	// limpa o buffer se estiver sujo
	if n > 0 {
		logger.Info("Limpando buffer de resposta.")
	}
	conexao.SetReadDeadline(time.Time{}) // reseta o timeout
}

func main() {
	//inicializa o veiculo e conecta ao servidor
	logger := logger.NewLogger(os.Stdout)
	conexao, erro := tcpIP.ConnectToServerTCP("servidor:5000")
	if erro != nil {
		logger.Erro(fmt.Sprintf("Erro em ConnectToServerTCP - veiculo: %v", erro))
		return
	}
	limparBuffer(conexao, logger)
	logger.Info("Veiculo conectado")
	defer conexao.Close()

	//exibe menu de opcoes
	manageVeiculo.MenuVeiculo(logger, conexao)
}
