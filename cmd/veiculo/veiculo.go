package main

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

func enviarLocalizacao(logger *logger.Logger, conexao net.Conn) {
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

func menuVeiculo() string {
	leitor := bufio.NewReader(os.Stdin)

	fmt.Println("\n==== Menu Veiculo ====")
	fmt.Println("(1) - Solicitar recarga")
	fmt.Println("(2) - Consultar pagamentos pendentes")
	fmt.Println("(3) - Sair")
	fmt.Print("Selecione uma opcao: ")
	opcao, _ := leitor.ReadString('\n')
	opcao = strings.TrimSpace(opcao)
	return opcao
}

func main() {
	logger := logger.NewLogger(os.Stdout)

	conexao, erro := tcpIP.ConnectToServerTCP("servidor:5000")
	if erro != nil {
		logger.Erro("Erro em ConnectToServerTCP - veiculo")
		return
	}
	logger.Info("Veiculo conectado")
	defer conexao.Close()

	var respostaServidor dataJson.Mensagem

	for {
		opcao := menuVeiculo()

		switch opcao {
		case "1": //recarga
			respostaServidor, erro = tcpIP.SendIdentification(conexao, "veiculo")
			if erro != nil {
				logger.Erro(fmt.Sprintf("Erro ao obter resposta do servidor - %v", erro))
				return
			}

			logger.Info(fmt.Sprintf("Mensagem recebida do servidor - %s", respostaServidor.Conteudo))

			if respostaServidor.Tipo == "get-localizacao" {
				enviarLocalizacao(logger, conexao)
				//recebeAsMelhoresOpcoes
				//Escolhe uma p fazer a reserva
				//carrega o veiculo
				//salva o valor da recarga
				//retorna ao menu
			}
		case "2": //pagamento
			msg := dataJson.Mensagem{
				Tipo:     "CONSULTAR_PAGAMENTO",
				Conteudo: "Ola servidor! Gostaria de consultar pagamentos pendentes",
				Origem:   "veiculo",
			}
			dataJson.SendMessage(conexao, msg)
			fmt.Println("Em construcao...")
			continue
		case "3":
			fmt.Println("Saindo...")
			return
		default:
			fmt.Println("Opcao invalida. Tente novamente.")
		}
	}
	/*
		for {
			respostaServidor, erro := dataJson.ReceiveMessage(conexao)
			if erro != nil {
				logger.Erro(fmt.Sprintf("Erro ao ler mensagem do servidor - %v", erro))
				return
			}

			logger.Info(fmt.Sprintf("Mensagem recebida do servidor - %s", respostaServidor.Conteudo))
		}*/
}
