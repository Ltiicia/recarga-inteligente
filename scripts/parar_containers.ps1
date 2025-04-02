Write-Host "Parando e removendo os containeres Docker..." -ForegroundColor Green

#Para e remove os containers em execucao
docker-compose down

Write-Host "Containeres parados e removidos com sucesso!" -ForegroundColor Green

#Exibe os containers em execucao
docker ps -a
docker-compose ps -a

#Para rodar, acessar o dir do projeto:
#.\scripts\parar_containers.ps1