// Paleta fixa de cores das listas. Os tokens são validados também na API;
// as classes precisam ser literais completos para o Tailwind gerá-las.
export type ListColor =
  'zinc' | 'red' | 'orange' | 'amber' | 'green' | 'teal' | 'blue' | 'violet' | 'pink';

export const LIST_COLORS: ListColor[] = [
  'zinc',
  'red',
  'orange',
  'amber',
  'green',
  'teal',
  'blue',
  'violet',
  'pink',
];

export const DEFAULT_COLOR: ListColor = 'zinc';

interface ColorClasses {
  /** bolinha da aba e do seletor de cor */
  dot: string;
  /** borda inferior da aba ativa */
  tab: string;
  /** acento do cabeçalho da lista */
  accent: string;
  /** checkbox marcado */
  checkbox: string;
}

const COLOR_CLASSES: Record<ListColor, ColorClasses> = {
  zinc: {
    dot: 'bg-zinc-400',
    tab: 'border-zinc-400',
    accent: 'text-zinc-300',
    checkbox: 'accent-zinc-400',
  },
  red: {
    dot: 'bg-red-500',
    tab: 'border-red-500',
    accent: 'text-red-400',
    checkbox: 'accent-red-500',
  },
  orange: {
    dot: 'bg-orange-500',
    tab: 'border-orange-500',
    accent: 'text-orange-400',
    checkbox: 'accent-orange-500',
  },
  amber: {
    dot: 'bg-amber-400',
    tab: 'border-amber-400',
    accent: 'text-amber-300',
    checkbox: 'accent-amber-400',
  },
  green: {
    dot: 'bg-green-500',
    tab: 'border-green-500',
    accent: 'text-green-400',
    checkbox: 'accent-green-500',
  },
  teal: {
    dot: 'bg-teal-500',
    tab: 'border-teal-500',
    accent: 'text-teal-400',
    checkbox: 'accent-teal-500',
  },
  blue: {
    dot: 'bg-blue-500',
    tab: 'border-blue-500',
    accent: 'text-blue-400',
    checkbox: 'accent-blue-500',
  },
  violet: {
    dot: 'bg-violet-500',
    tab: 'border-violet-500',
    accent: 'text-violet-400',
    checkbox: 'accent-violet-500',
  },
  pink: {
    dot: 'bg-pink-500',
    tab: 'border-pink-500',
    accent: 'text-pink-400',
    checkbox: 'accent-pink-500',
  },
};

export function colorClasses(color: string): ColorClasses {
  return COLOR_CLASSES[color as ListColor] ?? COLOR_CLASSES[DEFAULT_COLOR];
}
