import { apiFetch } from '@/lib/api';

export interface Animal {
  type: string;
  comfort: number; // 0..100
  comfortable: boolean;
}

export interface Crop {
  type: string;
  plotIndex: number;
  stage: number;
  ready: boolean;
}

export interface Item {
  item: string;
  qty: number;
}

export interface PriceEntry {
  type: string;
  price: number;
}

export interface Shop {
  animals: PriceEntry[];
  seeds: PriceEntry[];
  sellable: PriceEntry[];
  maxPlots: number;
}

export interface Farm {
  name: string;
  coins: number;
  animals: Animal[];
  crops: Crop[];
  items: Item[];
  shop: Shop;
  createdAt: string;
}

// A fazenda é acessada pelo userId do dono (resolvido no login).
export const farmApi = {
  get: (userId: string) => apiFetch<Farm>(`/api/farm/${userId}`),

  feed: (userId: string) => apiFetch<void>(`/api/farm/${userId}/feed`, { method: 'POST' }),

  rename: (userId: string, name: string) =>
    apiFetch<Farm>(`/api/farm/${userId}`, { method: 'PATCH', body: JSON.stringify({ name }) }),

  sell: (userId: string, item: string, qty: number) =>
    apiFetch<Farm>(`/api/farm/${userId}/sell`, {
      method: 'POST',
      body: JSON.stringify({ item, qty }),
    }),

  buy: (userId: string, type: string) =>
    apiFetch<Farm>(`/api/farm/${userId}/buy`, { method: 'POST', body: JSON.stringify({ type }) }),

  plant: (userId: string, type: string) =>
    apiFetch<Farm>(`/api/farm/${userId}/plant`, { method: 'POST', body: JSON.stringify({ type }) }),

  harvest: (userId: string, plotIndex: number) =>
    apiFetch<Farm>(`/api/farm/${userId}/harvest`, {
      method: 'POST',
      body: JSON.stringify({ plotIndex }),
    }),
};
