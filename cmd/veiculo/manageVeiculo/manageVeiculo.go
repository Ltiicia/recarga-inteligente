package manageVeiculo

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"recarga-inteligente/internal/coordenadas"
	"recarga-inteligente/internal/dataJson"
	"recarga-inteligente/internal/logger"
	"recarga-inteligente/internal/tcpIP"
	"strings"
)

func EnviarLocalizacao(logger *logger.Logger, conexao net.Conn) {
	dadosRegiao, _, erro := dataJson.ReceiveDadosRegiao(conexao)
	if erro != nil {
		logger.Erro(fmt.Sprintf("Erro ao receber dados da regiao - %v", erro))
		return
	}
	localizacaoAtual := coordenadas.GetLocalizacaoVeiculo(dadosRegiao.Area)
	msg_localizacao := dataJson.Mensagem{
		Tipo:     "localizacao",
		Conteudo: fmt.Sprintf("%f,%f", localizacaoAtual.Latitude, localizacaoAtual.Longitude),
		Origem:   "veiculo",
	}
	erro = dataJson.SendMessage(conexao, msg_localizacao)
	if erro != nil {
		logger.Erro(fmt.Sprintf("Erro ao enviar localizacao - %v", erro))
		return
	}
}

func IdentificacaoInicial(logger *logger.Logger, conexao net.Conn) {
	erro := tcpIP.SendIdentification(conexao, "veiculo")
	if erro != nil {
		logger.Erro(fmt.Sprintf("Erro ao obter resposta do servidor - %v", erro))
		return
	}
}

func SolicitarRecarga(logger *logger.Logger, conexao net.Conn) dataJson.Mensagem {
	solicitacao := dataJson.Mensagem{
		Tipo:     "get-recarga",
		Conteudo: "Ola Servidor! Quero recarregar",
		Origem:   "veiculo",
	}
	erro := dataJson.SendMessage(conexao, solicitacao)
	if erro != nil {
		logger.Erro(fmt.Sprintf("Erro ao solicitar recarga: %v", erro))
		return dataJson.Mensagem{}
	}
	resposta, erro := dataJson.ReceiveMessage(conexao)
	if erro != nil {
		logger.Erro(fmt.Sprintf("Erro ao obter resposta da solicitacao de recarga : %v", erro))
		return dataJson.Mensagem{}
	}
	return resposta
}

func MenuVeiculo(logger *logger.Logger, conexao net.Conn) {
	leitor := bufio.NewReader(os.Stdin)
	on := true
	//veiculo envia sua identificacao ao servidor e recebe resposta
	IdentificacaoInicial(logger, conexao)

	for on {
		fmt.Println("\n==== Menu Veiculo ====")
		fmt.Println("(1) - Solicitar recarga")
		fmt.Println("(2) - Consultar pagamentos pendentes")
		fmt.Println("(3) - Sair")
		fmt.Print("Selecione uma opcao: ")
		opcao, _ := leitor.ReadString('\n')
		opcao = strings.TrimSpace(opcao)

		switch opcao {
		case "1": //solicitou recarga
			respostaServidor := SolicitarRecarga(logger, conexao)
			if respostaServidor.Tipo == "get-localizacao" {
				EnviarLocalizacao(logger, conexao)
			}
			//recebeAsMelhoresOpcoes
			//Escolhe uma p fazer a reserva
			//carrega o veiculo
			//salva o valor da recarga
			//retorna ao menu

		case "2": //pagamento
			msg := dataJson.Mensagem{
				Tipo:     "consultar_pagamento",
				Conteudo: "Ola servidor! Gostaria de consultar pagamentos pendentes",
				Origem:   "veiculo",
			}
			erro := dataJson.SendMessage(conexao, msg)
			if erro != nil {
				logger.Erro(fmt.Sprintf("Erro ao enviar consulta de pagamento - %v", erro))
				return
			}
			fmt.Println("Em construcao...")
		case "3":
			fmt.Println("Saindo...")
			conexao.Close()
			on = false
		default:
			fmt.Println("Opcao invalida. Tente novamente.")
		}
	}
}
