# todo-simple

Listas de tarefas minimalistas, **sem login** — o link é a chave.

Cada lista recebe uma URL com um slug impossível de adivinhar (`/l/{slug}`,
nanoid de 21 caracteres). Quem tem o link vê e edita a lista; o navegador
guarda no `localStorage` o índice das listas que você criou ou abriu, exibidas
como abas coloridas numa única tela. Compartilhar é enviar o link; continuar
em outro dispositivo é abrir o link lá.

## Stack

- **Turborepo + pnpm workspaces** — `apps/api` (Go) e `apps/web` (Next.js)
- **Go 1.22 API** — chi router, `pgx` (sem ORM), migrations embedadas via
  `go:embed`, graceful shutdown, `log/slog`
- **Next.js 15 (App Router)** — TanStack Query com updates otimistas,
  Tailwind, drag-and-drop com dnd-kit
- **Postgres 16** — tabelas `lists` e `tasks`
- **Docker Compose** para infra local e produção; CI, Husky + lint-staged

Veja [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) para a arquitetura do
template base e [`docs/ADR/`](docs/ADR/) para as decisões.

## Rodando

```bash
pnpm bootstrap   # instala deps, sobe Postgres (Docker), roda migrations
pnpm dev         # API :8080, Web :3000
```

## Roadmap

1. ✅ Listas de tarefas: criar, reordenar (drag), concluir, editar — sem login
2. Múltiplas listas e compartilhamento (base pronta: abas + "Copiar link")
3. Continuar em outros dispositivos (código de sincronização agrupando listas)
4. Extensão Chrome para acompanhar as tarefas
5. Mini-game de bichinho/fazendinha alimentado pela conclusão de tarefas
   (`tasks.completed_at` já registra o histórico desde já)

## API

| Rota | Descrição |
|---|---|
| `POST /api/lists` | cria lista `{name, color?}` → `{slug, ...}` |
| `GET /api/lists/{slug}` | lista + tarefas ordenadas |
| `PATCH /api/lists/{slug}` | renomeia / troca cor |
| `DELETE /api/lists/{slug}` | exclui (cascade) |
| `POST /api/lists/{slug}/tasks` | adiciona tarefa ao final |
| `PATCH /api/lists/{slug}/tasks/{id}` | edita título / conclui |
| `DELETE /api/lists/{slug}/tasks/{id}` | exclui tarefa |
| `PUT /api/lists/{slug}/tasks/order` | reordena (permutação completa; 409 se defasada) |

## License

MIT — see [`LICENSE`](LICENSE).
