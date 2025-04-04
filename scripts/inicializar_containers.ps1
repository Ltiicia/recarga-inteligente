Write-Host "Construindo e inicializando os containeres Docker..." -ForegroundColor Green

#Executa todos os containers com o docker-compose
docker-compose up --build -d

#add -d para executar os containers em segundo plano no terminal

#Para rodar, acessar o dir do projeto:
#.\scripts\inicializar_containers.ps1

#Permissao: Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass