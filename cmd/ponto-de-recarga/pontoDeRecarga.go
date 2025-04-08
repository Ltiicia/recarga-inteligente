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

var mutex sync.Mutex
var veiculosEmEspera map[string]chan bool
var filaAtual []string
var proximoVeiculoSignal = make(chan struct{}, 1)

func processarFila(logger *logger.Logger, conexao net.Conn) {
	for {
		mutex.Lock()
		if len(filaAtual) == 0 {
			mutex.Unlock()
			select {
			case <-proximoVeiculoSignal:
				logger.Info("Sinal recebido para processar próximo veículo")
				continue
			case <-time.After(1 * time.Second):
				continue
			}
		}

		veiculoAtual := filaAtual[0]
		logger.Info(fmt.Sprintf("Processando veículo na fila: %s", veiculoAtual))
		mutex.Unlock()

		msg := dataJson.Mensagem{
			Tipo:     "chamando-veiculo",
			Conteudo: veiculoAtual,
			Origem:   "ponto-de-recarga",
		}

		if err := dataJson.SendMessage(conexao, msg); err != nil {
			logger.Erro(fmt.Sprintf("Erro ao enviar chamada de veículo: %v", err))
			time.Sleep(2 * time.Second)
			continue
		}

		chegou := make(chan bool, 1)
		mutex.Lock()
		veiculosEmEspera[veiculoAtual] = chegou
		mutex.Unlock()

		logger.Info(fmt.Sprintf("Aguardando chegada de %s", veiculoAtual))

		// Aguardar com timeout a chegada do veículo
		timeout := 60 * time.Second // Um minuto para o veículo chegar
		select {
		case <-chegou:
			logger.Info(fmt.Sprintf("Veículo %s informou chegada, iniciando carregamento", veiculoAtual))
			// Continuar com o carregamento
		case <-time.After(timeout):
			logger.Erro(fmt.Sprintf("Timeout aguardando veículo %s, removendo da fila", veiculoAtual))

			// Remover da fila e do mapa de espera
			mutex.Lock()
			if len(filaAtual) > 0 && filaAtual[0] == veiculoAtual {
				filaAtual = filaAtual[1:] // Remover da fila
			}
			delete(veiculosEmEspera, veiculoAtual) // Remover do mapa
			mutex.Unlock()
			continue // Processar próximo veículo
		}

		// Simulação do carregamento - só chega aqui se o veículo chegou
		logger.Info(fmt.Sprintf("Iniciando carregamento para: %s", veiculoAtual))
		time.Sleep(20 * time.Second) // Tempo de carregamento simulado

		tempoHoras := 1.0 + rand.Float64()      // Entre 1 e 2 horas
		taxaKwh := 0.80                         // Taxa por kWh
		potenciaKw := 22.0                      // Potência em kW
		consumoTotal := potenciaKw * tempoHoras // Consumo total em kWh
		valor := consumoTotal * taxaKwh         // Valor total

		logger.Info(fmt.Sprintf("Recarga finalizada para: %s - Consumo: %.2f kWh, Valor: R$ %.2f",
			veiculoAtual, consumoTotal, valor))

		// Remover o veículo da fila e do mapa de espera
		mutex.Lock()
		if len(filaAtual) > 0 && filaAtual[0] == veiculoAtual {
			filaAtual = filaAtual[1:] // Remover o primeiro elemento
		}
		delete(veiculosEmEspera, veiculoAtual)
		mutex.Unlock()

		// Notificar o servidor que a recarga foi concluída
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

func enviarDisponibilidade(logger *logger.Logger, conexao net.Conn) {
	mutex.Lock()
	defer mutex.Unlock()

	status := "Situacao atual: "
	if len(filaAtual) == 0 {
		status += "sem fila"
	} else {
		status += fmt.Sprintf("com %d na fila", len(filaAtual))
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

func IdentificacaoInicial(logger *logger.Logger, conexao net.Conn) {
	if err := tcpIP.SendIdentification(conexao, "ponto-de-recarga"); err != nil {
		logger.Erro(fmt.Sprintf("Erro ao enviar identificação: %v", err))
	}
}


func main() {
	veiculosEmEspera = make(map[string]chan bool)
	logger := logger.NewLogger(os.Stdout)

	conexao, err := tcpIP.ConnectToServerTCP("servidor:5000")
	if err != nil {
		logger.Erro("Erro ao conectar com o servidor")
		return
	}
	defer conexao.Close()

	logger.Info("Ponto de Recarga conectado")
	IdentificacaoInicial(logger, conexao)

	go processarFila(logger, conexao)

	for {
		msg, erro := dataJson.ReceiveMessage(conexao)
		if erro != nil {
			logger.Erro(fmt.Sprintf("Erro ao ler mensagem do servidor - %v", erro))
			return
		}
		// Processar cada tipo de mensagem em uma goroutine separada para não bloquear o loop principal
		go func(mensagem dataJson.Mensagem) {
			switch mensagem.Tipo {
			case "fila-atualizada":
				mutex.Lock()
				filaAtual = dataJson.ParseFila(mensagem.Conteudo)
				logger.Info("Fila atualizada")
				mutex.Unlock()
			case "nova-solicitacao":
				mutex.Lock()
				veiculoID := mensagem.Conteudo
				posicaoFila := len(filaAtual) + 1
				filaAtual = append(filaAtual, veiculoID)
				logger.Info(fmt.Sprintf("Veículo %s adicionado à fila.", veiculoID))

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
				for i, id := range filaAtual {
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
			case "get-disponibilidade":
				enviarDisponibilidade(logger, conexao)
				logger.Info("Disponibilidade atual enviada ao servidor")
			default:
			}
		}(msg)
	}
}
