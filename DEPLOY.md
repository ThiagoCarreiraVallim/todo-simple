# Deploy no Dokploy

Stack self-contained: **Postgres + API (Go) + Web (Next.js)** numa única
aplicação Compose. A Traefik do Dokploy cuida de TLS/Let's Encrypt e roteamento
por domínio. As migrations rodam sozinhas quando a API sobe.

Topologia: dois subdomínios.

- `WEB_DOMAIN` → app Next.js
- `API_DOMAIN` → API Go (o navegador chama a API direto; CORS já liberado para `WEB_DOMAIN`)

Arquivos que sustentam o deploy:

- `infra/docker-compose.dokploy.yml` — a stack
- `infra/api.Dockerfile`, `infra/web.Dockerfile` — builds multi-stage
- `.env.dokploy.example` — variáveis necessárias

## Pré-requisitos

1. Repositório num Git acessível pelo Dokploy (GitHub/GitLab/Gitea) — o Dokploy
   builda a partir do Git, então **faça commit e push** de tudo.
2. Dois registros DNS **A** apontando para o IP do servidor Dokploy:
   - `app.seudominio.com`
   - `api.seudominio.com`

## Passo a passo no painel do Dokploy

1. **Create Project** → dentro dele, **Create Service → Compose**.
2. **Provider / Source**: conecte o repositório Git e escolha a branch (`main`).
3. **Compose Path**: `infra/docker-compose.dokploy.yml`
4. **Environment**: cole as variáveis abaixo, ajustando os valores
   (base em `.env.dokploy.example`):

   ```
   WEB_DOMAIN=app.seudominio.com
   API_DOMAIN=api.seudominio.com
   POSTGRES_USER=todo
   POSTGRES_DB=todo
   POSTGRES_PASSWORD=<senha forte, ex.: openssl rand -base64 24>
   ```

5. **Deploy**. O Dokploy builda as imagens, sobe a stack e a Traefik emite os
   certificados. O primeiro build (Go + Next standalone) leva alguns minutos.

Não é preciso configurar domínios pela aba "Domains" do Dokploy: as labels da
Traefik já estão no compose. Basta a `dokploy-network` existir (padrão numa
instalação Dokploy) e os entrypoints `web`/`websecure` + certresolver
`letsencrypt` (também padrão).

## Verificação pós-deploy

- `https://api.seudominio.com/health` → `{"status":"ok",...}`
- `https://app.seudominio.com` → abre o app; criar uma lista deve redirecionar
  para `/l/{slug}` e as tarefas devem persistir ao recarregar.
- Abrir o `/l/{slug}` em outro navegador/dispositivo deve carregar a mesma lista
  (prova o modelo de acesso por link).

## Observações

- **Persistência**: o Postgres grava no volume `todo_pg_data`. Redeploys
  preservam os dados; só um `compose down -v` apaga o volume. Configure backup
  do volume (ou snapshots do servidor) se os dados importarem.
- **Trocar de domínio**: `NEXT_PUBLIC_API_URL` é embutido no bundle do Next no
  momento do build. Se mudar `API_DOMAIN`, faça **redeploy** (rebuild) para o
  front passar a apontar para o novo endereço.
- **Postgres gerenciado (alternativa)**: se preferir usar um banco separado
  (serviço "Database" do Dokploy ou externo), remova o serviço `postgres` e a
  rede `internal` do compose e defina `DATABASE_URL` apontando para ele.
- **Domínio único (alternativa)**: dá para servir web e API no mesmo host com a
  API sob `/api`, eliminando o CORS. Exige regras de PathPrefix/priority na
  Traefik — não incluído aqui por ser mais frágil; os dois subdomínios são o
  caminho robusto.
