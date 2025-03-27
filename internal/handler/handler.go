package handler

import (
	"fmt"
	"net"
	"recarga-inteligente/internal/dataJson"
	"recarga-inteligente/internal/logger"
	"recarga-inteligente/internal/store"
)

// Trata a comunicacao com os clientes
func HandleConnection(conexao net.Conn, connectionStore *store.ConnectionStore, logger *logger.Logger) {
	defer connectionStore.RemoveConnection(conexao)

	//recebe mensagem inicial
	mensagemInicial, erro := dataJson.ReceiveMessage(conexao)
	if erro != nil {
		logger.Erro(fmt.Sprintf("Erro ao ler mensagem inicial do %s: %v", mensagemInicial.Origem, erro))
		return
	}

	tipoCliente := mensagemInicial.Origem
	if tipoCliente != "veiculo" && tipoCliente != "ponto-de-recarga" {
		logger.Erro(fmt.Sprintf("Origem desconhecida, encerrando conexao de: %s", tipoCliente))
		conexao.Close()
		return
	}

	connectionStore.AddConnection(conexao, tipoCliente)
	logger.Info(fmt.Sprintf("Novo %s conectado: (%s)", tipoCliente, conexao.RemoteAddr()))

	//Personaliza a resposta
	var mensagemResposta dataJson.Mensagem
	if tipoCliente == "veiculo" {
		mensagemResposta = dataJson.Mensagem{
			Tipo:     "get-localizacao",
			Conteudo: "Ola Veiculo! Informe sua localizacao atual.",
			Origem:   "servidor",
		}
	} else if tipoCliente == "ponto-de-recarga" {
		mensagemResposta = dataJson.Mensagem{
			Tipo:     "get-disponibilidade",
			Conteudo: "Ola Ponto de Recarga! Informe sua disponibilidade.",
			Origem:   "servidor",
		}
	}
	//envia
	erro = dataJson.SendMessage(conexao, mensagemResposta)
	if erro != nil {
		logger.Erro(fmt.Sprintf("Erro ao enviar saudacao ao %s: %v", tipoCliente, erro))
		return
	}
	logger.Info(mensagemResposta.Conteudo)
	for {
		mensagemRecebida, erro := dataJson.ReceiveMessage(conexao)
		if erro != nil {
			logger.Erro(fmt.Sprintf("Erro ao ler mensagem do %s: %v", tipoCliente, erro))
			return
		}
		logger.Info(fmt.Sprintf("Mensagem recebida do %s (%s) => %s", tipoCliente, conexao.RemoteAddr(), mensagemRecebida.Conteudo))
	}
	logger.Info(fmt.Sprintf("%s desconectado: %s", tipoCliente, conexao.RemoteAddr()))
}
