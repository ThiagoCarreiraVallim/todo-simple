'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { ApiError } from '@/lib/api';
import { listsApi, type ListWithTasks, type Task } from '@/lib/api/lists';
import { removeKnownList, upsertKnownList } from '@/lib/local-index';

const listKey = (slug: string) => ['list', slug] as const;

export function useList(slug: string) {
  return useQuery({
    queryKey: listKey(slug),
    queryFn: async () => {
      const list = await listsApi.get(slug);
      upsertKnownList(list);
      return list;
    },
    retry: (failureCount, error) =>
      !(error instanceof ApiError && error.status === 404) && failureCount < 2,
  });
}

export function useCreateList(onCreated: (slug: string) => void) {
  return useMutation({
    mutationFn: listsApi.create,
    onSuccess: (list) => {
      upsertKnownList(list);
      onCreated(list.slug);
    },
  });
}

export function useUpdateList(slug: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: { name?: string; color?: string }) => listsApi.update(slug, input),
    onSuccess: (list) => {
      upsertKnownList(list);
      queryClient.setQueryData<ListWithTasks>(listKey(slug), (old) =>
        old ? { ...old, name: list.name, color: list.color } : old,
      );
    },
  });
}

export function useDeleteList(slug: string, onDeleted: () => void) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => listsApi.remove(slug),
    onSuccess: () => {
      removeKnownList(slug);
      queryClient.removeQueries({ queryKey: listKey(slug) });
      onDeleted();
    },
  });
}

export function useAddTask(slug: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (title: string) => listsApi.addTask(slug, title),
    onSettled: () => queryClient.invalidateQueries({ queryKey: listKey(slug) }),
  });
}

/**
 * Base das mutações otimistas de tarefa: aplica `apply` na cache imediatamente,
 * desfaz em caso de erro e re-sincroniza ao final.
 */
function useOptimisticTaskMutation<TInput>(
  slug: string,
  mutationFn: (input: TInput) => Promise<unknown>,
  apply: (tasks: Task[], input: TInput) => Task[],
) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn,
    onMutate: async (input: TInput) => {
      await queryClient.cancelQueries({ queryKey: listKey(slug) });
      const previous = queryClient.getQueryData<ListWithTasks>(listKey(slug));
      if (previous) {
        queryClient.setQueryData<ListWithTasks>(listKey(slug), {
          ...previous,
          tasks: apply(previous.tasks, input),
        });
      }
      return { previous };
    },
    onError: (_err, _input, context) => {
      if (context?.previous) queryClient.setQueryData(listKey(slug), context.previous);
    },
    onSettled: () => queryClient.invalidateQueries({ queryKey: listKey(slug) }),
  });
}

export function useToggleTask(slug: string) {
  return useOptimisticTaskMutation(
    slug,
    (input: { taskId: string; done: boolean }) =>
      listsApi.updateTask(slug, input.taskId, { done: input.done }),
    (tasks, input) => tasks.map((t) => (t.id === input.taskId ? { ...t, done: input.done } : t)),
  );
}

export function useEditTask(slug: string) {
  return useOptimisticTaskMutation(
    slug,
    (input: { taskId: string; title: string }) =>
      listsApi.updateTask(slug, input.taskId, { title: input.title }),
    (tasks, input) => tasks.map((t) => (t.id === input.taskId ? { ...t, title: input.title } : t)),
  );
}

export function useDeleteTask(slug: string) {
  return useOptimisticTaskMutation(
    slug,
    (taskId: string) => listsApi.removeTask(slug, taskId),
    (tasks, taskId) => tasks.filter((t) => t.id !== taskId),
  );
}

export function useReorderTasks(slug: string) {
  return useOptimisticTaskMutation(
    slug,
    (taskIds: string[]) => listsApi.reorderTasks(slug, taskIds),
    (tasks, taskIds) => {
      const byId = new Map(tasks.map((t) => [t.id, t]));
      const reordered = taskIds.map((id) => byId.get(id)).filter((t): t is Task => t !== undefined);
      return reordered.length === tasks.length ? reordered : tasks;
    },
  );
}
