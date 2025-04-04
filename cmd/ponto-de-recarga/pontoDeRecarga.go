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
		logger.Erro("Erro em ConnectToServerTCP - ponto de recarga")
		return
	}
	logger.Info("Ponto de Recarga conectado")
	defer conexao.Close()

	var respostaServidor dataJson.Mensagem
	respostaServidor, erro = tcpIP.SendIdentification(conexao, "ponto-de-recarga")
	idPonto := respostaServidor.Conteudo
	if erro != nil {
		conexao.Close()
		return
	}

	if respostaServidor.Tipo == "id" {
		msg := dataJson.Mensagem{
			Tipo:     "registro-id",
			Conteudo: "ID registrado",
			Origem:   fmt.Sprintf("ponto-de-recarga-%s", idPonto),
		}
		erro := dataJson.SendMessage(conexao, msg)
		if erro != nil {
			logger.Erro(fmt.Sprintf("Erro ao enviar disponibilidade - %v", erro))
		}
	}

	for {
		respostaServidor, erro := dataJson.ReceiveMessage(conexao)
		if erro != nil {
			logger.Erro(fmt.Sprintf("Erro ao ler mensagem do servidor - %v", erro))
			return
		}

		if respostaServidor.Tipo == "get-disponibilidade" {
			msg := dataJson.Mensagem{
				Tipo:     "disponibilidade",
				Conteudo: "Situacao atual: sem fila",
				Origem:   "ponto-de-recarga",
			}
			erro = dataJson.SendMessage(conexao, msg)
			if erro != nil {
				logger.Erro(fmt.Sprintf("Erro ao enviar disponibilidade - %v", erro))
			}
		}

		logger.Info(fmt.Sprintf("Mensagem recebida do servidor: %s", respostaServidor.Conteudo))
	}
}
