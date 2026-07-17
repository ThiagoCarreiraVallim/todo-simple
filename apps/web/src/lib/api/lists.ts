import { apiFetch } from '@/lib/api';

export interface List {
  slug: string;
  name: string;
  color: string;
  createdAt: string;
}

export interface Task {
  id: string;
  title: string;
  done: boolean;
  position: number;
  createdAt: string;
  completedAt: string | null;
}

export interface ListWithTasks extends List {
  tasks: Task[];
}

export const listsApi = {
  create: (input: { name: string; color?: string }) =>
    apiFetch<List>('/api/lists', { method: 'POST', body: JSON.stringify(input) }),

  get: (slug: string) => apiFetch<ListWithTasks>(`/api/lists/${slug}`),

  update: (slug: string, input: { name?: string; color?: string }) =>
    apiFetch<List>(`/api/lists/${slug}`, { method: 'PATCH', body: JSON.stringify(input) }),

  remove: (slug: string) => apiFetch<void>(`/api/lists/${slug}`, { method: 'DELETE' }),

  addTask: (slug: string, title: string) =>
    apiFetch<Task>(`/api/lists/${slug}/tasks`, { method: 'POST', body: JSON.stringify({ title }) }),

  updateTask: (slug: string, taskId: string, input: { title?: string; done?: boolean }) =>
    apiFetch<Task>(`/api/lists/${slug}/tasks/${taskId}`, {
      method: 'PATCH',
      body: JSON.stringify(input),
    }),

  removeTask: (slug: string, taskId: string) =>
    apiFetch<void>(`/api/lists/${slug}/tasks/${taskId}`, { method: 'DELETE' }),

  reorderTasks: (slug: string, taskIds: string[]) =>
    apiFetch<void>(`/api/lists/${slug}/tasks/order`, {
      method: 'PUT',
      body: JSON.stringify({ taskIds }),
    }),
};
