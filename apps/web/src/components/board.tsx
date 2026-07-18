'use client';

import { useEffect, useRef, useState } from 'react';
import { FarmPanel } from '@/components/farm-panel';
import { ListTabs } from '@/components/list-tabs';
import { ListView } from '@/components/list-view';
import { NewListForm } from '@/components/new-list-form';
import { useFarmSession } from '@/lib/farm-session';
import { getKnownLists, type KnownList } from '@/lib/local-index';

interface BoardProps {
  /** slug vindo da rota /l/[slug] (link aberto ou compartilhado) */
  initialSlug?: string;
}

// Tela única do app: abas com as listas conhecidas e a lista ativa abaixo.
// O índice de listas vive no localStorage, então tudo renderiza pós-mount.
export function Board({ initialSlug }: BoardProps) {
  const [mounted, setMounted] = useState(false);
  const [known, setKnown] = useState<KnownList[]>([]);
  const [activeSlug, setActiveSlug] = useState<string | null>(initialSlug ?? null);
  const [creating, setCreating] = useState(false);
  const { userId } = useFarmSession();

  useEffect(() => {
    const lists = getKnownLists();
    setKnown(lists);
    if (!initialSlug) {
      if (lists.length > 0) syncUrl(lists[0].slug);
      setActiveSlug(lists[0]?.slug ?? null);
    }
    setMounted(true);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Ao logar/deslogar, o índice local muda (listas do usuário são mescladas):
  // relê e, se nenhuma aba estiver ativa, seleciona a primeira.
  const firstUserEffect = useRef(true);
  useEffect(() => {
    if (firstUserEffect.current) {
      firstUserEffect.current = false;
      return;
    }
    const lists = getKnownLists();
    setKnown(lists);
    if (lists.length > 0) setCreating(false);
    setActiveSlug((current) => {
      if (current) return current;
      const next = lists[0]?.slug ?? null;
      syncUrl(next);
      return next;
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [userId]);

  // Troca de aba sem navegação: a URL acompanha para o link ser compartilhável.
  const syncUrl = (slug: string | null) => {
    window.history.replaceState(null, '', slug ? `/l/${slug}` : '/');
  };

  const refreshIndex = () => setKnown(getKnownLists());

  const select = (slug: string) => {
    setCreating(false);
    setActiveSlug(slug);
    syncUrl(slug);
  };

  const handleCreated = (slug: string) => {
    refreshIndex();
    select(slug);
  };

  const handleGone = () => {
    const lists = getKnownLists();
    setKnown(lists);
    const next = lists[0]?.slug ?? null;
    setActiveSlug(next);
    syncUrl(next);
    if (!next) setCreating(true);
  };

  if (!mounted) return null;

  const showNewForm = creating || known.length === 0;

  return (
    <div className="mx-auto max-w-xl space-y-5 px-5 py-8">
      <ListTabs
        lists={known}
        activeSlug={activeSlug}
        creating={showNewForm}
        onSelect={select}
        onNew={() => setCreating(true)}
      />
      {showNewForm ? (
        <NewListForm onCreated={handleCreated} />
      ) : (
        activeSlug && (
          <ListView
            key={activeSlug}
            slug={activeSlug}
            onGone={handleGone}
            onIndexChange={refreshIndex}
          />
        )
      )}
      {/* A fazenda é global (por-pessoa) e fica no rodapé, sob todas as listas. */}
      <FarmPanel />
    </div>
  );
}
