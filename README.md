# Trabalhos de Sistemas Distribuídos
## Trabalho 1 - RPC

## Trabalho 2 - Raft
### Configurando ambiente
1. Clone o repositório: <br/>
`$ git clone https://github.com/SerranoZz/trabalho-sd.git`

#### Configurando variáveis de ambiente (Windows)
1.1 Configurando variável GOROOT: <br/>
`$env:GOROOT = "C:\Program Files\Go"`
2.1 Configurando variável PATH: <br/>
`$env:PATH += ";$env:GOROOT\bin"`
3.1 Configurando variável GOPATH: <br/>
`$env:GOPATH = "~/trabalho-sd/'Trabalho 2 - Raft'"`

#### Configurando repositório
1. Vá para o diretório: <br/>
`$ cd $env:GOPATH`
2. Execute o comando Go: <br/>
`$ go mod init trab-sd`
3. Vá para o diretório: <br/>
`$ cd src/labrpc`
4. Execute o comando Go: <br/>
`$ go mod init labrpc`
5. Vá para o diretório: <br/>
`$ cd ../..`
6. Execute os comandos Go: <br/>
`$ go mod edit -replace=labrpc=".\src\labrpc" && go get labrpc`
7. Vá para o diretório: <br/>
`$ cd src/raft`
8. Execute o comando: <br/>
`$ go test -run 2A`
