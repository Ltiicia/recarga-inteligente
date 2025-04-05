package main

import (
	"fmt"
	"net"
	"os"
	"recarga-inteligente/internal/dataJson"
	"recarga-inteligente/internal/logger"
	"recarga-inteligente/internal/tcpIP"
)

func enviarDisponibilidade(logger *logger.Logger, conexao net.Conn) {
	msg := dataJson.Mensagem{
		Tipo:     "disponibilidade",
		Conteudo: "Situacao atual: sem fila",
		Origem:   "ponto-de-recarga",
	}
	erro := dataJson.SendMessage(conexao, msg)
	if erro != nil {
		logger.Erro(fmt.Sprintf("Erro ao enviar disponibilidade - %v", erro))
	}
}

func IdentificacaoInicial(logger *logger.Logger, conexao net.Conn) {
	erro := tcpIP.SendIdentification(conexao, "ponto-de-recarga")
	if erro != nil {
		logger.Erro(fmt.Sprintf("Erro ao obter resposta do servidor - %v", erro))
		return
	}
}

func main() {
	//inicializa o ponto de recarga e conecta ao servidor
	logger := logger.NewLogger(os.Stdout)
	conexao, erro := tcpIP.ConnectToServerTCP("servidor:5000")
	if erro != nil {
		logger.Erro("Erro em ConnectToServerTCP - ponto de recarga")
		return
	}
	logger.Info("Ponto de Recarga conectado")
	defer conexao.Close()
	//envia identificacao inicial
	IdentificacaoInicial(logger, conexao)

	//recebe solicitacoes do servidor
	for {
		solicitacaoServidor, erro := dataJson.ReceiveMessage(conexao)
		if erro != nil {
			logger.Erro(fmt.Sprintf("Erro ao ler mensagem do servidor - %v", erro))
			return
		}

		if solicitacaoServidor.Tipo == "get-disponibilidade" {
			enviarDisponibilidade(logger, conexao)
		} else {
			logger.Info(fmt.Sprintf("Mensagem recebida do servidor: %s", solicitacaoServidor.Conteudo))
		}
	}
}
