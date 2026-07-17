'use client';

import { useEffect, useState } from 'react';
import { Button } from '@/components/ui/button';
import { ColorPicker } from '@/components/color-picker';
import { NewTaskForm } from '@/components/new-task-form';
import { TaskList } from '@/components/task-list';
import { useDeleteList, useList, useUpdateList } from '@/hooks/use-list';
import { ApiError } from '@/lib/api';
import { colorClasses, type ListColor } from '@/lib/colors';
import { removeKnownList } from '@/lib/local-index';
import { cn } from '@/lib/utils';

interface ListViewProps {
  slug: string;
  /** chamada quando a lista deixa de existir (404 ou exclusão) */
  onGone: () => void;
  /** chamada quando nome/cor mudam, para as abas refletirem */
  onIndexChange: () => void;
}

export function ListView({ slug, onGone, onIndexChange }: ListViewProps) {
  const { data: list, isLoading, error } = useList(slug);
  const updateList = useUpdateList(slug);
  const deleteList = useDeleteList(slug, onGone);
  const [renaming, setRenaming] = useState(false);
  const [draft, setDraft] = useState('');
  const [copied, setCopied] = useState(false);

  const notFound = error instanceof ApiError && error.status === 404;

  useEffect(() => {
    if (notFound) removeKnownList(slug);
  }, [notFound, slug]);

  useEffect(() => {
    if (list) onIndexChange();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [list?.name, list?.color]);

  if (notFound) {
    return (
      <div className="space-y-3 py-8 text-center">
        <p className="text-sm text-muted-foreground">
          Lista não encontrada. Ela pode ter sido excluída.
        </p>
        <Button variant="outline" onClick={onGone}>
          Voltar
        </Button>
      </div>
    );
  }

  if (isLoading) {
    return <p className="py-8 text-center text-sm text-muted-foreground">Carregando...</p>;
  }

  if (error || !list) {
    return (
      <p className="py-8 text-center text-sm text-destructive">
        Não foi possível carregar a lista.
      </p>
    );
  }

  const saveRename = () => {
    const trimmed = draft.trim();
    setRenaming(false);
    if (trimmed && trimmed !== list.name) updateList.mutate({ name: trimmed });
  };

  const copyLink = async () => {
    await navigator.clipboard.writeText(`${window.location.origin}/l/${slug}`);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const confirmDelete = () => {
    if (window.confirm(`Excluir a lista "${list.name}" e todas as suas tarefas?`)) {
      deleteList.mutate();
    }
  };

  return (
    <div className="space-y-5">
      <header className="space-y-3">
        <div className="flex items-center gap-3">
          <span className={cn('h-3 w-3 shrink-0 rounded-full', colorClasses(list.color).dot)} />
          {renaming ? (
            <input
              value={draft}
              onChange={(e) => setDraft(e.target.value)}
              onBlur={saveRename}
              onKeyDown={(e) => {
                if (e.key === 'Enter') saveRename();
                if (e.key === 'Escape') setRenaming(false);
              }}
              maxLength={120}
              autoFocus
              className="flex-1 bg-transparent text-xl font-semibold outline-none"
            />
          ) : (
            <h1
              onDoubleClick={() => {
                setDraft(list.name);
                setRenaming(true);
              }}
              title="Clique duas vezes para renomear"
              className="flex-1 truncate text-xl font-semibold"
            >
              {list.name}
            </h1>
          )}
          <Button variant="outline" size="sm" onClick={copyLink}>
            {copied ? 'Link copiado!' : 'Copiar link'}
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={confirmDelete}
            className="text-muted-foreground hover:text-destructive"
          >
            Excluir
          </Button>
        </div>
        <ColorPicker
          value={list.color}
          onChange={(color: ListColor) => updateList.mutate({ color })}
        />
      </header>

      <NewTaskForm slug={slug} />
      <TaskList slug={slug} color={list.color} tasks={list.tasks} />
    </div>
  );
}
