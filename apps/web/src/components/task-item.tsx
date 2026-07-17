'use client';

import { useState } from 'react';
import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import type { Task } from '@/lib/api/lists';
import { colorClasses } from '@/lib/colors';
import { cn } from '@/lib/utils';

interface TaskItemProps {
  task: Task;
  color: string;
  onToggle: (done: boolean) => void;
  onEdit: (title: string) => void;
  onDelete: () => void;
}

export function TaskItem({ task, color, onToggle, onEdit, onDelete }: TaskItemProps) {
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(task.title);
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: task.id,
  });

  const saveDraft = () => {
    const trimmed = draft.trim();
    setEditing(false);
    if (trimmed && trimmed !== task.title) onEdit(trimmed);
    else setDraft(task.title);
  };

  return (
    <li
      ref={setNodeRef}
      style={{ transform: CSS.Transform.toString(transform), transition }}
      className={cn(
        'group flex items-center gap-2 rounded-md border border-border bg-card px-3 py-2 text-sm',
        isDragging && 'z-10 opacity-70 shadow-lg',
      )}
    >
      <button
        type="button"
        aria-label="Arrastar para reordenar"
        className="cursor-grab touch-none text-muted-foreground/50 hover:text-muted-foreground active:cursor-grabbing"
        {...attributes}
        {...listeners}
      >
        <svg width="10" height="16" viewBox="0 0 10 16" fill="currentColor" aria-hidden>
          <circle cx="2.5" cy="3" r="1.5" />
          <circle cx="7.5" cy="3" r="1.5" />
          <circle cx="2.5" cy="8" r="1.5" />
          <circle cx="7.5" cy="8" r="1.5" />
          <circle cx="2.5" cy="13" r="1.5" />
          <circle cx="7.5" cy="13" r="1.5" />
        </svg>
      </button>

      <input
        type="checkbox"
        checked={task.done}
        onChange={(e) => onToggle(e.target.checked)}
        aria-label={task.done ? 'Desmarcar tarefa' : 'Concluir tarefa'}
        className={cn('h-4 w-4 shrink-0 cursor-pointer', colorClasses(color).checkbox)}
      />

      {editing ? (
        <input
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          onBlur={saveDraft}
          onKeyDown={(e) => {
            if (e.key === 'Enter') saveDraft();
            if (e.key === 'Escape') {
              setDraft(task.title);
              setEditing(false);
            }
          }}
          maxLength={500}
          autoFocus
          className="flex-1 bg-transparent outline-none"
        />
      ) : (
        <span
          onDoubleClick={() => {
            setDraft(task.title);
            setEditing(true);
          }}
          className={cn('flex-1 break-words', task.done && 'text-muted-foreground line-through')}
        >
          {task.title}
        </span>
      )}

      <button
        type="button"
        onClick={onDelete}
        aria-label="Excluir tarefa"
        className="text-muted-foreground/50 opacity-0 transition-opacity hover:text-destructive group-hover:opacity-100 focus-visible:opacity-100"
      >
        ✕
      </button>
    </li>
  );
}
