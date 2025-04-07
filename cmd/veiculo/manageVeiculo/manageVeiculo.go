package manageVeiculo

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"recarga-inteligente/internal/coordenadas"
	"recarga-inteligente/internal/dataJson"
	"recarga-inteligente/internal/logger"
	"strconv"
	"strings"
	"time"
)

// Corrigir o bug na função EnviarLocalizacao:
func EnviarLocalizacao(logger *logger.Logger, conexao net.Conn) bool {
	dadosRegiao, _, erro := dataJson.ReceiveDadosRegiao(conexao)
	if erro != nil {
		logger.Erro(fmt.Sprintf("Erro ao receber dados da regiao - %v", erro))
		return false
		// remover o "return placa" errado que está aqui
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
		return false
	}
	fmt.Println("Localização enviada, aguardando ranking de pontos...")
	return true
}

// Corrigir a função processarRankingPontos para exibir mensagens limpas:

func processarRankingPontos(logger *logger.Logger, conexao net.Conn, placa string) {
	// Esperar a resposta com o ranking
	resposta, erro := dataJson.ReceiveMessage(conexao)
	if erro != nil {
		logger.Erro(fmt.Sprintf("Erro ao receber ranking: %v", erro))
		return
	}

	if resposta.Tipo != "ranking-pontos" {
		logger.Erro(fmt.Sprintf("Tipo de resposta inesperado: %s", resposta.Tipo))
		return
	}

	// Limpar a tela e mostrar o ranking formatado de maneira mais limpa
	linhas := strings.Split(resposta.Conteudo, "\n")
	fmt.Println("\n----- RANKING DE PONTOS DE RECARGA -----")
	for _, linha := range linhas {
		if strings.TrimSpace(linha) != "" {
			fmt.Println(linha)
		}
	}
	fmt.Println("-----------------------------------------")

	// Solicitar escolha do usuário
	leitor := bufio.NewReader(os.Stdin)
	fmt.Print("\nSelecione o número do ponto de recarga (1-3): ")
	escolha, _ := leitor.ReadString('\n')
	escolha = strings.TrimSpace(escolha)

	indice, erro := strconv.Atoi(escolha)
	if erro != nil || indice < 1 || indice > 3 {
		fmt.Println("❌ Escolha inválida. Retornando ao menu principal.")
		return
	}

	// Extrair o ID do ponto da linha escolhida
	if indice <= 0 || indice > len(linhas)-1 {
		fmt.Println("❌ Escolha fora do range de opções disponíveis.")
		return
	}

	// Usar uma variável temporária para o número da opção para não sobrescrever o índice
	var numeroOpcao int
	var pontoID int
	_, err := fmt.Sscanf(linhas[indice-1], "%d. Ponto ID: %d", &numeroOpcao, &pontoID)
	if err != nil {
		logger.Erro(fmt.Sprintf("Erro ao obter ID do ponto: %v", err))
		fmt.Println("❌ Erro ao processar sua escolha. Tente novamente.")
		return
	}

	// Enviar solicitação de reserva
	msg := dataJson.Mensagem{
		Tipo:     "solicitar-reserva",
		Conteudo: fmt.Sprintf("%d", pontoID),
		Origem:   "veiculo",
	}

	erro = dataJson.SendMessage(conexao, msg)
	if erro != nil {
		logger.Erro(fmt.Sprintf("Erro ao enviar solicitação de reserva: %v", erro))
		fmt.Println("❌ Erro de comunicação. Tente novamente.")
		return
	}

	fmt.Printf("\n⏳ Reserva solicitada para o ponto ID %d. Aguardando confirmação...\n", pontoID)

	// Aguardar confirmação
	confirmacao, erro := dataJson.ReceiveMessage(conexao)
	if erro != nil {
		logger.Erro(fmt.Sprintf("Erro ao receber confirmação: %v", erro))
		fmt.Println("❌ Erro ao receber confirmação. Tente novamente.")
		return
	}

	// Processar a resposta do servidor de forma limpa
	fmt.Println("\n----- STATUS DA RESERVA -----")

	if confirmacao.Tipo == "reserva-confirmada" || confirmacao.Tipo == "sua-vez" {
		// Se for mensagem de confirmação de reserva
		if confirmacao.Tipo == "reserva-confirmada" {
			fmt.Println("✅ " + confirmacao.Conteudo)

			// Não devemos nos deslocar imediatamente, apenas confirmar a reserva
			// e aguardar nossa vez na fila
			if strings.Contains(confirmacao.Conteudo, "Você é o próximo") {
				fmt.Println("Aguardando autorização para iniciar deslocamento...")
			} else if strings.Contains(confirmacao.Conteudo, "fila") {
				fmt.Println("Aguardando sua vez na fila...")
			}
		}

		// Se já for diretamente nossa vez (sem passar pela fila)
		if confirmacao.Tipo == "sua-vez" {
			fmt.Println("🔄 É sua vez! Autorizado a se deslocar ao ponto de recarga.")
			fmt.Println("🚗 Iniciando deslocamento até o ponto de recarga...")
			time.Sleep(10 * time.Second) // Simulando deslocamento

			// Informar ao servidor que chegou
			msgChegada := dataJson.Mensagem{
				Tipo:     "veiculo-chegou",
				Conteudo: placa,
				Origem:   "veiculo",
			}
			dataJson.SendMessage(conexao, msgChegada)
			fmt.Println("🚶 Chegou ao ponto de recarga, aguardando início do carregamento...")
		}

		// Loop para receber mensagens do servidor enquanto aguarda
		recargaConcluida := make(chan struct{})
		erroRecarga := make(chan error)

		go func() {
			for {
				mensagem, erro := dataJson.ReceiveMessage(conexao)
				if erro != nil {
					erroRecarga <- erro
					return
				}

				switch mensagem.Tipo {
				case "posicao-fila":
					// Mostrar posição na fila
					fmt.Println("🔄 " + mensagem.Conteudo)

				case "sua-vez":
					// Agora é a vez do veículo - deve iniciar deslocamento
					fmt.Println("🔄 É sua vez! Autorizado a se deslocar ao ponto de recarga.")
					fmt.Println("🚗 Iniciando deslocamento até o ponto de recarga...")
					time.Sleep(10 * time.Second) // Simulando deslocamento

					// Informar ao servidor que chegou
					msgChegada := dataJson.Mensagem{
						Tipo:     "veiculo-chegou",
						Conteudo: placa,
						Origem:   "veiculo",
					}
					dataJson.SendMessage(conexao, msgChegada)
					fmt.Println("🚶 Chegou ao ponto de recarga, aguardando início do carregamento...")

				case "recarga-iniciada":
					// Só agora inicia-se o carregamento de fato
					fmt.Println("🔌 Iniciando carregamento...")
					fmt.Println("⚡ " + mensagem.Conteudo)

				case "recarga-finalizada":
					fmt.Println("✅ " + mensagem.Conteudo)
					fmt.Println("Recarga concluída! Retornando ao menu principal...")
					close(recargaConcluida)
					return

				default:
					fmt.Println("Mensagem recebida: " + mensagem.Conteudo)
				}
			}
		}()

		select {
		case <-recargaConcluida:
			return
		case err := <-erroRecarga:
			logger.Erro(fmt.Sprintf("Erro durante processo de recarga: %v", err))
			return
		case <-time.After(5 * time.Minute):
			logger.Erro("Timeout aguardando conclusão da recarga")
			return
		}
	} else if confirmacao.Tipo == "reserva-falhou" {
		fmt.Println("❌ " + confirmacao.Conteudo)
		fmt.Println("Retornando ao menu principal...")
		return
	} else {
		// Tratar outros tipos de mensagens inesperadas
		logger.Erro(fmt.Sprintf("Tipo de confirmação inesperado: %s", confirmacao.Tipo))
		fmt.Println("❓ Resposta inesperada do servidor. Tente novamente.")
		fmt.Println("-------------------------------")
		return
	}
}

// Modificar a função IdentificacaoInicial para verificar com o servidor
func IdentificacaoInicial(logger *logger.Logger, conexao net.Conn) string {
	leitor := bufio.NewReader(os.Stdin)
	placa := ""
	placaValida := false

	for !placaValida {
		fmt.Println("Por favor, informe a placa do veículo (6-8 caracteres): ")
		input, _ := leitor.ReadString('\n')
		placa = strings.TrimSpace(input)

		// Validar formato da placa
		if len(placa) < 6 || len(placa) > 8 {
			fmt.Println("❌ Placa inválida! A placa deve ter entre 6 e 8 caracteres.")
			continue
		}

		// Enviar solicitação ao servidor para verificar se a placa está disponível
		msg := dataJson.Mensagem{
			Tipo:     "verificar-placa",
			Conteudo: placa,
			Origem:   "veiculo",
		}

		erro := dataJson.SendMessage(conexao, msg)
		if erro != nil {
			logger.Erro(fmt.Sprintf("Erro ao verificar placa: %v", erro))
			return ""
		}

		// Esperar resposta do servidor
		resposta, erro := dataJson.ReceiveMessage(conexao)
		if erro != nil {
			logger.Erro(fmt.Sprintf("Erro ao receber resposta de verificação de placa: %v", erro))
			return ""
		}

		if resposta.Tipo == "placa-disponivel" {
			placaValida = true
		} else if resposta.Tipo == "placa-indisponivel" {
			fmt.Println("❌ Esta placa já está em uso por outro veículo!")
		} else {
			logger.Erro(fmt.Sprintf("Resposta inesperada do servidor: %s", resposta.Tipo))
			return ""
		}
	}

	// Agora que sabemos que a placa é válida, enviar a identificação final
	msg := dataJson.Mensagem{
		Tipo:     "identificacao",
		Conteudo: fmt.Sprintf("veiculo conectado placa %s", placa),
		Origem:   "veiculo",
	}

	erro := dataJson.SendMessage(conexao, msg)
	if erro != nil {
		logger.Erro(fmt.Sprintf("Erro ao enviar identificacao: %v", erro))
		return ""
	}

	return placa
}

// Nova função para verificar se a placa já existe
func placaJaExiste(placa string) bool {
	path := filepath.Join("app", "internal", "dataJson", "veiculos.json")

	// Tentar abrir o arquivo
	file, err := os.Open(path)
	if err != nil {
		// Se o arquivo não existe, a placa não existe
		if os.IsNotExist(err) {
			return false
		}
		// Em caso de outros erros, assumimos que não conseguimos verificar
		// então retornamos false para permitir o cadastro
		return false
	}
	defer file.Close()

	// Decodificar dados existentes
	var dadosVeiculos dataJson.DadosVeiculos
	if err := json.NewDecoder(file).Decode(&dadosVeiculos); err != nil {
		// Se não conseguir decodificar, também assumimos que a placa não existe
		return false
	}

	// Verificar se a placa já existe
	for _, v := range dadosVeiculos.Veiculos {
		if v.Placa == placa {
			return true // Placa já existe
		}
	}

	return false // Placa não existe
}

// Também precisamos corrigir a função MenuVeiculo para não chamar IdentificacaoInicial duas vezes:
func MenuVeiculo(logger *logger.Logger, conexao net.Conn) {
	leitor := bufio.NewReader(os.Stdin)
	on := true
	placa := IdentificacaoInicial(logger, conexao)

	// Verificar se a identificação falhou
	if placa == "" {
		logger.Erro("Falha na identificação do veículo")
		return
	}

	fmt.Printf("Veículo com placa %s registrado com sucesso!\n", placa)

	for on {
		fmt.Println("\n==== Menu Veiculo ====")
		fmt.Println("(1) - Solicitar recarga")
		fmt.Println("(2) - Consultar pagamentos pendentes")
		fmt.Println("(3) - Sair")
		fmt.Println("Selecione uma opcao: ")
		opcao, _ := leitor.ReadString('\n')
		opcao = strings.TrimSpace(opcao)

		switch opcao {
		case "1":
			SolicitarRecarga(logger, conexao, placa)
			//recebeAsMelhoresOpcoes
			//Escolhe uma p fazer a reserva
			//carrega o veiculo
			//salva o valor da recarga
			//retorna ao menu

		case "2": // consultar pagamentos/recargas
			// Enviar solicitação ao servidor
			msgConsulta := dataJson.Mensagem{
				Tipo:     "consultar-historico",
				Conteudo: placa,
				Origem:   "veiculo",
			}

			erro := dataJson.SendMessage(conexao, msgConsulta)
			if erro != nil {
				logger.Erro(fmt.Sprintf("Erro ao solicitar histórico: %v", erro))
				fmt.Println("❌ Erro ao consultar pagamentos. Tente novamente mais tarde.")
				continue
			}

			// Aguardar resposta do servidor
			resposta, erro := dataJson.ReceiveMessage(conexao)
			if erro != nil {
				logger.Erro(fmt.Sprintf("Erro ao receber histórico: %v", erro))
				fmt.Println("❌ Erro ao consultar pagamentos. Tente novamente mais tarde.")
				continue
			}

			if resposta.Tipo == "historico-erro" {
				fmt.Println("❌ " + resposta.Conteudo)
				continue
			}

			if resposta.Tipo != "historico-recargas" {
				logger.Erro(fmt.Sprintf("Tipo de resposta inesperado: %s", resposta.Tipo))
				fmt.Println("❌ Resposta inesperada do servidor. Tente novamente mais tarde.")
				continue
			}

			// Deserializar o histórico
			var recargas []dataJson.Recarga
			erro = json.Unmarshal([]byte(resposta.Conteudo), &recargas)
			if erro != nil {
				logger.Erro(fmt.Sprintf("Erro ao deserializar histórico: %v", erro))
				fmt.Println("❌ Erro ao processar histórico recebido. Tente novamente mais tarde.")
				continue
			}

			// Exibir o histórico para o usuário
			if len(recargas) == 0 {
				fmt.Println("Nenhum histórico de pagamento encontrado para este veículo.")
			} else {
				fmt.Println("\n==== Histórico de Recargas ====")
				fmt.Println("Data                | Ponto ID | Valor (R$)")
				fmt.Println("------------------------------------------")

				valorTotal := 0.0
				for _, p := range recargas {
					fmt.Printf("%s | %d        | R$ %.2f\n", p.Data, p.PontoID, p.Valor)
					valorTotal += p.Valor
				}

				fmt.Println("------------------------------------------")
				fmt.Printf("Total: R$ %.2f\n", valorTotal)
			}

		case "3":
			fmt.Println("Saindo...")
			conexao.Close()
			on = false
		default:
			fmt.Println("Opcao invalida. Tente novamente.")
		}
	}
}

func SolicitarRecarga(logger *logger.Logger, conexao net.Conn, placa string) {
	solicitacao := dataJson.Mensagem{
		Tipo:     "get-recarga",
		Conteudo: "Ola Servidor! Quero recarregar",
		Origem:   "veiculo",
	}
	erro := dataJson.SendMessage(conexao, solicitacao)
	if erro != nil {
		logger.Erro(fmt.Sprintf("Erro ao solicitar recarga: %v", erro))
		return
	}

	// Aguardar resposta do servidor solicitando localização
	resposta, erro := dataJson.ReceiveMessage(conexao)
	if erro != nil {
		logger.Erro(fmt.Sprintf("Erro ao obter resposta da solicitacao de recarga: %v", erro))
		return
	}

	if resposta.Tipo == "get-localizacao" {
		// Enviar localização
		if !EnviarLocalizacao(logger, conexao) {
			return
		}

		// Processar o ranking e fazer reserva
		processarRankingPontos(logger, conexao, placa)
	} else {
		logger.Erro(fmt.Sprintf("Resposta inesperada do servidor: %s", resposta.Tipo))
	}
}

// Adicione essa função no final do arquivo:

// Função para consultar o histórico de pagamentos de um veículo
func consultarPagamentosVeiculo(placa string) ([]dataJson.Recarga, error) {
	path := filepath.Join("app", "internal", "dataJson", "veiculos.json")

	// Tentar abrir o arquivo
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Se o arquivo não existe, não há histórico
			return []dataJson.Recarga{}, nil
		}
		return nil, err
	}
	defer file.Close()

	// Decodificar dados existentes
	var dadosVeiculos dataJson.DadosVeiculos
	if err := json.NewDecoder(file).Decode(&dadosVeiculos); err != nil {
		return nil, err
	}

	// Procurar o veículo pela placa
	for _, v := range dadosVeiculos.Veiculos {
		if v.Placa == placa {
			// Retornar o histórico de recargas desse veículo
			return v.Recargas, nil
		}
	}

	// Veículo não encontrado (não deveria acontecer se o veículo está logado)
	return []dataJson.Recarga{}, nil
}
