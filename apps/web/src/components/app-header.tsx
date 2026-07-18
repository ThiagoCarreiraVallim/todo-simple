'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { useLogin } from '@/hooks/use-farm';
import { ApiError } from '@/lib/api';
import { useFarmSession } from '@/lib/farm-session';

// Cabeçalho do app: título à esquerda e, no canto superior direito, o login
// (só nome de usuário). Logado, mostra o usuário e "Sair".
export function AppHeader() {
  const { username, name, ready, logout } = useFarmSession();

  return (
    <header className="border-b border-border">
      <div className="mx-auto flex max-w-xl items-center justify-between px-5 py-3">
        <span className="text-sm font-semibold">Tarefas</span>
        {ready &&
          (username ? (
            <div className="flex items-center gap-3 text-sm">
              <span className="font-medium">🚜 {name || username}</span>
              <Button variant="ghost" size="sm" onClick={logout}>
                Sair
              </Button>
            </div>
          ) : (
            <LoginControl />
          ))}
      </div>
    </header>
  );
}

function LoginControl() {
  const [open, setOpen] = useState(false);
  const [value, setValue] = useState('');
  const login = useLogin();

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    const u = value.trim();
    if (!u) return;
    login.mutate(u, { onSuccess: () => setOpen(false) });
  };

  const error =
    login.error instanceof ApiError && login.error.status === 400
      ? 'Use 3–20 caracteres: letras, números, _ ou -'
      : login.error
        ? 'Não foi possível entrar. Tente de novo.'
        : null;

  return (
    <div className="relative">
      <Button size="sm" onClick={() => setOpen((o) => !o)}>
        Entrar
      </Button>
      {open && (
        <div className="absolute right-0 top-full z-10 mt-2 w-64 rounded-md border border-border bg-card p-3 shadow-lg">
          <p className="mb-2 text-xs text-muted-foreground">
            Entre com um nome de usuário para ganhar sua fazenda. Use o mesmo nome em qualquer
            aparelho para continuar.
          </p>
          <form onSubmit={submit} className="space-y-2">
            <Input
              value={value}
              onChange={(e) => setValue(e.target.value)}
              placeholder="nome de usuário"
              maxLength={20}
              autoFocus
            />
            {error && <p className="text-xs text-red-500">{error}</p>}
            <Button type="submit" size="sm" className="w-full" disabled={login.isPending}>
              {login.isPending ? 'Entrando…' : 'Entrar'}
            </Button>
          </form>
        </div>
      )}
    </div>
  );
}
