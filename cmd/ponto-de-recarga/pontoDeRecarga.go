package main

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"

	"recarga-inteligente/internal/dataJson"
	"recarga-inteligente/internal/logger"
	"recarga-inteligente/internal/tcpIP"
)

var fila []string
var mutex sync.Mutex
var veiculosEmEspera map[string]chan bool

// Adicionar um canal para sinalizar processamento de próximo veículo
var (
	proximoVeiculoSignal = make(chan struct{}, 1)
)

func enviarDisponibilidade(logger *logger.Logger, conexao net.Conn) {
	mutex.Lock()
	defer mutex.Unlock()

	status := "Situacao atual: "
	if len(fila) == 0 {
		status += "sem fila"
	} else {
		status += fmt.Sprintf("com %d na fila", len(fila))
	}

	msg := dataJson.Mensagem{
		Tipo:     "disponibilidade",
		Conteudo: status,
		Origem:   "ponto-de-recarga",
	}
	erro := dataJson.SendMessage(conexao, msg)
	if erro != nil {
		logger.Erro(fmt.Sprintf("Erro ao enviar disponibilidade - %v", erro))
	}
}

// Modificar a função processarFila para torná-la mais robusta:

func processarFila(logger *logger.Logger, conexao net.Conn) {
	for {
		// Verificar se há veículos na fila
		mutex.Lock()
		if len(fila) == 0 {
			mutex.Unlock()
			// Esperar por um sinal ou timeout antes de verificar novamente
			select {
			case <-proximoVeiculoSignal:
				// Recebeu sinal para processar próximo veículo
				logger.Info("Recebido sinal para processar próximo veículo")
				continue
			case <-time.After(1 * time.Second):
				// Timeout normal, verificar novamente
				continue
			}
		}

		// Há pelo menos um veículo na fila - pegar o primeiro
		veiculoAtual := fila[0]
		logger.Info(fmt.Sprintf("Processando próximo veículo na fila: %s", veiculoAtual))
		mutex.Unlock()

		// 1. Notificar o servidor que estamos chamando este veículo
		msg := dataJson.Mensagem{
			Tipo:     "chamando-veiculo",
			Conteudo: veiculoAtual, // Usar a placa do veículo consistentemente
			Origem:   "ponto-de-recarga",
		}

		erro := dataJson.SendMessage(conexao, msg)
		if erro != nil {
			logger.Erro(fmt.Sprintf("Erro ao notificar servidor sobre veículo em atendimento: %v", erro))
			time.Sleep(2 * time.Second) // Esperar um pouco antes de tentar novamente
			continue
		}

		// 2. Criar canal para aguardar chegada do veículo
		chegou := make(chan bool, 1)

		// Importante: Lock antes de modificar veiculosEmEspera
		mutex.Lock()
		veiculosEmEspera[veiculoAtual] = chegou
		mutex.Unlock()

		logger.Info(fmt.Sprintf("Aguardando chegada do veículo: %s", veiculoAtual))

		// 3. Aguardar com timeout a chegada do veículo
		timeout := 60 * time.Second // Um minuto para o veículo chegar
		select {
		case <-chegou:
			logger.Info(fmt.Sprintf("Veículo %s informou chegada, iniciando carregamento", veiculoAtual))
			// Continuar com o carregamento
		case <-time.After(timeout):
			logger.Erro(fmt.Sprintf("Timeout aguardando veículo %s, removendo da fila", veiculoAtual))

			// Remover da fila e do mapa de espera
			mutex.Lock()
			if len(fila) > 0 && fila[0] == veiculoAtual {
				fila = fila[1:] // Remover da fila
			}
			delete(veiculosEmEspera, veiculoAtual) // Remover do mapa
			mutex.Unlock()
			continue // Processar próximo veículo
		}

		// 4. Simulação do carregamento - só chega aqui se o veículo chegou
		logger.Info(fmt.Sprintf("Iniciando carregamento para: %s", veiculoAtual))
		time.Sleep(20 * time.Second) // Tempo de carregamento simulado

		// 5. Cálculo do valor (como já implementado)
		tempoHoras := 1.0 + rand.Float64()      // Entre 1 e 2 horas
		taxaKwh := 0.80                         // Taxa por kWh
		potenciaKw := 22.0                      // Potência em kW
		consumoTotal := potenciaKw * tempoHoras // Consumo total em kWh
		valor := consumoTotal * taxaKwh         // Valor total

		logger.Info(fmt.Sprintf("Recarga finalizada para: %s - Consumo: %.2f kWh, Valor: R$ %.2f",
			veiculoAtual, consumoTotal, valor))

		// 6. Remover o veículo da fila e do mapa de espera
		mutex.Lock()
		if len(fila) > 0 && fila[0] == veiculoAtual {
			fila = fila[1:] // Remover o primeiro elemento
		}
		delete(veiculosEmEspera, veiculoAtual)
		mutex.Unlock()

		// 7. Notificar o servidor que a recarga foi concluída
		msgFinalizada := dataJson.Mensagem{
			Tipo: "recarga-finalizada",
			Conteudo: fmt.Sprintf("Veículo %s atendido. Consumo: %.2f kWh, Valor: R$ %.2f",
				veiculoAtual, consumoTotal, valor),
			Origem: "ponto-de-recarga",
		}

		dataJson.SendMessage(conexao, msgFinalizada)

		// Pequena pausa antes de processar o próximo veículo
		time.Sleep(1 * time.Second)
	}
}

// Quando o ponto chama o próximo veículo
func chamarProximoVeiculo(conexao net.Conn, logger *logger.Logger) {
	mutex.Lock()
	if len(fila) == 0 {
		mutex.Unlock()
		return
	}

	veiculoID := fila[0]
	mutex.Unlock()

	// Notificar o servidor que está chamando este veículo
	msgChamando := dataJson.Mensagem{
		Tipo:     "chamando-veiculo",
		Conteudo: veiculoID,
		Origem:   "ponto-de-recarga",
	}

	// Enviar em uma goroutine para não bloquear
	go func() {
		erro := dataJson.SendMessage(conexao, msgChamando)
		if erro != nil {
			logger.Erro(fmt.Sprintf("Erro ao notificar servidor sobre chamada do veículo %s: %v",
				veiculoID, erro))
		}
	}()

	// O restante do código...
}

func IdentificacaoInicial(logger *logger.Logger, conexao net.Conn) {
	erro := tcpIP.SendIdentification(conexao, "ponto-de-recarga")
	if erro != nil {
		logger.Erro(fmt.Sprintf("Erro ao obter resposta do servidor - %v", erro))
		return
	}
}

func main() {
	// Inicializar o gerador de números aleatórios
	rand.Seed(time.Now().UnixNano())

	// Inicializar o mapa de veículos em espera
	veiculosEmEspera = make(map[string]chan bool)

	//inicializa o ponto de recarga e conecta ao servidor
	logger := logger.NewLogger(os.Stdout)
	conexao, erro := tcpIP.ConnectToServerTCP("servidor:5000")
	if erro != nil {
		logger.Erro("Erro em ConnectToServerTCP - ponto de recarga")
		return
	}
	defer conexao.Close()
	//envia identificacao inicial

	logger.Info("Ponto de Recarga conectado")
	IdentificacaoInicial(logger, conexao)

	//recebe solicitacoes do servidor
	go processarFila(logger, conexao)

	// No loop principal, garantir que o processamento de mensagens não bloqueie:

	for {
		msg, erro := dataJson.ReceiveMessage(conexao)
		if erro != nil {
			logger.Erro(fmt.Sprintf("Erro ao ler mensagem do servidor - %v", erro))
			return
		}

		// Processar cada tipo de mensagem em uma goroutine separada
		// para não bloquear o loop principal
		go func(mensagem dataJson.Mensagem) {
			switch mensagem.Tipo {
			case "get-disponibilidade":
				// Código modificado para responder rapidamente
				mutex.Lock()
				status := "Situacao atual: "
				if len(fila) == 0 {
					status += "sem fila"
				} else {
					status += fmt.Sprintf("com %d na fila", len(fila))
				}
				mutex.Unlock()

				respMsg := dataJson.Mensagem{
					Tipo:     "disponibilidade",
					Conteudo: status,
					Origem:   "ponto-de-recarga",
				}

				dataJson.SendMessage(conexao, respMsg)

			case "nova-solicitacao":
				mutex.Lock()
				veiculoID := mensagem.Conteudo // Agora é a placa do veículo
				posicaoFila := len(fila) + 1
				fila = append(fila, veiculoID)
				logger.Info(fmt.Sprintf("Veículo %s adicionado à fila. Posição: %d", veiculoID, posicaoFila))

				// Enviar status da fila ao servidor
				statusMsg := dataJson.Mensagem{
					Tipo:     "status-fila",
					Conteudo: fmt.Sprintf("%d", posicaoFila), // Posição na fila (1 = próximo, >1 = esperar)
					Origem:   "ponto-de-recarga",
				}
				mutex.Unlock()

				// Enviar o status da fila para o servidor
				dataJson.SendMessage(conexao, statusMsg)
			case "veiculo-chegou":
				placaVeiculo := mensagem.Conteudo
				logger.Info(fmt.Sprintf("Servidor informou chegada do veículo: %s", placaVeiculo))

				mutex.Lock()
				// Verificar se este veículo está em nossa fila
				encontrado := false
				for i, id := range fila {
					if id == placaVeiculo {
						encontrado = true

						// Processar apenas se for o primeiro da fila
						if i == 0 {
							// Verificar se o veículo está no mapa de espera
							if ch, ok := veiculosEmEspera[placaVeiculo]; ok {
								mutex.Unlock()
								close(ch) // Sinalizar que chegou
								logger.Info(fmt.Sprintf("Veículo %s informou chegada", placaVeiculo))
							} else {
								mutex.Unlock()
								logger.Erro(fmt.Sprintf("Veículo %s está na fila mas não tem canal de espera", placaVeiculo))
							}
						} else {
							mutex.Unlock()
							logger.Erro(fmt.Sprintf("Veículo %s informou chegada, mas não é o primeiro da fila (posição %d)", placaVeiculo, i+1))
						}
						break
					}
				}

				if !encontrado {
					mutex.Unlock()
					logger.Erro(fmt.Sprintf("Veículo %s informou chegada, mas não está na fila", placaVeiculo))
				}
			case "liberar-ponto":
				// Sinalizar imediatamente para o loop de processamento
				select {
				case proximoVeiculoSignal <- struct{}{}:
					logger.Info("Sinal para processar próximo veículo enviado")
				default:
					// Canal já tem um sinal, então não precisa enviar outro
				}
			default:
				logger.Info(fmt.Sprintf("Mensagem recebida do servidor: %s", mensagem.Conteudo))
			}
		}(msg)
	}
}
