package main

import (
	"fmt"
	"os"
	"recarga-inteligente/internal/dataJson"
	"recarga-inteligente/internal/logger"
	"recarga-inteligente/internal/tcpIP"
)

func main() {
	logger := logger.NewLogger(os.Stdout)

	conexao, erro := tcpIP.ConnectToServerTCP("servidor:5000")
	if erro != nil {
		os.Exit(1)
	}
	logger.Info("Ponto de Recarga conectado")
	defer conexao.Close()

	var respostaServidor dataJson.Mensagem
	respostaServidor, erro = tcpIP.SendIdentification(conexao, "ponto-de-recarga")

	if erro != nil {
		logger.Erro(fmt.Sprintf("Erro ao obter resposta do servidor - %v", erro))
		return
	}

	if respostaServidor.Tipo == "get-disponibilidade" {
		msg := dataJson.Mensagem{
			Tipo:     "disponibilidade",
			Conteudo: "Situacao atual: sem fila",
			Origem:   "ponto-de-recarga",
		}
		erro := dataJson.SendMessage(conexao, msg)
		if erro != nil {
			logger.Erro(fmt.Sprintf("Erro ao enviar disponibilidade - %v", erro))
		}
		//logger.Info(msg.Conteudo)
	}

	for {
		mensagemRecebida, erro := dataJson.ReceiveMessage(conexao)
		if erro != nil {
			logger.Erro(fmt.Sprintf("Erro ao ler mensagem do servidor - %v", erro))
			return
		}
		logger.Info(fmt.Sprintf("Mensagem recebida do servidor: %s", mensagemRecebida.Conteudo))
	}
}
