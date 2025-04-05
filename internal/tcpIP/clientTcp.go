package tcpIP

import (
	"fmt"
	"net"
	"recarga-inteligente/internal/dataJson"
)

func ConnectToServerTCP(serverAddress string) (net.Conn, error) {
	conexao, erro := net.Dial("tcp", serverAddress)
	if erro != nil {
		return nil, fmt.Errorf("erro ao conectar ao servidor: %v", erro)
	}
	return conexao, nil
}

func SendIdentification(conexao net.Conn, origem string) error {
	msg := dataJson.Mensagem{
		Tipo:     "identificacao",
		Conteudo: fmt.Sprintf("%s conectado", origem),
		Origem:   origem,
	}

	erro := dataJson.SendMessage(conexao, msg)
	if erro != nil {
		return fmt.Errorf("erro ao enviar identificacao: %v", erro)
	}

	return nil
}
