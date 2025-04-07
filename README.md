# Trabalhos de Sistemas Distribuídos
 
## Trabalho 1 - RPC

Este projeto tem como objetivo implementar uma chamada remota de procedimento (RPC) utilizando a biblioteca [RPyC](https://rpyc.readthedocs.io/en/latest/) em Python.

### Objetivos

- Implementar um servidor com métodos e atributos acessíveis remotamente.
- Criar um cliente que interaja com o servidor via RPC.
- Realizar experimentos de desempenho local e remoto.

## Trabalho 2 - Raft

Este projeto foi desenvolvido com base no artigo que define o algoritmo de consenso RAFT, intitulado [*In Search of an Understandable Consensus Algorithm*](https://raft.github.io/raft.pdf), de Diego Ongaro e John Ousterhout, da Universidade de Stanford.

Utilizamos como referência o laboratório [Lab2 - RAFT](http://www.cs.bu.edu/~jappavoo/jappavoo.github.com/451/labs/lab-raft.html), parte do curso **CS451/651 - Distributed Systems**, oferecido pela Universidade de Boston.

### Configurando ambiente
Primeiro clone o repositório: <br/>
`git clone https://github.com/SerranoZz/trabalho-sd.git`

#### Configurando variáveis de ambiente (Windows)
1. Configurando variável GOROOT: <br/>
`$env:GOROOT = "Local de instalação do seu Go"`
2. Configurando variável PATH: <br/>
`$env:PATH += ";$env:GOROOT\bin"`
3. Configurando variável GOPATH: <br/>
`$env:GOPATH = "~/trabalho-sd/'Trabalho 2 - Raft'"`

#### Configurando variáveis de ambiente (Linux)
1. Configurando variável GOROOT: <br/>
`export GOROOT= Local de instalação do seu Go`
2. Configurando variável PATH: <br/>
`export PATH=$PATH:$GOROOT/bin`
3. Configurando variável GOPATH: <br/>
`export GOPATH="~/trabalho-sd/'Trabalho 2 - Raft'"`

#### Configurando repositório
1. Vá para o diretório: <br/>
 Windows: `cd $env:GOPATH` <br/> Linux: `cd $GOPATH`
2. Execute o comando Go: <br/>
`go mod init trab-sd`
3. Vá para o diretório: <br/>
`cd src/labrpc`
4. Execute o comando Go: <br/>
`go mod init labrpc`
5. Vá para o diretório: <br/>
`cd ../..`
6. Execute os comandos Go: <br/>
`go mod edit -replace=labrpc=".\src\labrpc" && go get labrpc`
7. Vá para o diretório: <br/>
`cd src/raft`
8. Execute o comando: <br/>
`go test -run 2A`

## Colaboradores

Os trabalhos foram realizados em grupo com:

- [Bruno Adji](https://github.com/brunoadji)
- [Lucas Serrano](https://github.com/SerranoZz)
