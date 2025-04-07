package main

import (
	"fmt"
	"math/rand"
	"os"
	"recarga-inteligente/cmd/veiculo/manageVeiculo"
	"recarga-inteligente/internal/logger"
	"recarga-inteligente/internal/tcpIP"
	"time"
)

func main() {
	// Inicializar o gerador de números aleatórios
	rand.Seed(time.Now().UnixNano())

	// Inicializar o gerador de números aleatórios
	rand.Seed(time.Now().UnixNano())

	//inicializa o veiculo e conecta ao servidor
	logger := logger.NewLogger(os.Stdout)
	conexao, erro := tcpIP.ConnectToServerTCP("servidor:5000")
	if erro != nil {
		logger.Erro(fmt.Sprintf("Erro em ConnectToServerTCP - veiculo: %v", erro))
		return
	}
	logger.Info("Veiculo conectado")
	defer conexao.Close()

	//exibe menu de opcoes
	manageVeiculo.MenuVeiculo(logger, conexao)
}
