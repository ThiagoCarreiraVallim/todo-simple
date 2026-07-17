'use client';

import { LIST_COLORS, colorClasses, type ListColor } from '@/lib/colors';
import { cn } from '@/lib/utils';

interface ColorPickerProps {
  value: string;
  onChange: (color: ListColor) => void;
}

export function ColorPicker({ value, onChange }: ColorPickerProps) {
  return (
    <div className="flex flex-wrap gap-2">
      {LIST_COLORS.map((color) => (
        <button
          key={color}
          type="button"
          aria-label={`Cor ${color}`}
          onClick={() => onChange(color)}
          className={cn(
            'h-6 w-6 rounded-full transition-transform hover:scale-110',
            colorClasses(color).dot,
            value === color && 'ring-2 ring-foreground ring-offset-2 ring-offset-background',
          )}
        />
      ))}
    </div>
  );
}
