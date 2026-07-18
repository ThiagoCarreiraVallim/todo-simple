'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { authApi } from '@/lib/api/auth';
import { farmApi, type Farm } from '@/lib/api/farm';
import { listsApi } from '@/lib/api/lists';
import { useFarmSession } from '@/lib/farm-session';
import { getKnownLists, upsertKnownList } from '@/lib/local-index';

const farmKey = (userId: string) => ['farm', userId] as const;

export function useFarm(userId: string | null) {
  return useQuery({
    queryKey: farmKey(userId ?? ''),
    queryFn: () => farmApi.get(userId as string),
    enabled: !!userId,
    // Reflete o tempo idle passado ao voltar à aba (override do default global).
    refetchOnWindowFocus: true,
  });
}

// Login = get-or-create do usuário. Ao entrar, vincula as listas conhecidas
// deste aparelho ao usuário e puxa as listas dele (de outros aparelhos) para o
// índice local — depois abre a sessão, o que faz o board reler as listas.
export function useLogin() {
  const { login } = useFarmSession();
  return useMutation({
    mutationFn: (username: string) => authApi.login(username),
    onSuccess: async (user) => {
      try {
        const known = getKnownLists();
        if (known.length)
          await listsApi.claim(
            user.id,
            known.map((k) => k.slug),
          );
        const mine = await listsApi.listsByUser(user.id);
        mine.forEach((l) => upsertKnownList(l));
      } catch {
        // best-effort: a sincronização de listas não deve impedir o login
      }
      login({ userId: user.id, username: user.username, name: user.username });
    },
  });
}

export function useRenameFarm(userId: string) {
  const queryClient = useQueryClient();
  const { setName } = useFarmSession();
  return useMutation({
    mutationFn: (name: string) => farmApi.rename(userId, name),
    onSuccess: (farm) => {
      queryClient.setQueryData(farmKey(userId), farm);
      setName(farm.name);
    },
  });
}

// Mutações da economia: cada endpoint devolve o estado já materializado, então
// só precisamos gravar na cache (moedas, inventário e tudo mais atualizam).
function useFarmStateMutation<TInput>(
  userId: string,
  mutationFn: (input: TInput) => Promise<Farm>,
) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn,
    onSuccess: (farm) => queryClient.setQueryData(farmKey(userId), farm),
  });
}

export function useSell(userId: string) {
  return useFarmStateMutation(userId, ({ item, qty }: { item: string; qty: number }) =>
    farmApi.sell(userId, item, qty),
  );
}

export function useBuy(userId: string) {
  return useFarmStateMutation(userId, (type: string) => farmApi.buy(userId, type));
}

export function usePlant(userId: string) {
  return useFarmStateMutation(userId, (type: string) => farmApi.plant(userId, type));
}

export function useHarvest(userId: string) {
  return useFarmStateMutation(userId, (plotIndex: number) => farmApi.harvest(userId, plotIndex));
}
