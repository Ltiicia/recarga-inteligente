package handler

import (
	"fmt"
	"net"
	"recarga-inteligente/internal/dataJson"
	"recarga-inteligente/internal/distancia"
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

	//Personaliza a resposta
	var mensagemResposta dataJson.Mensagem
	if tipoCliente == "veiculo" {
		connectionStore.AddVeiculo(conexao)
		logger.Info(fmt.Sprintf("Novo veiculo conectado: (%s)", conexao.RemoteAddr()))
		mensagemResposta = dataJson.Mensagem{
			Tipo:     "get-localizacao",
			Conteudo: "Ola Veiculo! Informe sua localizacao atual.",
			Origem:   "servidor",
		}
		//envia
		erro = dataJson.SendMessage(conexao, mensagemResposta)
		if erro != nil {
			logger.Erro(fmt.Sprintf("Erro ao enviar saudacao ao %s: %v", tipoCliente, erro))
			return
		}
		erro = dataJson.SendDadosRegiao(conexao)
		if erro != nil {
			logger.Erro(fmt.Sprintf("Erro ao enviar dados da regiao: %v", erro))
			return
		}
	} else if tipoCliente == "ponto-de-recarga" {
		idPonto := connectionStore.AddPontoRecarga(conexao)
		if idPonto == -1 {
			logger.Erro("Ponto de recarga nao cadastrado tentando se conectar")
			logger.Info(fmt.Sprintf("Ponto de recarga desconhecido desconectado: %s", conexao.RemoteAddr()))
			conexao.Close()
			return
		}
		logger.Info(fmt.Sprintf("Novo ponto de recarga conectado id: (%d)", idPonto))
		mensagemResposta = dataJson.Mensagem{
			Tipo:     "id",
			Conteudo: fmt.Sprintf("%d", idPonto),
			Origem:   "servidor",
		}
		//envia
		erro = dataJson.SendMessage(conexao, mensagemResposta)
		if erro != nil {
			logger.Erro(fmt.Sprintf("Erro ao enviar saudacao ao %s: %v", tipoCliente, erro))
			return
		}
	}

	for {
		mensagemRecebida, erro := dataJson.ReceiveMessage(conexao)
		if erro != nil {
			logger.Erro(fmt.Sprintf("Erro ao ler mensagem do %s: %v", tipoCliente, erro))
			return
		}

		if mensagemRecebida.Origem == "ponto-de-recarga" {
			if mensagemRecebida.Tipo == "registro-id" {
				mensagemResposta := dataJson.Mensagem{
					Tipo:     "get-disponibilidade",
					Conteudo: "Ola ponto de recarga! Informe sua disponibilidade / fila",
					Origem:   "servidor",
				}
				erro = dataJson.SendMessage(conexao, mensagemResposta)
				if erro != nil {
					logger.Erro(fmt.Sprintf("Erro ao enviar resposta ao %s: %v", mensagemInicial.Origem, erro))
					return
				}
			}
			logger.Info(fmt.Sprintf("Mensagem recebida do %s id (%d) => %s", tipoCliente, connectionStore.GetIdPonto(conexao), mensagemRecebida.Conteudo))
		} else if mensagemRecebida.Origem == "veiculo" {
			if mensagemRecebida.Tipo == "localizacao" {
				var latitude, longitude float64
				_, erro := fmt.Sscanf(mensagemRecebida.Conteudo, "%f,%f", &latitude, &longitude)
				if erro != nil {
					logger.Erro(fmt.Sprintf("Erro ao receber localizacao: %v", erro))
				} else {
					logger.Info(fmt.Sprintf("Localizacao recebida: Latitude %f, Longitude %f", latitude, longitude))
				}
				totalPontosConectados := connectionStore.GetTotalPontosConectados()

				mapDistancias, erro := calcDistancia(latitude, longitude, totalPontosConectados)
				if erro != nil {
					logger.Erro(fmt.Sprintf("Erro ao calcular distâncias: %v", erro))
					return
				}

				for id, d := range mapDistancias {
					logger.Info(fmt.Sprintf("Distancia para o ponto Id (%d) = %.2f metros", id, d))
				}
			}
			logger.Info(fmt.Sprintf("Mensagem recebida do %s (%s) => %s", tipoCliente, conexao.RemoteAddr(), mensagemRecebida.Conteudo))
		}
	}
	logger.Info(fmt.Sprintf("%s desconectado: %s", tipoCliente, conexao.RemoteAddr()))
}

func calcDistancia(latVeiculo float64, lonVeiculo float64, totalPontos int) (map[int]float64, error) {
	mapDistancias := make(map[int]float64)

	for id := 1; id <= totalPontos; id++ {
		ponto, erro := dataJson.GetPontoId(id)
		if erro == 0 {
			d := distancia.GetDistancia(latVeiculo, lonVeiculo, ponto.Latitude, ponto.Longitude)
			mapDistancias[id] = d
		} else if erro == 2 {
			return mapDistancias, fmt.Errorf("ponto id (%d) nao localizado", id)
		} else {
			return mapDistancias, fmt.Errorf("Erro ao carregar arquivo json")
		}
	}
	return mapDistancias, nil
}
