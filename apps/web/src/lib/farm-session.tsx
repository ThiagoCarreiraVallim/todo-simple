'use client';

import { createContext, useContext, useEffect, useState, type ReactNode } from 'react';
import { clearSession, getSession, setSession, type StoredSession } from '@/lib/session';

interface FarmSession {
  userId: string | null;
  username: string | null;
  name: string;
  /** true depois de ler o localStorage no mount (evita flicker/hydration). */
  ready: boolean;
  login: (session: StoredSession) => void;
  logout: () => void;
  setName: (name: string) => void;
}

const FarmSessionContext = createContext<FarmSession | null>(null);

// Estado de login compartilhado entre o header (botão Entrar), o board (listas
// do usuário) e o painel da fazenda. Persiste no localStorage via lib/session.
export function FarmSessionProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<StoredSession | null>(null);
  const [ready, setReady] = useState(false);

  useEffect(() => {
    setState(getSession());
    setReady(true);
  }, []);

  const login = (session: StoredSession) => {
    setSession(session);
    setState(session);
  };

  const logout = () => {
    clearSession();
    setState(null);
  };

  const setName = (name: string) => {
    setState((s) => {
      if (!s) return s;
      const next = { ...s, name };
      setSession(next);
      return next;
    });
  };

  return (
    <FarmSessionContext.Provider
      value={{
        userId: state?.userId ?? null,
        username: state?.username ?? null,
        name: state?.name ?? '',
        ready,
        login,
        logout,
        setName,
      }}
    >
      {children}
    </FarmSessionContext.Provider>
  );
}

export function useFarmSession(): FarmSession {
  const ctx = useContext(FarmSessionContext);
  if (!ctx) throw new Error('useFarmSession precisa estar dentro de <FarmSessionProvider>');
  return ctx;
}
