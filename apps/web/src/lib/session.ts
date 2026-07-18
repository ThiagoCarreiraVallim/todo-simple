// Sessão do usuário no navegador. O login (sem senha) resolve um userId (uuid),
// que é a chave que vincula a fazenda e as listas. Guardamos userId + username
// (para reexibir/relogar) + name (nome de exibição da fazenda).
const STORAGE_KEY = 'todo.session.v1';

export interface StoredSession {
  userId: string;
  username: string;
  name: string;
}

export function getSession(): StoredSession | null {
  if (typeof window === 'undefined') return null;
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) return null;
    const parsed = JSON.parse(raw);
    if (typeof parsed?.userId !== 'string' || typeof parsed?.username !== 'string') return null;
    return {
      userId: parsed.userId,
      username: parsed.username,
      name: typeof parsed.name === 'string' ? parsed.name : '',
    };
  } catch {
    return null;
  }
}

export function setSession(session: StoredSession) {
  if (typeof window === 'undefined') return;
  window.localStorage.setItem(STORAGE_KEY, JSON.stringify(session));
}

export function clearSession() {
  if (typeof window === 'undefined') return;
  window.localStorage.removeItem(STORAGE_KEY);
}
