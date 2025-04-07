Este projeto foi desenvolvido para facilitar a comunicação eficiente entre veículos elétricos e pontos de recarga. Utilizando uma arquitetura cliente-servidor baseada no protocolo TCP/IP, o sistema permite que veículos solicitem recargas, informem sua localização atual e recebam recomendações para pontos de recarga próximos.  

O objetivo é otimizar o processo de recarga, garantindo eficiência e gerenciamento adequado da concorrência.  

---

## Sumário

- [Introdução](#introdução)
- [Funcionalidades](#funcionalidades)
- [Arquitetura do Sistema](#arquitetura-do-sistema)
- [Protocolo de Comunicação](#protocolo-de-comunicação)
- [Gerenciamento de Concorrência](#gerenciamento-de-concorrência)
- [Como Executar](#como-executar)
- ...
- [Referências](#referências)

---

## Introdução

O presente sistema foi desenvolvido para implementar comunicação entre cliente-servidor simulando o contexto de recarga de veículos elétricos. O projeto viabiliza a solicitação e gestão de recargas por parte dos veículos, utilizando o protocolo TCP/IP e desenvolvimento em Go, com suporte para múltiplas conexões simultâneas.

A aplicação está contida em containers Docker, que isolam e orquestram a execução dos serviços. Onde o servidor gerencia os pontos de recarga disponíveis, recebe solicitações dos veículos, calcula distâncias, gerencia as filas e administra as reservas. Ele é responsável por validar as transações de recarga, verificando a disponibilidade dos pontos, e tratando o armazenamento das informações. O veículo, por sua vez, permite ao usuário solicitar recargas, informa sua localização atual para consultar pontos de recarga disponíveis e escolher onde realizar a operação. Já o ponto de recarga, é responsável por conectar-se ao servidor, informar sua disponibilidade ou fila de espera e gerenciar localmente sua própria fila de recargas. Ao receber uma reserva, o ponto de recarga processa o atendimento do veículo, atualiza sua fila e libera o ponto após a conclusão do carregamento.Porcionando então, uma solução que permite aos veículos encontrar, reservar e utilizar pontos de recarga de forma otimizada, considerando fatores como proximidade e disponibilidade.

---

## Funcionalidades

- **Solicitação de Recarga**: O veículo pode solicitar uma recarga ao servidor.
- **Envio de Localização**: O servidor solicita e recebe a localização atual do veículo.
- **Consulta de Disponibilidade**: O servidor consulta os pontos de recarga conectados sobre sua disponibilidade ou fila de espera.
- **Cálculo de Distância**: O servidor calcula a distância entre o veículo e os pontos de recarga disponíveis.
- **Seleção de Ponto de Recarga**: O veículo recebe as opções e seleciona o ponto desejado.
- **Gerenciamento de Fila**: O servidor efetua a reserva adicionando o veículo à fila do ponto de recarga escolhido.
- **Finalização e Liberação**: O veículo é removido da fila ao final da recarga e recebe o valor para pagamento.

---

## Arquitetura do Sistema

O sistema é baseado em uma arquitetura **cliente-servidor**, onde:

- **Veículo**: Responsável por solicitar recargas, informar sua localização e escolher o ponto de recarga.
- **Servidor**: Gerencia as solicitações, consulta os pontos, calcula distâncias e gerenciar as filas.
- **Ponto de Recarga**: Mantém sua fila local e responde às requisições de disponibilidade do servidor.

A comunicação entre as partes ocorre via **sockets TCP/IP**. Abaixo, a ilustração da arquitetura:

<div align="center">  
  <img align="center" width=100% src= public/sistema-recarga.png alt="Comunicação sistema">
  <p><em>Arquitetura do Sistema</em></p>
</div>


---

### Comunicação

1. Veículo solicita recarga ao servidor.
2. Servidor solicita localização atual ao veículo.
3. Veículo envia latitude e longitude atual.
4. Servidor solicita disponibilidade/fila aos pontos de recarga conectados.
5. Pontos enviam disponibilidade/fila atual.
6. Servidor calcula distâncias do veículo até os pontos.
7. Servidor envia melhores opções ao veículo.
8. Veículo seleciona um ponto e solicita reserva.
9. Servidor confirma a reserva e adiciona o veículo à fila do ponto.
10. Veículo se desloca e realiza a recarga.
11. Ponto remove o veículo da sua fila ao final da recarga.
12. O valor da recarga é vinculado ao veículo.

## Protocolo de Comunicação
...

## Gerenciamento de Concorrência
...

## Tecnologias Utilizadas
- Linguagem: Go (Golang)
- Comunicação: TCP/IP com net.Conn
- Contêineres: Docker, Docker Compose
- Mock de dados: JSON

## Como Executar
...

## Desenvolvedoras
<table>
  <tr>
    <td align="center"><img style="" src="https://avatars.githubusercontent.com/u/142849685?v=4" width="100px;" alt=""/><br /><sub><b> Brenda Araújo </b></sub></a><br />👨‍💻</a></td>
    <td align="center"><img style="" src="https://avatars.githubusercontent.com/u/89545660?v=4" width="100px;" alt=""/><br /><sub><b> Naylane Ribeiro </b></sub></a><br />👨‍💻</a></td>
    <td align="center"><img style="" src="https://avatars.githubusercontent.com/u/124190885?v=4" width="100px;" alt=""/><br /><sub><b> Letícia Gonçalves </b></sub></a><br />👨‍💻</a></td>    
  </tr>
</table>

## Referências
