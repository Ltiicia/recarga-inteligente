package main

import (
	"fmt"
	"os"
	"recarga-inteligente/internal/coordenadas"
	"recarga-inteligente/internal/dataJson"
	"recarga-inteligente/internal/logger"
	"recarga-inteligente/internal/tcpIP"
)

func main() {
	logger := logger.NewLogger(os.Stdout)

	conexao, erro := tcpIP.ConnectToServerTCP("servidor:5000")
	if erro != nil {
		logger.Erro("Erro em ConnectToServerTCP - veiculo")
		return
	}
	logger.Info("Veiculo conectado")
	defer conexao.Close()

	respostaServidor, erro := tcpIP.SendIdentification(conexao, "veiculo")

	if erro != nil {
		logger.Erro(fmt.Sprintf("Erro ao obter resposta do servidor - %v", erro))
		return
	}
	logger.Info(fmt.Sprintf("Mensagem recebida do servidor - %s", respostaServidor.Conteudo))

	if respostaServidor.Tipo == "get-localizacao" {
		var dadosRegiao dataJson.DadosRegiao
		dadosRegiao, erro = dataJson.OpenFile("regiao.json")
		if erro != nil {
			logger.Erro(fmt.Sprintf("Erro ao obter dados da regiao - %v", erro))
			return
		}

		localizacaoAtual := coordenadas.GetLocalizacaoVeiculo(dadosRegiao.Area)

		msg := dataJson.Mensagem{
			Tipo:     "localizacao",
			Conteudo: fmt.Sprintf("%f,%f", localizacaoAtual.Latitude, localizacaoAtual.Longitude),
			Origem:   "veiculo",
		}
		erro := dataJson.SendMessage(conexao, msg)
		if erro != nil {
			logger.Erro(fmt.Sprintf("Erro ao enviar localizacao - %v", erro))
		}
	}

	for {
		respostaServidor, erro := dataJson.ReceiveMessage(conexao)
		if erro != nil {
			logger.Erro(fmt.Sprintf("Erro ao ler mensagem do servidor - %v", erro))
			return
		}

		logger.Info(fmt.Sprintf("Mensagem recebida do servidor - %s", respostaServidor.Conteudo))
	}
}
