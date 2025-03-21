package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	fmt.Println("Inicializando Ponto de Recarga...")

	//Conecta ao servidor (fora do docker trocar servidor por localhost)
	conn, err := net.Dial("tcp", "servidor:5000")
	if err != nil {
		log.Fatalf("Erro ao conectar ao servidor: %v\n", err)
	}
	defer conn.Close()

	//Envia uma mensagem para o servidor
	mensagem := "Servidor, estou disponivel para recargas."
	_, err = conn.Write([]byte(mensagem))
	if err != nil {
		log.Fatalf("Erro ao enviar mensagem: %v\n", err)
	}

	//Recebe a resposta do servidor
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Fatalf("Erro ao ler resposta: %v\n", err)
	}

	fmt.Printf("Resposta do servidor: %s\n", string(buffer[:n]))
}
