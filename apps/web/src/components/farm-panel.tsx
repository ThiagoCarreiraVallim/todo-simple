'use client';

import { useEffect, useRef, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { useBuy, useFarm, useHarvest, usePlant, useRenameFarm, useSell } from '@/hooks/use-farm';
import { useFarmSession } from '@/lib/farm-session';
import {
  animalEmoji,
  animalLabel,
  comfortBarClass,
  cropEmoji,
  cropLabel,
  itemEmoji,
  itemLabel,
} from '@/lib/farm-sprites';
import { cn } from '@/lib/utils';

// Painel da fazenda no rodapé do board. A fazenda é global (por-pessoa): concluir
// tarefas em qualquer lista a alimenta. Sem login não renderiza nada — o botão
// "Entrar" vive no header.
export function FarmPanel() {
  const { userId, ready } = useFarmSession();
  if (!ready || !userId) return null;
  return <FarmDisplay userId={userId} />;
}

function FarmDisplay({ userId }: { userId: string }) {
  const { username, name: sessionName } = useFarmSession();
  const { data: farm, isLoading } = useFarm(userId);
  const rename = useRenameFarm(userId);
  const sell = useSell(userId);
  const buy = useBuy(userId);
  const plant = usePlant(userId);
  const harvest = useHarvest(userId);

  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState('');

  // Reação: pulo quando o conforto sobe (ao alimentar) ou surgem produtos.
  const [reacting, setReacting] = useState(false);
  const prevSignal = useRef<number | null>(null);
  const comfortSum = (farm?.animals ?? []).reduce((s, a) => s + a.comfort, 0);
  const itemSum = (farm?.items ?? []).reduce((s, i) => s + i.qty, 0);
  const signal = comfortSum + itemSum;
  useEffect(() => {
    if (prevSignal.current !== null && signal > prevSignal.current) {
      setReacting(true);
      const t = setTimeout(() => setReacting(false), 500);
      prevSignal.current = signal;
      return () => clearTimeout(t);
    }
    prevSignal.current = signal;
  }, [signal]);

  if (isLoading || !farm) {
    return (
      <Card>
        <CardContent className="py-4 text-sm text-muted-foreground">
          Carregando fazenda…
        </CardContent>
      </Card>
    );
  }

  const saveRename = () => {
    const trimmed = draft.trim();
    setEditing(false);
    if (trimmed && trimmed !== farm.name) rename.mutate(trimmed);
  };

  const sellPrice = (item: string) => farm.shop.sellable.find((s) => s.type === item)?.price ?? 0;
  const busy =
    sell.isPending || buy.isPending || plant.isPending || harvest.isPending || rename.isPending;
  const freePlots = farm.shop.maxPlots - farm.crops.length;

  return (
    <Card>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between gap-2">
          {editing ? (
            <input
              value={draft}
              onChange={(e) => setDraft(e.target.value)}
              onBlur={saveRename}
              onKeyDown={(e) => {
                if (e.key === 'Enter') saveRename();
                if (e.key === 'Escape') setEditing(false);
              }}
              maxLength={40}
              autoFocus
              className="flex-1 bg-transparent text-base font-semibold outline-none"
            />
          ) : (
            <CardTitle
              className="cursor-text text-base"
              onDoubleClick={() => {
                setDraft(farm.name);
                setEditing(true);
              }}
              title="Clique duas vezes para renomear"
            >
              🚜 {farm.name || sessionName || username}
            </CardTitle>
          )}
          <span className="text-sm font-medium text-amber-400">🪙 {farm.coins}</span>
        </div>
      </CardHeader>

      <CardContent className="space-y-5">
        {/* Animais */}
        <div className="flex flex-wrap gap-4">
          {farm.animals.map((a, i) => (
            <div key={i} className="flex items-center gap-3">
              <span
                className={cn(
                  'text-4xl transition-transform',
                  a.comfortable ? 'animate-bounce' : 'opacity-60 grayscale',
                  reacting && 'scale-125',
                )}
                title={a.comfortable ? 'Feliz!' : 'Com fome — conclua tarefas'}
              >
                {animalEmoji(a.type)}
              </span>
              <div className="min-w-28">
                <div className="mb-1 text-sm font-medium">{animalLabel(a.type)}</div>
                <div className="h-2 w-28 overflow-hidden rounded-full bg-secondary">
                  <div
                    className={cn('h-full rounded-full transition-all', comfortBarClass(a.comfort))}
                    style={{ width: `${a.comfort}%` }}
                  />
                </div>
              </div>
            </div>
          ))}
        </div>

        {/* Canteiros */}
        {farm.crops.length > 0 && (
          <div className="flex flex-wrap gap-2">
            {farm.crops.map((c) => (
              <div
                key={c.plotIndex}
                className="flex items-center gap-1.5 rounded-md border border-border px-2 py-1 text-sm"
              >
                <span className="text-xl">{cropEmoji(c.type, c.stage)}</span>
                {c.ready ? (
                  <Button
                    variant="secondary"
                    size="sm"
                    disabled={busy}
                    onClick={() => harvest.mutate(c.plotIndex)}
                  >
                    Colher
                  </Button>
                ) : (
                  <span className="text-muted-foreground">{cropLabel(c.type)}</span>
                )}
              </div>
            ))}
          </div>
        )}

        {/* Inventário + venda */}
        {farm.items.length > 0 && (
          <div className="space-y-2">
            <div className="text-xs font-semibold uppercase text-muted-foreground">Inventário</div>
            <div className="flex flex-wrap gap-2">
              {farm.items.map((it) => (
                <div
                  key={it.item}
                  className="flex items-center gap-2 rounded-md border border-border px-2 py-1 text-sm"
                >
                  <span className="text-lg">{itemEmoji(it.item)}</span>
                  <span>
                    {it.qty} {itemLabel(it.item)}
                  </span>
                  {sellPrice(it.item) > 0 && (
                    <Button
                      variant="outline"
                      size="sm"
                      disabled={busy}
                      onClick={() => sell.mutate({ item: it.item, qty: it.qty })}
                      title={`Vender por ${sellPrice(it.item)} 🪙 cada`}
                    >
                      Vender · {sellPrice(it.item) * it.qty} 🪙
                    </Button>
                  )}
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Loja */}
        <div className="space-y-3 border-t border-border pt-4">
          <div>
            <div className="mb-2 text-xs font-semibold uppercase text-muted-foreground">
              Loja de animais
            </div>
            <div className="flex flex-wrap gap-2">
              {farm.shop.animals.map((a) => (
                <Button
                  key={a.type}
                  variant="outline"
                  size="sm"
                  disabled={busy || farm.coins < a.price}
                  onClick={() => buy.mutate(a.type)}
                  title={`Comprar ${animalLabel(a.type)}`}
                >
                  {animalEmoji(a.type)} {animalLabel(a.type)} · {a.price} 🪙
                </Button>
              ))}
            </div>
          </div>

          <div>
            <div className="mb-2 text-xs font-semibold uppercase text-muted-foreground">
              Plantar {freePlots > 0 ? `(${freePlots} canteiros livres)` : '(canteiros cheios)'}
            </div>
            <div className="flex flex-wrap gap-2">
              {farm.shop.seeds.map((s) => (
                <Button
                  key={s.type}
                  variant="outline"
                  size="sm"
                  disabled={busy || freePlots <= 0 || farm.coins < s.price}
                  onClick={() => plant.mutate(s.type)}
                  title={`Plantar ${cropLabel(s.type)}`}
                >
                  {cropEmoji(s.type, 99)} {cropLabel(s.type)} · {s.price} 🪙
                </Button>
              ))}
            </div>
          </div>
        </div>

        <p className="text-xs text-muted-foreground">
          Concluir e criar tarefas alimenta os bichos. Entre com o mesmo nome de usuário em outro
          aparelho para continuar esta fazenda.
        </p>
      </CardContent>
    </Card>
  );
}
