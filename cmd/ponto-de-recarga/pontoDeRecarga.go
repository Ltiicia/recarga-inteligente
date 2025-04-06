package main

import (
	"fmt"
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

func processarFila(logger *logger.Logger, conexao net.Conn) {
	for {
		mutex.Lock()
		if len(fila) == 0 {
			mutex.Unlock()
			time.Sleep(1 * time.Second)
			continue
		}
		veiculo := fila[0]
		fila = fila[1:]
		mutex.Unlock()

		logger.Info(fmt.Sprintf("Atendendo veículo: %s", veiculo))
		time.Sleep(5 * time.Second) // Simula tempo de recarga
		logger.Info(fmt.Sprintf("Recarga finalizada para: %s", veiculo))

		msg := dataJson.Mensagem{
			Tipo:     "recarga-finalizada",
			Conteudo: fmt.Sprintf("Veículo %s atendido", veiculo),
			Origem:   "ponto-de-recarga",
		}
		dataJson.SendMessage(conexao, msg)
	}
}

func IdentificacaoInicial(logger *logger.Logger, conexao net.Conn) {
	erro := tcpIP.SendIdentification(conexao, "ponto-de-recarga")
	if erro != nil {
		logger.Erro(fmt.Sprintf("Erro ao obter resposta do servidor - %v", erro))
		return
	}
}

func main() {
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

	for {
		msg, erro := dataJson.ReceiveMessage(conexao)
		if erro != nil {
			logger.Erro(fmt.Sprintf("Erro ao ler mensagem do servidor - %v", erro))
			return
		}

		switch msg.Tipo {
		case "get-disponibilidade":
			enviarDisponibilidade(logger, conexao)
		case "nova-solicitacao":
			mutex.Lock()
			fila = append(fila, msg.Conteudo)
			logger.Info(fmt.Sprintf("Veículo adicionado à fila: %s", msg.Conteudo))
			mutex.Unlock()
		default:
			logger.Info(fmt.Sprintf("Mensagem recebida do servidor: %s", msg.Conteudo))
		}
	}
}
