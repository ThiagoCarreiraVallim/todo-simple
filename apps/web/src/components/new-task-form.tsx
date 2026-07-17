'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { useAddTask } from '@/hooks/use-list';

export function NewTaskForm({ slug }: { slug: string }) {
  const [title, setTitle] = useState('');
  const addTask = useAddTask(slug);

  return (
    <form
      className="flex gap-2"
      onSubmit={(e) => {
        e.preventDefault();
        const trimmed = title.trim();
        if (!trimmed) return;
        addTask.mutate(trimmed, { onSuccess: () => setTitle('') });
      }}
    >
      <Input
        value={title}
        onChange={(e) => setTitle(e.target.value)}
        placeholder="Nova tarefa"
        maxLength={500}
      />
      <Button type="submit" disabled={addTask.isPending || !title.trim()}>
        Adicionar
      </Button>
    </form>
  );
}
