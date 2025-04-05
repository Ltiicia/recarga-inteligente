package handler

import (
	"fmt"
	"net"
	"recarga-inteligente/internal/dataJson"
	"recarga-inteligente/internal/distancia"
	"recarga-inteligente/internal/logger"
	"recarga-inteligente/internal/store"
	"strings"
)

// Trata a comunicacao com os clientes
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
		}

		switch mensagemRecebida.Origem {
		case "ponto-de-recarga":
			if mensagemRecebida.Tipo == "identificacao" {
				idPonto := connectionStore.AddPontoRecarga(conexao)
				if idPonto == -1 {
					logger.Erro(fmt.Sprintf("Ponto de recarga nao cadastrado tentando se conectar -> desconectado: %s", conexao.RemoteAddr()))
					connectionStore.RemoveConnection(conexao)
					on = false
				}
				logger.Info(fmt.Sprintf("Novo ponto de recarga conectado id: (%d)", idPonto))
			}
			//servidor solicita ao ponto sua disponibilidade
			id := connectionStore.GetIdPonto(conexao)
			disponibilidade := disponibilidadePonto(logger, conexao, id)
			logger.Info(fmt.Sprintf("Disponibilidade do Ponto id (%d) recebida: %s", id, disponibilidade.Conteudo))

		case "veiculo":
			switch mensagemRecebida.Tipo {
			case "identificacao":
				connectionStore.AddVeiculo(conexao)
				logger.Info(fmt.Sprintf("Novo veiculo conectado: (%s)", conexao.RemoteAddr()))
			case "get-recarga":
				solicitacao := dataJson.Mensagem{
					Tipo:     "get-localizacao",
					Conteudo: "Ola Veiculo! Informe sua localizacao atual.",
					Origem:   "servidor",
				}
				erro = dataJson.SendMessage(conexao, solicitacao)
				if erro != nil {
					logger.Erro(fmt.Sprintf("Erro ao solicitar localizacao ao veiculo: %v", erro))
					return
				}
				erro = dataJson.SendDadosRegiao(conexao)
				if erro != nil {
					if strings.Contains(erro.Error(), "broken pipe") {
						logger.Erro(fmt.Sprintf("Veiculo desconectado durante comunicacao: %v", erro))
					} else {
						logger.Erro(fmt.Sprintf("Erro ao enviar dados da regiao: %v", erro))
					}
					return
				}
			case "localizacao":
				totalPontosConectados := connectionStore.GetTotalPontosConectados()
				mapDistancias := calcDistanciaParaPontos(logger, mensagemRecebida, totalPontosConectados)
				for id, d := range mapDistancias {
					logger.Info(fmt.Sprintf("Distancia para o ponto Id (%d) = %.2f km", id, d))
				}
			case "consultar_pagamento":
				logger.Info("Em breve consulta disponivel")
			default:
				logger.Erro(fmt.Sprintf("Tipo de solicitacao ainda nao foi mapeada - %s", mensagemRecebida.Tipo))
			}
		default:
			logger.Info("Encerrando conexao")
			return
		}
	}
}

func calcDistanciaParaPontos(logger *logger.Logger, mensagemRecebida dataJson.Mensagem, totalPontosConectados int) map[int]float64 {
	var latitude, longitude float64
	_, erro := fmt.Sscanf(mensagemRecebida.Conteudo, "%f,%f", &latitude, &longitude)
	if erro != nil {
		logger.Erro(fmt.Sprintf("Erro ao receber localizacao: %v", erro))
		return map[int]float64{}
	}
	logger.Info(fmt.Sprintf("Localizacao recebida: Latitude %f, Longitude %f", latitude, longitude))

	mapDistancias, erro := calcDistancia(latitude, longitude, totalPontosConectados)
	if erro != nil {
		logger.Erro(fmt.Sprintf("Erro ao calcular distâncias: %v", erro))
		return map[int]float64{}
	}

	return mapDistancias
}

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

func disponibilidadePonto(logger *logger.Logger, conexao net.Conn, id int) dataJson.Mensagem {
	solicitacao := dataJson.Mensagem{
		Tipo:     "get-disponibilidade",
		Conteudo: fmt.Sprintf("Ola ponto de recarga id (%d)! Informe sua disponibilidade / fila", id),
		Origem:   "servidor",
	}
	erro := dataJson.SendMessage(conexao, solicitacao)
	if erro != nil {
		logger.Erro(fmt.Sprintf("Erro ao solicitar disponibilidade ao ponto-de-recarga id (%d): %v", id, erro))
		return dataJson.Mensagem{}
	}
	disponibilidade, erro := dataJson.ReceiveMessage(conexao)
	if erro != nil {
		logger.Erro(fmt.Sprintf("Erro ao receber disponibilidade do %s id (%d): %v", disponibilidade.Origem, id, erro))
		return dataJson.Mensagem{}
	}
	return disponibilidade
}
