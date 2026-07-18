'use client';

import { useMutation, useQuery, useQueryClient, type QueryClient } from '@tanstack/react-query';
import { ApiError } from '@/lib/api';
import { listsApi, type ListWithTasks, type Task } from '@/lib/api/lists';
import { farmApi } from '@/lib/api/farm';
import { useFarmSession } from '@/lib/farm-session';
import { getSession } from '@/lib/session';
import { removeKnownList, upsertKnownList } from '@/lib/local-index';

const listKey = (slug: string) => ['list', slug] as const;

// Alimenta a fazenda (se logado) quando uma tarefa é criada/concluída.
// Best-effort: um erro aqui nunca deve quebrar a ação de tarefa.
async function feedFarm(queryClient: QueryClient) {
  const session = getSession();
  if (!session) return;
  try {
    await farmApi.feed(session.userId);
    queryClient.invalidateQueries({ queryKey: ['farm', session.userId] });
  } catch {
    // silencioso de propósito
  }
}

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
  // Se logado, a lista já nasce vinculada ao usuário (userId).
  const { userId } = useFarmSession();
  return useMutation({
    mutationFn: (input: { name: string; color?: string }) =>
      listsApi.create(userId ? { ...input, userId } : input),
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
    onSuccess: () => feedFarm(queryClient),
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
  onSuccess?: (input: TInput, queryClient: QueryClient) => void,
) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn,
    onSuccess: onSuccess ? (_data, input) => onSuccess(input, queryClient) : undefined,
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
    // alimenta a fazenda só quando conclui (não ao desmarcar)
    (input, queryClient) => {
      if (input.done) feedFarm(queryClient);
    },
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
