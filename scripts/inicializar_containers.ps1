Write-Host "Construindo e inicializando os containeres Docker..." -ForegroundColor Green

#Executa todos os containers com o docker-compose
docker-compose up --build

Write-Host "Containeres inicializados com sucesso!" -ForegroundColor Green

#Para rodar, acessar o dir do projeto:
#.\scripts\inicializar-container.ps1

#Permissao: Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass