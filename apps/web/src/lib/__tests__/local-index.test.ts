import { beforeEach, describe, expect, it } from 'vitest';
import { getKnownLists, removeKnownList, upsertKnownList } from '../local-index';

describe('local-index', () => {
  beforeEach(() => {
    window.localStorage.clear();
  });

  it('starts empty', () => {
    expect(getKnownLists()).toEqual([]);
  });

  it('upserts and orders by most recently opened', () => {
    upsertKnownList({ slug: 'aaaaaaaaaaaaaaaaaaaaa', name: 'Antiga', color: 'zinc' });
    upsertKnownList({ slug: 'bbbbbbbbbbbbbbbbbbbbb', name: 'Recente', color: 'blue' });

    const lists = getKnownLists();
    expect(lists).toHaveLength(2);
    expect(lists[0].name).toBe('Recente');
    expect(lists[1].name).toBe('Antiga');
  });

  it('updates an existing entry instead of duplicating', () => {
    upsertKnownList({ slug: 'aaaaaaaaaaaaaaaaaaaaa', name: 'Original', color: 'zinc' });
    upsertKnownList({ slug: 'aaaaaaaaaaaaaaaaaaaaa', name: 'Renomeada', color: 'pink' });

    const lists = getKnownLists();
    expect(lists).toHaveLength(1);
    expect(lists[0].name).toBe('Renomeada');
    expect(lists[0].color).toBe('pink');
  });

  it('removes entries by slug', () => {
    upsertKnownList({ slug: 'aaaaaaaaaaaaaaaaaaaaa', name: 'Uma', color: 'zinc' });
    removeKnownList('aaaaaaaaaaaaaaaaaaaaa');
    expect(getKnownLists()).toEqual([]);
  });

  it('ignores corrupted storage contents', () => {
    window.localStorage.setItem('todo.lists.v1', 'not json');
    expect(getKnownLists()).toEqual([]);
  });
});
