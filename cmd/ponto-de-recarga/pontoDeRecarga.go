package main

import (
	"fmt"
	"net"
	"os"
	"recarga-inteligente/internal/dataJson"
	"recarga-inteligente/internal/logger"
	"recarga-inteligente/internal/tcpIP"
)

func confirmaRecebimentoId(idPonto string, logger *logger.Logger, conexao net.Conn) {
	msg := dataJson.Mensagem{
		Tipo:     "return-id",
		Conteudo: fmt.Sprintf("%s", idPonto),
		Origem:   "ponto-de-recarga",
	}
	erro := dataJson.SendMessage(conexao, msg)
	if erro != nil {
		logger.Erro(fmt.Sprintf("Erro ao enviar disponibilidade - %v", erro))
	}
}

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
	//envia identificacao ao servidor
	respostaServidor, erro := tcpIP.SendIdentification(conexao, "ponto-de-recarga")
	if erro != nil {
		logger.Erro(fmt.Sprintf("Erro ao obter resposta do servidor - %v", erro))
		return
	}
	//recebe id atribuido
	idPonto := respostaServidor.Conteudo
	if respostaServidor.Tipo == "id" {
		//envia ccnfirmacao
		confirmaRecebimentoId(idPonto, logger, conexao)
	}
	//recebe mensagens do servidor
	for {
		respostaServidor, erro := dataJson.ReceiveMessage(conexao)
		if erro != nil {
			logger.Erro(fmt.Sprintf("Erro ao ler mensagem do servidor - %v", erro))
			return
		}

		if respostaServidor.Tipo == "get-disponibilidade" {
			enviarDisponibilidade(logger, conexao)
		}

		logger.Info(fmt.Sprintf("Mensagem recebida do servidor: %s", respostaServidor.Conteudo))
	}
}
