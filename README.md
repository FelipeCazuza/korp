# Projeto DevOps com Go, Docker, Prometheus, Grafana e Ansible

Projeto desenvolvido como solução prática para um desafio técnico de DevOps.

A solução contém uma aplicação HTTP escrita em Go, empacotada em uma imagem Docker e publicada por meio de um proxy reverso NGINX.

A aplicação é monitorada pelo Prometheus, com visualização das métricas em um dashboard provisionado automaticamente no Grafana.

Todo o ambiente pode ser instalado, implantado e validado automaticamente por meio de um playbook Ansible.

---

## Arquitetura

```text
Cliente
   |
   v
NGINX :80
   |
   v
Aplicação Go :8080
   |
   +-- /projeto-korp
   |
   +-- /metrics
          |
          v
     Prometheus :9090
          |
          v
       Grafana :3000
```

Todos os containers são conectados à rede Docker bridge:

```text
projeto-korp-network
```

A aplicação Go utiliza internamente a porta `8080`, mas essa porta não é publicada diretamente no host.

O acesso externo à aplicação ocorre exclusivamente pelo NGINX na porta `80`.

---

## Tecnologias utilizadas

- Go
- Docker
- Docker Compose
- NGINX
- Prometheus
- Grafana
- Ansible
- Ansible Vault

---

## Funcionalidades

- API HTTP desenvolvida em Go.
- Resposta JSON com horário UTC dinâmico.
- Middleware para contabilização de requisições.
- Métricas compatíveis com Prometheus.
- Build Docker multi-stage.
- Execução da aplicação com usuário sem privilégios.
- Rede Docker bridge dedicada.
- Proxy reverso com NGINX.
- Monitoramento com Prometheus.
- Dashboard Grafana provisionado automaticamente.
- Senha do Grafana armazenada fora do código.
- Provisionamento automatizado com Ansible.
- Validação automática da aplicação, Prometheus e Grafana.
- Uso de Ansible Vault para proteção da senha.
- Comportamento idempotente nas tarefas de arquivos, diretórios e rede.

---

## Estrutura do projeto

```text
.
├── ansible/
│   ├── inventory.ini
│   ├── playbook.yml
│   └── vars/
│       ├── main.yml
│       └── vault.yml.example
├── cmd/
│   └── http-server/
│       └── main.go
├── grafana/
│   ├── dashboards/
│   │   └── http-server-projeto-korp-dashboard.json
│   └── provisioning/
│       ├── dashboards/
│       │   └── dashboards.yml
│       └── datasources/
│           └── prometheus.yml
├── nginx/
│   └── conf.d/
│       └── http-server-projeto-korp.conf
├── prometheus/
│   └── prometheus.yml
├── secrets/
│   └── .gitkeep
├── .dockerignore
├── .gitignore
├── compose.yaml
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```

---

## Aplicação Go

A aplicação expõe o endpoint:

```http
GET /projeto-korp
```

Exemplo de resposta:

```json
{
  "nome": "Projeto Korp",
  "horario": "2026-08-07T00:50:38Z"
}
```

O campo `horario` é gerado dinamicamente no momento da requisição e utiliza UTC no formato RFC 3339.

A aplicação também disponibiliza:

```http
GET /metrics
```

Esse endpoint fornece métricas no formato esperado pelo Prometheus.

---

## Métricas

A aplicação registra o total de requisições HTTP por:

- método HTTP;
- caminho acessado;
- código de status retornado.

A métrica personalizada é:

```promql
projeto_korp_http_requests_total
```

Exemplo:

```text
projeto_korp_http_requests_total{
  method="GET",
  path="/projeto-korp",
  status="200"
}
```

O Prometheus também utiliza a métrica padrão de disponibilidade:

```promql
up{job="http-server-projeto-korp"}
```

Interpretação:

```text
1 = aplicação disponível
0 = aplicação indisponível
```

Outras consultas utilizadas no dashboard:

```promql
sum(projeto_korp_http_requests_total)
```

```promql
sum(rate(projeto_korp_http_requests_total[$__rate_interval]))
```

```promql
sum by (status) (
  rate(projeto_korp_http_requests_total[$__rate_interval])
)
```

---

## Dockerfile

O Dockerfile utiliza build multi-stage.

Na primeira etapa, uma imagem Go é utilizada para baixar as dependências e compilar a aplicação.

Na segunda etapa, somente o binário compilado é copiado para uma imagem Alpine menor.

A compilação utiliza:

```text
CGO_ENABLED=0
GOOS=linux
```

A aplicação é executada dentro do container por um usuário sem privilégios administrativos.

---

## Pré-requisitos

Para execução manual com Docker Compose:

- Linux ou WSL2;
- Docker Engine;
- Docker Compose Plugin.

Para provisionamento automatizado:

- Ansible;
- Python 3;
- usuário com acesso a `sudo`.

---

## Execução com Docker Compose

### 1. Clonar o repositório

```bash
git clone URL_DO_REPOSITORIO
cd NOME_DO_REPOSITORIO
```

### 2. Criar a rede Docker

A rede é definida como externa no arquivo Compose:

```bash
docker network inspect projeto-korp-network >/dev/null 2>&1 ||
  docker network create \
    --driver bridge \
    projeto-korp-network
```

### 3. Criar a senha administrativa do Grafana

```bash
mkdir -p secrets

printf '%s' 'DEFINA_UMA_SENHA_SEGURA' \
  > secrets/grafana_admin_password

chmod 700 secrets
chmod 600 secrets/grafana_admin_password
```

O arquivo contendo a senha está ignorado pelo Git.

### 4. Construir e iniciar os containers

```bash
docker compose up -d --build
```

### 5. Verificar o ambiente

```bash
docker compose ps
```

São iniciados quatro serviços:

```text
http-server-projeto-korp
nginx
prometheus
grafana
```

---

## Endpoints

| Componente | Endereço |
|---|---|
| Aplicação pelo NGINX | `http://localhost/projeto-korp` |
| Prometheus | `http://localhost:9090` |
| Grafana | `http://localhost:3000` |

O usuário administrativo padrão do Grafana é:

```text
admin
```

A senha é definida localmente no arquivo:

```text
secrets/grafana_admin_password
```

---

## Testar a aplicação

```bash
curl -i http://localhost/projeto-korp
```

Resposta esperada:

```text
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
```

```json
{
  "nome": "Projeto Korp",
  "horario": "2026-08-07T00:50:38Z"
}
```

A porta `8080` não é publicada no host. Portanto, o acesso direto deve falhar:

```bash
curl http://localhost:8080/projeto-korp
```

O comportamento demonstra que o NGINX é o ponto de entrada da aplicação.

---

## Gerar tráfego para as métricas

Gerar requisições HTTP `200`:

```bash
for i in $(seq 1 100); do
  curl -s http://localhost/projeto-korp >/dev/null
  sleep 0.05
done
```

Gerar requisições HTTP `405`:

```bash
for i in $(seq 1 10); do
  curl -s -X POST http://localhost/projeto-korp >/dev/null
done
```

Depois de alguns segundos, os dados podem ser visualizados no dashboard do Grafana.

---

## Prometheus

A configuração de coleta está em:

```text
prometheus/prometheus.yml
```

O Prometheus acessa a aplicação pelo nome do serviço dentro da rede Docker:

```text
http-server-projeto-korp:8080
```

O endpoint coletado é:

```text
/metrics
```

Validar a prontidão do Prometheus:

```bash
curl -i http://localhost:9090/-/ready
```

Consultar a disponibilidade da aplicação:

```bash
curl -s \
  'http://localhost:9090/api/v1/query?query=up%7Bjob%3D%22http-server-projeto-korp%22%7D'
```

O resultado deve conter:

```json
"value": [
  "timestamp",
  "1"
]
```

---

## Grafana

O datasource Prometheus é provisionado automaticamente por:

```text
grafana/provisioning/datasources/prometheus.yml
```

O provider responsável pelo carregamento dos dashboards está em:

```text
grafana/provisioning/dashboards/dashboards.yml
```

O dashboard versionado está em:

```text
grafana/dashboards/http-server-projeto-korp-dashboard.json
```

O dashboard apresenta:

- disponibilidade da aplicação;
- total de requisições;
- requisições por segundo;
- requisições agrupadas por código HTTP.

Validar a saúde do Grafana:

```bash
curl -s http://localhost:3000/api/health
```

Resposta esperada:

```json
{
  "database": "ok"
}
```

---

## Provisionamento com Ansible

O playbook automatiza as seguintes etapas:

1. Verifica se Docker e Docker Compose estão instalados.
2. Instala o Docker quando necessário.
3. Cria a estrutura da aplicação em `/opt`.
4. Copia os arquivos do projeto.
5. Cria o arquivo secreto do Grafana.
6. Verifica ou cria a rede Docker bridge.
7. Valida o arquivo Docker Compose.
8. Constrói a imagem da aplicação.
9. Inicia os containers.
10. Valida o endpoint HTTP.
11. Consulta a disponibilidade no Prometheus.
12. Valida a saúde do Grafana.
13. Exibe o resultado no console.

---

## Configurar o Ansible Vault

Crie o arquivo local a partir do exemplo:

```bash
cp ansible/vars/vault.yml.example \
  ansible/vars/vault.yml
```

O conteúdo deve possuir:

```yaml
---
vault_grafana_admin_password: "DEFINA_UMA_SENHA_SEGURA"
```

Criptografe o arquivo:

```bash
ansible-vault encrypt ansible/vars/vault.yml
```

Será solicitada uma senha para proteger o Vault.

O arquivo real:

```text
ansible/vars/vault.yml
```

está ignorado pelo Git e não é publicado no repositório.

---

## Inventário Ansible

Por padrão, o inventário executa o provisionamento na própria máquina:

```ini
[projeto_korp]
localhost ansible_connection=local ansible_python_interpreter=/usr/bin/python3
```

Para utilizar um servidor remoto, o inventário pode ser alterado:

```ini
[projeto_korp]
servidor-korp ansible_host=192.168.1.100 ansible_user=ubuntu
```

---

## Validar a sintaxe do playbook

```bash
ansible-playbook \
  -i ansible/inventory.ini \
  ansible/playbook.yml \
  --syntax-check \
  --ask-vault-pass
```

Resultado esperado:

```text
playbook: ansible/playbook.yml
```

---

## Executar o provisionamento completo

```bash
ansible-playbook \
  -i ansible/inventory.ini \
  ansible/playbook.yml \
  -K \
  --ask-vault-pass
```

Parâmetros:

- `-K`: solicita a senha do `sudo`;
- `--ask-vault-pass`: solicita a senha do Ansible Vault.

Exemplo do resultado apresentado pelo playbook:

```text
Status HTTP: 200
Resposta: {"nome":"Projeto Korp","horario":"2026-08-07T00:50:38Z"}
Prometheus: UP
Grafana database: ok
```

---

## Idempotência

O playbook foi executado mais de uma vez para validar seu comportamento.

Na segunda execução:

- os diretórios permaneceram com estado `ok`;
- os arquivos permaneceram com estado `ok`;
- o secret permaneceu sem alteração;
- a rede Docker não foi recriada;
- a instalação do Docker foi ignorada;
- todas as validações funcionais passaram.

A tarefa que executa o Docker Compose pode aparecer como `changed` porque utiliza:

```bash
docker compose up -d --build
```

A opção `--build` solicita uma nova verificação e construção da imagem da aplicação.

---

## Validações realizadas

### Código Go

```bash
go fmt ./...
go vet ./...
go test ./...
```

### Docker Compose

```bash
docker compose config
docker compose ps
```

### Aplicação

```bash
curl -i http://localhost/projeto-korp
```

### Prometheus

```bash
curl -s \
  'http://localhost:9090/api/v1/query?query=up%7Bjob%3D%22http-server-projeto-korp%22%7D'
```

### Grafana

```bash
curl -s http://localhost:3000/api/health
```

### Ansible

```bash
ansible-playbook \
  -i ansible/inventory.ini \
  ansible/playbook.yml \
  --syntax-check \
  --ask-vault-pass
```

---

## Segurança

As seguintes medidas foram aplicadas:

- aplicação executada com usuário sem privilégios;
- build Docker multi-stage;
- senha do Grafana fora do código-fonte;
- variável sensível protegida com Ansible Vault;
- uso de `no_log: true` na tarefa que cria o secret;
- arquivo de senha com permissão `0600`;
- diretório de secrets com permissão `0700`;
- arquivos sensíveis ignorados pelo Git;
- porta `8080` não publicada diretamente no host;
- arquivos de configuração montados como somente leitura;
- cadastro de usuários no Grafana desabilitado;
- datasource e dashboard versionados como código.

---

## Encerrar o ambiente

Parar e remover os containers:

```bash
docker compose down
```

Remover também os volumes de dados:

```bash
docker compose down -v
```

> O segundo comando remove os dados persistidos do Prometheus e do Grafana.

---

## Possíveis melhorias

- Adicionar testes unitários para os handlers e middlewares Go.
- Adicionar métricas de duração das requisições.
- Adicionar healthchecks aos serviços do Docker Compose.
- Utilizar módulos da coleção `community.docker` no Ansible.
- Criar regras de alerta no Prometheus.
- Criar alertas no Grafana.
- Implementar pipeline de integração contínua.
- Publicar a imagem em um container registry.
- Adicionar análise de segurança da imagem.
- Executar o provisionamento em uma máquina virtual limpa durante o pipeline.

---

## Objetivo do projeto

Este projeto foi desenvolvido para praticar e demonstrar conhecimentos em:

- desenvolvimento de serviços HTTP com Go;
- containerização;
- proxy reverso;
- redes Docker;
- observabilidade;
- métricas Prometheus;
- dashboards Grafana;
- automação com Ansible;
- gerenciamento seguro de segredos;
- validação automatizada de infraestrutura.

---

## Licença

Este projeto está disponível para fins de estudo e demonstração técnica.
