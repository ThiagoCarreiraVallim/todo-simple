'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { ColorPicker } from '@/components/color-picker';
import { useCreateList } from '@/hooks/use-list';
import { DEFAULT_COLOR, type ListColor } from '@/lib/colors';

interface NewListFormProps {
  onCreated: (slug: string) => void;
}

export function NewListForm({ onCreated }: NewListFormProps) {
  const [name, setName] = useState('');
  const [color, setColor] = useState<ListColor>(DEFAULT_COLOR);
  const createList = useCreateList(onCreated);

  return (
    <Card>
      <CardHeader>
        <CardTitle>Nova lista</CardTitle>
        <CardDescription>
          Guarde o link — ele é a chave da sua lista. Quem tem o link pode ver e editar.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form
          className="space-y-4"
          onSubmit={(e) => {
            e.preventDefault();
            if (name.trim()) createList.mutate({ name: name.trim(), color });
          }}
        >
          <Input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Nome da lista"
            autoFocus
            maxLength={120}
          />
          <ColorPicker value={color} onChange={setColor} />
          <Button type="submit" disabled={createList.isPending || !name.trim()}>
            Criar lista
          </Button>
          {createList.isError && (
            <p className="text-sm text-destructive">Não foi possível criar a lista.</p>
          )}
        </form>
      </CardContent>
    </Card>
  );
}
