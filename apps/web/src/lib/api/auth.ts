import { apiFetch } from '@/lib/api';

export interface AuthUser {
  id: string;
  username: string;
}

export const authApi = {
  // Login sem senha: get-or-create do usuário (e garante a fazenda no servidor).
  login: (username: string) =>
    apiFetch<AuthUser>('/api/login', { method: 'POST', body: JSON.stringify({ username }) }),
};
