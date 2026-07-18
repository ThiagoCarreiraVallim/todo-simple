// Índice local de listas conhecidas. Como não há login, o navegador guarda
// os slugs (chaves de acesso) das listas que o usuário criou ou abriu.
const STORAGE_KEY = 'todo.lists.v1';

export interface KnownList {
  slug: string;
  name: string;
  color: string;
  lastOpenedAt: string;
}

function read(): KnownList[] {
  if (typeof window === 'undefined') return [];
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    return parsed.filter(
      (entry): entry is KnownList =>
        typeof entry?.slug === 'string' && typeof entry?.name === 'string',
    );
  } catch {
    return [];
  }
}

function write(lists: KnownList[]) {
  if (typeof window === 'undefined') return;
  window.localStorage.setItem(STORAGE_KEY, JSON.stringify(lists));
}

export function getKnownLists(): KnownList[] {
  // Ordena por mais recente; empate de timestamp (mesmo ms) desempata pela
  // ordem de inserção — o último inserido é o mais recente.
  return read()
    .map((list, index) => ({ list, index }))
    .sort((a, b) => {
      const cmp = (b.list.lastOpenedAt ?? '').localeCompare(a.list.lastOpenedAt ?? '');
      return cmp !== 0 ? cmp : b.index - a.index;
    })
    .map((entry) => entry.list);
}

export function upsertKnownList(list: { slug: string; name: string; color: string }) {
  const lists = read().filter((l) => l.slug !== list.slug);
  lists.push({ ...list, lastOpenedAt: new Date().toISOString() });
  write(lists);
}

export function removeKnownList(slug: string) {
  write(read().filter((l) => l.slug !== slug));
}
