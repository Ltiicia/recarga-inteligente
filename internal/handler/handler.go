package handler

import (
	"encoding/json"
	"fmt"
	"math"
	"net"
	"recarga-inteligente/internal/dataJson"
	"recarga-inteligente/internal/distancia"
	"recarga-inteligente/internal/logger"
	"recarga-inteligente/internal/store"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type PontoRanking struct {
	ID        int
	Distancia float64
	Fila      int
	Score     float64
}

var (
	reservasAtivas = make(map[string]int) // mapa de placa -> pontoID
	reservasMutex  sync.Mutex
)

// ok
func HandleConnection(conexao net.Conn, connectionStore *store.ConnectionStore, logger *logger.Logger) {
	defer connectionStore.RemoveConnection(conexao)
	on := true
	for on {
		//recebe mensagem inicial
		mensagemRecebida, erro := dataJson.ReceiveMessage(conexao)
		if erro != nil {
			e := fmt.Sprintf("%v", erro)
			if e == "erro: EOF" {
				logger.Info(fmt.Sprintf("conexao (%s) desconectada", conexao.RemoteAddr()))
			} else {
				logger.Erro(fmt.Sprintf("Erro ao ler mensagem inicial: %v", erro))
			}
			on = false
			continue
		}

		// Processa cada mensagem em uma goroutine separada para não bloquear o loop principal
		go func(mensagem dataJson.Mensagem, conn net.Conn) {
			switch mensagem.Origem {
			case "ponto-de-recarga":
				handlePontoDeRecarga(logger, connectionStore, conn, mensagem)
			case "veiculo":
				handleVeiculo(logger, connectionStore, conn, mensagem)
			default:
				logger.Info("Origem desconhecida, ignorando mensagem")
			}
		}(mensagemRecebida, conexao)
	}
}

func handlePontoDeRecarga(logger *logger.Logger, connectionStore *store.ConnectionStore, conexao net.Conn, mensagem dataJson.Mensagem) {

	if mensagem.Tipo == "identificacao" {
		idPonto := connectionStore.AddPontoRecarga(conexao)
		if idPonto == -1 {
			logger.Erro(fmt.Sprintf("Ponto de recarga nao cadastrado tentando se conectar -> desconectado: %s", conexao.RemoteAddr()))
			connectionStore.RemoveConnection(conexao)
			return
		}
		logger.Info(fmt.Sprintf("Novo ponto de recarga conectado id: (%d)", idPonto))

		// Solicitar disponibilidade inicial
		disponibilidade := disponibilidadePonto(logger, connectionStore, idPonto)
		logger.Info(fmt.Sprintf("Disponibilidade inicial do Ponto id (%d) recebida: %s", idPonto, disponibilidade.Conteudo))
		return
	}

	id := connectionStore.GetIdPonto(conexao)
	switch mensagem.Tipo {
	case "chamando-veiculo":
		// O ponto está chamando um veículo para atendimento
		placaVeiculo := mensagem.Conteudo
		logger.Info(fmt.Sprintf("Ponto ID %d está chamando o veículo %s", id, placaVeiculo))

		// Localizar a conexão do veículo
		veiculoCon := connectionStore.GetConexaoPorPlaca(placaVeiculo)
		if veiculoCon != nil {
			// Criar uma goroutine para não bloquear o processamento do ponto
			go func() {
				msgSuaVez := dataJson.Mensagem{
					Tipo:     "sua-vez",
					Conteudo: fmt.Sprintf("É sua vez de ser atendido no ponto ID %d!", id),
					Origem:   "servidor",
				}
				erro := dataJson.SendMessage(veiculoCon, msgSuaVez)
				if erro != nil {
					logger.Erro(fmt.Sprintf("Erro ao notificar veículo %s que é sua vez: %v", placaVeiculo, erro))
				}
			}()
		}

	case "recarga-finalizada":
		// Extrair informações da recarga
		var placaVeiculo string
		var consumoTotal, valor float64
		var pontoID int

		pontoID = id

		n, err := fmt.Sscanf(mensagem.Conteudo, "Veiculo %s atendido. Consumo: %f kWh, Valor: R$ %f",
			&placaVeiculo, &consumoTotal, &valor)

		if err != nil || n != 3 {
			logger.Erro(fmt.Sprintf("Erro ao extrair informacoes da recarga: %v (itens extraídos: %d)", err, n))

			// Método alternativo de extração como fallback
			if strings.Contains(mensagem.Conteudo, "Veículo") {
				parts := strings.Split(mensagem.Conteudo, "Veículo ")
				if len(parts) > 1 {
					placaParts := strings.Split(parts[1], " ")
					placaVeiculo = placaParts[0]

					// Tentar extrair valores novamente
					valorParts := strings.Split(mensagem.Conteudo, "Valor: R$ ")
					if len(valorParts) > 1 {
						valor, _ = strconv.ParseFloat(strings.TrimSpace(valorParts[1]), 64)
					}

					consumoParts := strings.Split(mensagem.Conteudo, "Consumo: ")
					if len(consumoParts) > 1 {
						consumoStr := strings.Split(consumoParts[1], " kWh")[0]
						consumoTotal, _ = strconv.ParseFloat(strings.TrimSpace(consumoStr), 64)
					}
				}
			}
		}

		logger.Info(fmt.Sprintf("Recarga finalizada pelo ponto ID %d para veículo %s: Consumo: %.2f kWh, Valor: R$ %.2f",
			pontoID, placaVeiculo, consumoTotal, valor))

		// 1. Remover do mapa de reservas ativas primeiro para liberar o ponto
		reservasMutex.Lock()
		delete(reservasAtivas, placaVeiculo)
		reservasMutex.Unlock()

		// 2. Notificar o ponto que pode processar o próximo veículo imediatamente
		msgLiberarPonto := dataJson.Mensagem{
			Tipo:     "liberar-ponto",
			Conteudo: "Ponto liberado para atender próximo veículo",
			Origem:   "servidor",
		}

		// Enviamos em uma goroutine para não bloquear
		go func() {
			err := dataJson.SendMessage(conexao, msgLiberarPonto)
			if err != nil {
				logger.Erro(fmt.Sprintf("Erro ao notificar ponto sobre liberação: %v", err))
			}
		}()

		// 3. Registrar a recarga no histórico
		erro := dataJson.RegistrarRecarga(placaVeiculo, pontoID, valor)
		if erro != nil {
			logger.Erro(fmt.Sprintf("Erro ao registrar recarga: %v", erro))
		} else {
			logger.Info(fmt.Sprintf("Recarga de R$ %.2f registrada com sucesso para o veículo %s",
				valor, placaVeiculo))
		}

		// 4. Notificar o veículo - isso deve estar em uma goroutine e usar o tipo correto
		veiculoCon := connectionStore.GetConexaoPorPlaca(placaVeiculo)
		if veiculoCon != nil {
			// Usar uma goroutine para não bloquear
			go func() {
				msgVeiculo := dataJson.Mensagem{
					Tipo:     "recarga-finalizada",
					Conteudo: mensagem.Conteudo,
					Origem:   "servidor",
				}

				erro := dataJson.SendMessage(veiculoCon, msgVeiculo)
				if erro != nil {
					logger.Erro(fmt.Sprintf("Erro ao notificar veículo %s sobre recarga finalizada: %v",
						placaVeiculo, erro))
				} else {
					logger.Info(fmt.Sprintf("Veículo %s notificado sobre recarga finalizada", placaVeiculo))
				}
			}()
		} else {
			logger.Erro(fmt.Sprintf("Conexão do veículo %s não encontrada para notificar sobre recarga finalizada", placaVeiculo))
		}
	}
}

// ok
func disponibilidadePonto(logger *logger.Logger, connectionStore *store.ConnectionStore, pontoId int) dataJson.Mensagem {
	fila := connectionStore.GetFilaPorPonto(pontoId)
	conexaoPonto := connectionStore.GetConexaoPorID(pontoId)
	filaJSON, err := json.Marshal(fila)
	if err != nil {
		logger.Erro(fmt.Sprintf("Erro ao serializar fila do ponto %d: %v", pontoId, err))
	} else {
		msgFila := dataJson.Mensagem{
			Tipo:     "atualizar-fila",
			Conteudo: string(filaJSON),
			Origem:   "servidor",
		}
		err = dataJson.SendMessage(conexaoPonto, msgFila)
		if err != nil {
			logger.Erro(fmt.Sprintf("Erro ao enviar fila para ponto %d: %v", pontoId, err))
		} else {
			logger.Info(fmt.Sprintf("Fila Recebida do ponto %d com sucesso", pontoId))
		}
	}

	solicitacao := dataJson.Mensagem{
		Tipo:     "get-disponibilidade",
		Conteudo: fmt.Sprintf("Ola ponto de recarga id (%d)! Informe sua disponibilidade / fila", pontoId),
		Origem:   "servidor",
	}
	erro := dataJson.SendMessage(conexaoPonto, solicitacao)
	if erro != nil {
		logger.Erro(fmt.Sprintf("Erro ao solicitar disponibilidade ao ponto-de-recarga id (%d): %v", pontoId, erro))
		return dataJson.Mensagem{}
	}

	disponibilidade := connectionStore.GetFilaPorPonto(pontoId)

	status := "Situacao atual: "
	if len(disponibilidade) == 0 {
		status += "sem fila"
	} else {
		status += fmt.Sprintf("com %d na fila", len(disponibilidade))
	}

	msg := dataJson.Mensagem{
		Tipo:     "disponibilidade",
		Conteudo: status,
		Origem:   "ponto-de-recarga",
	}

	return msg
}

// ok
func processarLocalizacao(logger *logger.Logger, connectionStore *store.ConnectionStore, conexao net.Conn, mensagem dataJson.Mensagem) {
	var latitude, longitude float64
	_, erro := fmt.Sscanf(mensagem.Conteudo, "%f,%f", &latitude, &longitude)
	if erro != nil {
		logger.Erro(fmt.Sprintf("Erro ao receber localizacao: %v", erro))
		return
	}
	logger.Info(fmt.Sprintf("Localizacao recebida: Latitude %f, Longitude %f", latitude, longitude))

	logger.Info("Calculando ranking dos pontos de recarga...")

	// Calcular ranking
	ranking := calcularRankingPontos(logger, latitude, longitude, connectionStore)

	for i, ponto := range ranking {
		logger.Info(fmt.Sprintf("Ranking[%d]: ID=%d, Distância=%.2f, Fila=%d, Score=%.2f",
			i, ponto.ID, ponto.Distancia, ponto.Fila, ponto.Score))
	}

	// Enviar ranking ao veículo
	logger.Info("Enviando ranking ao veículo...")

	var rankingStr string
	for i, ponto := range ranking {
		// Modificar a exibição da fila para não mostrar '999' ao usuário
		filaExibicao := ponto.Fila
		if filaExibicao == 999 {
			filaExibicao = 0 // Mostrar como 0 para o usuário quando não temos informação
		}

		rankingStr += fmt.Sprintf("%d. Ponto ID: %d, Distância: %.2f km, Fila: %d veículos\n",
			i+1, ponto.ID, ponto.Distancia, filaExibicao)
	}

	msg := dataJson.Mensagem{
		Tipo:     "ranking-pontos",
		Conteudo: rankingStr,
		Origem:   "servidor",
	}

	// Tentar enviar a mensagem com retry
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		erro = dataJson.SendMessage(conexao, msg)
		if erro == nil {
			logger.Info("Ranking enviado com sucesso ao veículo")
			return
		}

		logger.Erro(fmt.Sprintf("Tentativa %d: Erro ao enviar ranking ao veículo: %v", i+1, erro))
		time.Sleep(500 * time.Millisecond) // Aguardar antes de tentar novamente
	}

	logger.Erro("Falha ao enviar ranking após várias tentativas")
}

//

func processarReserva(logger *logger.Logger, connectionStore *store.ConnectionStore, conexao net.Conn, mensagem dataJson.Mensagem) {
	pontoID, _ := strconv.Atoi(mensagem.Conteudo)
	logger.Info(fmt.Sprintf("Reserva solicitada para ponto ID %d", pontoID))

	// Obter placa do veículo
	placa := connectionStore.GetVeiculoPlaca(conexao)

	// Encontrar a conexão do ponto pelo ID
	pontoCon := connectionStore.GetConexaoPorID(pontoID)
	if pontoCon == nil {
		logger.Erro(fmt.Sprintf("Ponto ID %d não encontrado", pontoID))
		// Informar ao veículo que a reserva falhou
		msg := dataJson.Mensagem{
			Tipo:     "reserva-falhou",
			Conteudo: fmt.Sprintf("Ponto ID %d não encontrado", pontoID),
			Origem:   "servidor",
		}
		dataJson.SendMessage(conexao, msg)
		return
	}

	//filaAtual := connectionStore.GetFilaPorPonto(pontoID)
	//logger.Info(fmt.Sprintf("Verificação em tempo real: Ponto ID %d tem %d veículos na fila", pontoID, len(filaAtual)+1))

	// Enviar solicitação para o ponto
	msgPonto := dataJson.Mensagem{
		Tipo:     "nova-solicitacao",
		Conteudo: placa,
		Origem:   "servidor",
	}

	erro := dataJson.SendMessage(pontoCon, msgPonto)
	if erro != nil {
		logger.Erro(fmt.Sprintf("Erro ao enviar solicitação ao ponto: %v", erro))

		// Notificar o veículo sobre a falha mesmo em caso de erro
		msgFalha := dataJson.Mensagem{
			Tipo:     "reserva-falhou",
			Conteudo: fmt.Sprintf("Falha ao comunicar com o ponto ID %d", pontoID),
			Origem:   "servidor",
		}
		dataJson.SendMessage(conexao, msgFalha)
		return
	}

	// Registrar a reserva temporariamente em memória
	reservasMutex.Lock()
	reservasAtivas[placa] = pontoID
	reservasMutex.Unlock()

	// Consulta a fila diretamente no servidor
	fila := connectionStore.GetFilaPorPonto(pontoID)
	posicaoFila := len(fila) + 1 // Posição padrão se o veículo ainda não estiver na fila

	// Verifica a posição real do veículo (placa)
	for i, v := range fila {
		if v.Placa == placa {
			posicaoFila = i + 1
			break
		}
	}

	var mensagemStatus string
	// Sempre enviar uma mensagem ao veículo, independente do resultado
	if posicaoFila <= 1 {
		mensagemStatus = fmt.Sprintf("Reserva confirmada para ponto ID %d. Você é o próximo a ser atendido!", pontoID)
	} else {
		mensagemStatus = fmt.Sprintf("Reserva confirmada para ponto ID %d. Você está na posição %d da fila, aguarde sua vez.", pontoID, posicaoFila)
	}

	msgConfirmacao := dataJson.Mensagem{
		Tipo:     "reserva-confirmada",
		Conteudo: mensagemStatus,
		Origem:   "servidor",
	}

	// Tentar enviar várias vezes se necessário
	maxTentativas := 3
	for i := 0; i < maxTentativas; i++ {
		err := dataJson.SendMessage(conexao, msgConfirmacao)
		if err == nil {
			break
		}
		logger.Erro(fmt.Sprintf("Tentativa %d: Erro ao enviar confirmação ao veículo: %v", i+1, err))
		time.Sleep(100 * time.Millisecond)
	}

	go monitorarFilaParaVeiculo(logger, connectionStore, conexao, placa, pontoID)
}

func verificarFilaPontoEspecifico(logger *logger.Logger, connectionStore *store.ConnectionStore, pontoID int) int {
	resp := disponibilidadePonto(logger, connectionStore, pontoID)
	//logger.Info(fmt.Sprintf("disponibilidade recebida em verificarFilaPontoEspecifico: %s", resp.Conteudo))

	// Extrair o tamanho da fila da resposta
	var tamanhoFila int = 0
	if resp.Tipo == "disponibilidade" {
		if strings.Contains(resp.Conteudo, "sem fila") {
			tamanhoFila = 0
		} else {
			//fmt.Sscanf(resp.Conteudo, "Situacao atual: com %d na fila", &tamanhoFila)
		}
	}

	return tamanhoFila
}

func monitorarFilaParaVeiculo(logger *logger.Logger, connectionStore *store.ConnectionStore, veiculoCon net.Conn, placa string, pontoID int) {
	// Monitorar por no máximo 10 minutos
	timeout := time.After(10 * time.Minute)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Verificar se o veículo ainda está na fila e em qual posição
	for {
		select {
		case <-ticker.C:
			// Verificar se a reserva ainda existe
			reservasMutex.Lock()
			pontoReservado, existe := reservasAtivas[placa]
			reservasMutex.Unlock()

			if !existe || pontoReservado != pontoID {
				logger.Info(fmt.Sprintf("Monitoramento finalizado para veículo %s - não possui mais reserva ativa", placa))
				return
			}

			// Verificar a fila atual
			tamanhoFila := verificarFilaPontoEspecifico(logger, connectionStore, pontoID)

			// Calcular posição estimada do veículo
			resp := disponibilidadePonto(logger, connectionStore, pontoID)
			//logger.Info(fmt.Sprintf("disponibilidade recebida em monitorarFilaParaVeiculo: %s", resp.Conteudo))

			if resp.Tipo == "disponibilidade" {
				// Enviar atualização ao veículo
				msgAtualizar := dataJson.Mensagem{
					Tipo: "posicao-fila",
					Conteudo: fmt.Sprintf("Atualização: Você está na fila do ponto ID %d. Existem %d veículos na fila.",
						pontoID, tamanhoFila),
					Origem: "servidor",
				}

				// Não interromper o monitoramento se falhar ao enviar uma atualização
				err := dataJson.SendMessage(veiculoCon, msgAtualizar)
				if err != nil {
					logger.Erro(fmt.Sprintf("Erro ao enviar atualização da fila para veículo %s: %v", placa, err))
				}
			}

		case <-timeout:
			logger.Info(fmt.Sprintf("Monitoramento de fila expirado para veículo %s", placa))
			return
		}
	}
}

// ok
func consultarDisponibilidadePontos(logger *logger.Logger, connectionStore *store.ConnectionStore) map[int]int {
	filas := make(map[int]int)
	var mutex sync.Mutex // Para proteger o mapa de filas durante acessos concorrentes

	// Criar um WaitGroup para esperar todas as goroutines terminarem
	var wg sync.WaitGroup

	// Canal para timeout global da operação
	timeout := time.After(5 * time.Second)

	// Canal para receber os resultados das goroutines
	resultados := make(chan struct {
		id          int
		tamanhoFila int
	}, connectionStore.GetTotalPontosConectados())

	// Para cada ponto de recarga conectado, consulta sua disponibilidade em uma goroutine separada
	pontosMap := connectionStore.GetPontosMap()
	for conexao, id := range pontosMap {
		wg.Add(1)
		go func(conexao net.Conn, id int) {
			defer wg.Done()

			// Criar um canal para timeout individual da consulta
			consultaTimeout := time.After(2 * time.Second)

			// Canal para receber a resposta da consulta
			respChan := make(chan dataJson.Mensagem, 1)

			// Fazer a consulta em uma goroutine
			go func() {
				resp := disponibilidadePonto(logger, connectionStore, id)
				//logger.Info(fmt.Sprintf("disponibilidade recebida em consultarDisponibilidadePontos: %s", resp.Conteudo))

				if resp.Tipo == "disponibilidade" {
					respChan <- resp
				}
			}()

			// Esperar pela resposta ou timeout
			select {
			case resp := <-respChan:
				// Extrair o número de veículos na fila da resposta
				var tamanhoFila int
				if strings.Contains(resp.Conteudo, "sem fila") {
					tamanhoFila = 0
				} else {
					fmt.Sscanf(resp.Conteudo, "Situacao atual: com %d na fila", &tamanhoFila)
				}

				// Enviar o resultado para o canal principal
				resultados <- struct {
					id          int
					tamanhoFila int
				}{id, tamanhoFila}

			case <-consultaTimeout:
				logger.Erro(fmt.Sprintf("Timeout ao consultar disponibilidade do ponto ID %d", id))
			}
		}(conexao, id)
	}

	// Goroutine para fechar o canal de resultados quando todas as consultas terminarem
	go func() {
		wg.Wait()
		close(resultados)
	}()

	// Coletar os resultados ou encerrar por timeout
	done := false
	for !done {
		select {
		case resultado, ok := <-resultados:
			if !ok {
				done = true // Canal fechado, todas as consultas terminaram
				break
			}
			mutex.Lock()
			filas[resultado.id] = resultado.tamanhoFila
			mutex.Unlock()

		case <-timeout:
			logger.Erro("Timeout global ao consultar disponibilidade dos pontos")
			done = true
		}
	}

	return filas
}

// ok
func calcDistancia(latVeiculo float64, lonVeiculo float64, totalPontos int) (map[int]float64, error) {
	mapDistancias := make(map[int]float64)

	for id := 1; id <= totalPontos; id++ {
		ponto, erro := dataJson.GetPontoId(id)
		if erro == 0 {
			d := distancia.GetDistancia(latVeiculo, lonVeiculo, ponto.Latitude, ponto.Longitude)
			km := d / 1000
			mapDistancias[id] = km
		} else if erro == 2 {
			return mapDistancias, fmt.Errorf("ponto id (%d) nao localizado", id)
		} else {
			return mapDistancias, fmt.Errorf("Erro ao carregar arquivo json")
		}
	}
	return mapDistancias, nil
}

// ok
func calcularRankingPontos(logger *logger.Logger, latVeiculo, lonVeiculo float64, connectionStore *store.ConnectionStore) []PontoRanking {
	// Calcular distâncias
	mapDistancias, _ := calcDistancia(latVeiculo, lonVeiculo, connectionStore.GetTotalPontosConectados())

	// Consultar tamanho das filas em tempo real
	mapFilas := consultarDisponibilidadePontos(logger, connectionStore)

	// Criar lista de pontos com seus scores
	var pontos []PontoRanking
	for id, distancia := range mapDistancias {
		tamanhoFila, ok := mapFilas[id]
		if !ok {
			tamanhoFila = 999 // Valor alto para pontos sem informação de fila
		}

		// Calcular score (menor é melhor)
		pesoFila := 0.6
		pesoDistancia := 0.4

		// Score baseado na distância (valores entre 0-10)
		scoreDistancia := math.Min(distancia, 10.0) * pesoDistancia

		// Score baseado na fila (com penalidade para filas grandes)
		var scoreFila float64
		if tamanhoFila <= 3 {
			scoreFila = float64(tamanhoFila) * pesoFila
		} else {
			// Penalidade para filas maiores que 3
			scoreFila = (3.0 + math.Pow(float64(tamanhoFila-3), 1.5)) * pesoFila
		}

		score := scoreDistancia + scoreFila

		pontos = append(pontos, PontoRanking{
			ID:        id,
			Distancia: distancia,
			Fila:      tamanhoFila,
			Score:     score,
		})
	}

	// Ordenar por score (menor é melhor)
	sort.Slice(pontos, func(i, j int) bool {
		return pontos[i].Score < pontos[j].Score
	})

	// Retornar até 3 melhores opções
	if len(pontos) > 3 {
		return pontos[:3]
	}
	return pontos
}

// ok
func handleVeiculo(logger *logger.Logger, connectionStore *store.ConnectionStore, conexao net.Conn, mensagem dataJson.Mensagem) {

	switch mensagem.Tipo {
	case "identificacao":
		var placa string
		if strings.Contains(mensagem.Conteudo, "placa") {
			parts := strings.Split(mensagem.Conteudo, "placa ")
			if len(parts) > 1 {
				placa = strings.Split(parts[1], " ")[0]
				logger.Info(fmt.Sprintf("Novo veículo placa %s conectado: (%s)", placa, conexao.RemoteAddr()))

				if connectionStore.PlacaJaEmUso(placa, conexao) {
					msgErro := dataJson.Mensagem{
						Tipo:     "placa-em-uso",
						Conteudo: "Esta placa já está em uso por outro veículo.",
						Origem:   "servidor",
					}
					dataJson.SendMessage(conexao, msgErro)
					return
				}

				// Armazenar a placa do veículo
				connectionStore.AddVeiculo(conexao, placa)

				// Salvar a placa no JSON de veículos
				erro := dataJson.SalvarVeiculo(placa)
				if erro != nil {
					logger.Erro(fmt.Sprintf("Erro ao salvar dados do veículo: %v", erro))
				}
			} else {
				logger.Info(fmt.Sprintf("Novo veículo conectado: (%s)", conexao.RemoteAddr()))
				connectionStore.AddVeiculo(conexao, "")
			}
		} else {
			logger.Info(fmt.Sprintf("Novo veículo conectado: (%s)", conexao.RemoteAddr()))
			connectionStore.AddVeiculo(conexao, "")
		}

	case "get-recarga":
		go processarSolicitacaoRecarga(logger, conexao)

	case "localizacao":
		go processarLocalizacao(logger, connectionStore, conexao, mensagem)

	case "solicitar-reserva":
		go processarReserva(logger, connectionStore, conexao, mensagem)

	case "veiculo-chegou":
		// Veículo informou que chegou ao ponto de recarga
		placaVeiculo := mensagem.Conteudo
		logger.Info(fmt.Sprintf("Veículo %s informou chegada ao ponto", placaVeiculo))

		// Obter o ID do ponto do mapa de reservas ativas
		reservasMutex.Lock()
		pontoID, existe := reservasAtivas[placaVeiculo]
		reservasMutex.Unlock()

		if !existe {
			// Fallback: tentar buscar nos registros permanentes
			_, err := dataJson.ObterUltimoReserva(placaVeiculo)
			if err != nil {
				logger.Erro(fmt.Sprintf("Não foi possível determinar o ponto reservado para o veículo %s: %v", placaVeiculo, err))
				return
			}
		}

		// Encaminhar a mensagem APENAS para o ponto específico
		pontoCon := connectionStore.GetConexaoPorID(pontoID)
		if pontoCon == nil {
			logger.Erro(fmt.Sprintf("Ponto ID %d não encontrado para notificar sobre chegada do veículo %s", pontoID, placaVeiculo))
			return
		}
		logger.Info(fmt.Sprintf("ponto %d conexao recebida %s", pontoID, pontoCon.RemoteAddr()))

		msgPonto := dataJson.Mensagem{
			Tipo:     "veiculo-chegou",
			Conteudo: placaVeiculo,
			Origem:   "servidor",
		}

		erro := dataJson.SendMessage(pontoCon, msgPonto)
		if erro != nil {
			logger.Erro(fmt.Sprintf("Erro ao notificar ponto %d sobre chegada do veículo: %v", pontoID, erro))
		} else {
			logger.Info(fmt.Sprintf("Ponto ID %d notificado sobre a chegada do veículo %s", pontoID, placaVeiculo))
		}

	case "verificar-placa":
		placa := mensagem.Conteudo
		logger.Info(fmt.Sprintf("Verificando disponibilidade da placa: %s", placa))

		// Verificar se a placa já está em uso em alguma conexão ativa
		placaEmUso := false

		// Verificar nas conexões ativas
		for _, placaAtiva := range connectionStore.GetTodasPlacasAtivas() {
			if placaAtiva == placa {
				placaEmUso = true
				break
			}
		}

		// Enviar resposta
		var msgResposta dataJson.Mensagem
		if placaEmUso {
			msgResposta = dataJson.Mensagem{
				Tipo:     "placa-indisponivel",
				Conteudo: "Esta placa já está em uso.",
				Origem:   "servidor",
			}
		} else {
			msgResposta = dataJson.Mensagem{
				Tipo:     "placa-disponivel",
				Conteudo: "Placa disponível para uso.",
				Origem:   "servidor",
			}
		}

		dataJson.SendMessage(conexao, msgResposta)

	case "consultar-historico":
		// Veículo está solicitando seu histórico de recargas
		placa := connectionStore.GetVeiculoPlaca(conexao)
		logger.Info(fmt.Sprintf("Veículo %s solicitou histórico de recargas", placa))

		// Buscar histórico no arquivo JSON
		recargas, erro := dataJson.ObterHistoricoRecargas(placa)
		if erro != nil {
			logger.Erro(fmt.Sprintf("Erro ao obter histórico de recargas para %s: %v", placa, erro))

			// Enviar mensagem de erro
			msgErro := dataJson.Mensagem{
				Tipo:     "historico-erro",
				Conteudo: "Erro ao buscar histórico de recargas",
				Origem:   "servidor",
			}
			dataJson.SendMessage(conexao, msgErro)
			return
		}

		// Serializar o histórico para JSON
		historicoJSON, erro := json.Marshal(recargas)
		if erro != nil {
			logger.Erro(fmt.Sprintf("Erro ao serializar histórico: %v", erro))
			return
		}

		// Enviar o histórico ao veículo
		msgHistorico := dataJson.Mensagem{
			Tipo:     "historico-recargas",
			Conteudo: string(historicoJSON),
			Origem:   "servidor",
		}

		erro = dataJson.SendMessage(conexao, msgHistorico)
		if erro != nil {
			logger.Erro(fmt.Sprintf("Erro ao enviar histórico para veículo %s: %v", placa, erro))
		} else {
			logger.Info(fmt.Sprintf("Histórico enviado com sucesso para veículo %s", placa))
		}

	case "limpar-historico":
		placa := connectionStore.GetVeiculoPlaca(conexao)
		logger.Info(fmt.Sprintf("Veículo %s solicitou limpeza do histórico de recargas", placa))

		err := dataJson.LimparHistoricoRecargas(placa)
		if err != nil {
			logger.Erro(fmt.Sprintf("Erro ao limpar histórico de %s: %v", placa, err))
			dataJson.SendMessage(conexao, dataJson.Mensagem{
				Tipo:     "erro-pagamento",
				Conteudo: "Erro ao processar pagamento.",
				Origem:   "servidor",
			})
			return
		}

		dataJson.SendMessage(conexao, dataJson.Mensagem{
			Tipo:     "pagamento-confirmado",
			Conteudo: "Histórico de recargas limpo com sucesso.",
			Origem:   "servidor",
		})

	default:
		logger.Erro(fmt.Sprintf("Tipo de solicitacao ainda nao foi mapeada - %s", mensagem.Tipo))
	}
}

// ok
func processarSolicitacaoRecarga(logger *logger.Logger, conexao net.Conn) {
	solicitacao := dataJson.Mensagem{
		Tipo:     "get-localizacao",
		Conteudo: "Ola Veiculo! Informe sua localizacao atual.",
		Origem:   "servidor",
	}

	// Enviar solicitação de localização
	erro := dataJson.SendMessage(conexao, solicitacao)
	if erro != nil {
		logger.Erro(fmt.Sprintf("Erro ao solicitar localizacao ao veiculo: %v", erro))
		return
	}

	// Enviar dados da região
	erro = dataJson.SendDadosRegiao(conexao)
	if erro != nil {
		if strings.Contains(erro.Error(), "broken pipe") {
			logger.Erro(fmt.Sprintf("Veiculo desconectado durante comunicacao: %v", erro))
		} else {
			logger.Erro(fmt.Sprintf("Erro ao enviar dados da regiao: %v", erro))
		}
		return
	}
}
