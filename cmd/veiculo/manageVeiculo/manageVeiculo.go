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

func MenuVeiculo(logger *logger.Logger, conexao net.Conn) {
	leitor := bufio.NewReader(os.Stdin)
	on := true

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
			//veiculo envia sua identificacao ao servidor e recebe resposta
			respostaServidor, erro := tcpIP.SendIdentification(conexao, "veiculo")
			if erro != nil {
				logger.Erro(fmt.Sprintf("Erro ao obter resposta do servidor - %v", erro))
				return
			}
			logger.Info(fmt.Sprintf("Mensagem recebida do servidor - %s", respostaServidor.Conteudo))

			if respostaServidor.Tipo == "get-localizacao" {
				EnviarLocalizacao(logger, conexao)
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
			erro := dataJson.SendMessage(conexao, msg)
			if erro != nil {
				logger.Erro(fmt.Sprintf("Erro ao enviar consulta de pagamento - %v", erro))
				return
			}
			fmt.Println("Em construcao...")
		case "3":
			fmt.Println("Saindo...")
			on = false
		default:
			fmt.Println("Opcao invalida. Tente novamente.")
		}
	}
}
