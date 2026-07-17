'use client';

import { colorClasses } from '@/lib/colors';
import type { KnownList } from '@/lib/local-index';
import { cn } from '@/lib/utils';

interface ListTabsProps {
  lists: KnownList[];
  activeSlug: string | null;
  creating: boolean;
  onSelect: (slug: string) => void;
  onNew: () => void;
}

export function ListTabs({ lists, activeSlug, creating, onSelect, onNew }: ListTabsProps) {
  return (
    <div className="flex flex-wrap items-center gap-1 border-b border-border">
      {lists.map((list) => {
        const active = !creating && list.slug === activeSlug;
        return (
          <button
            key={list.slug}
            type="button"
            onClick={() => onSelect(list.slug)}
            className={cn(
              'flex items-center gap-2 rounded-t-md border-b-2 px-3 py-2 text-sm transition-colors',
              active
                ? cn('font-medium text-foreground', colorClasses(list.color).tab)
                : 'border-transparent text-muted-foreground hover:text-foreground',
            )}
          >
            <span className={cn('h-2 w-2 shrink-0 rounded-full', colorClasses(list.color).dot)} />
            <span className="max-w-40 truncate">{list.name}</span>
          </button>
        );
      })}
      <button
        type="button"
        onClick={onNew}
        aria-label="Nova lista"
        className={cn(
          'rounded-t-md border-b-2 px-3 py-2 text-sm transition-colors',
          creating
            ? 'border-foreground font-medium text-foreground'
            : 'border-transparent text-muted-foreground hover:text-foreground',
        )}
      >
        +
      </button>
    </div>
  );
}
