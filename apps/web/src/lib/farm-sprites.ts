// Sprites da fazenda em emoji (v1). Trocar por sprite sheets pixel-art depois é
// só reimplementar estas funções — o resto do app não muda.
// As classes Tailwind são literais completos (o scanner não vê strings montadas).

const ANIMAL_EMOJI: Record<string, string> = {
  chicken: '🐔',
  rooster: '🐓',
  rabbit: '🐰',
  cow: '🐄',
  pig: '🐷',
  duck: '🦆',
};

export function animalEmoji(type: string): string {
  return ANIMAL_EMOJI[type] ?? '🐔';
}

export function animalLabel(type: string): string {
  const labels: Record<string, string> = {
    chicken: 'Galinha',
    rooster: 'Galo',
    rabbit: 'Coelho',
    cow: 'Vaca',
    pig: 'Porco',
    duck: 'Pato',
  };
  return labels[type] ?? type;
}

// Emoji por estágio de crescimento (0..maxStage). O último = pronta.
const CROP_STAGES: Record<string, string[]> = {
  corn: ['🌱', '🌿', '🌾', '🌽'],
  wheat: ['🌱', '🌿', '🌾', '🌾'],
  apple: ['🌱', '🌿', '🌳', '🌳', '🍎'],
};

export function cropEmoji(type: string, stage: number): string {
  const stages = CROP_STAGES[type];
  if (!stages) return '🌱';
  return stages[Math.min(stage, stages.length - 1)];
}

export function cropLabel(type: string): string {
  const labels: Record<string, string> = { corn: 'Milho', wheat: 'Trigo', apple: 'Maçã' };
  return labels[type] ?? type;
}

const ITEM_EMOJI: Record<string, string> = {
  egg: '🥚',
  milk: '🥛',
  corn: '🌽',
  wheat: '🌾',
  apple: '🍎',
};

export function itemEmoji(item: string): string {
  return ITEM_EMOJI[item] ?? '📦';
}

export function itemLabel(item: string): string {
  const labels: Record<string, string> = {
    egg: 'Ovos',
    milk: 'Leite',
    corn: 'Milho',
    wheat: 'Trigo',
    apple: 'Maçãs',
  };
  return labels[item] ?? item;
}

// Cor da barra de conforto por faixa (literais p/ o Tailwind).
export function comfortBarClass(comfort: number): string {
  if (comfort >= 60) return 'bg-green-500';
  if (comfort >= 30) return 'bg-amber-400';
  return 'bg-red-500';
}
