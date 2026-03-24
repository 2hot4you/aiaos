import { create } from 'zustand';
import api from '@/lib/api';

interface User {
  id: string;
  username: string;
  display_name: string;
  role: string;
  enabled: boolean;
}

interface AuthState {
  user: User | null;
  token: string | null;
  loading: boolean;
  login: (username: string, password: string) => Promise<void>;
  logout: () => void;
  fetchMe: () => Promise<void>;
  init: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  token: null,
  loading: true,

  init: () => {
    if (typeof window !== 'undefined') {
      const token = localStorage.getItem('token');
      if (token) {
        set({ token });
      } else {
        set({ loading: false });
      }
    }
  },

  login: async (username: string, password: string) => {
    const res: any = await api.post('/api/v1/auth/login', { username, password });
    const { token, user } = res.data;
    localStorage.setItem('token', token);
    set({ token, user, loading: false });
  },

  logout: () => {
    localStorage.removeItem('token');
    set({ user: null, token: null });
    window.location.href = '/login';
  },

  fetchMe: async () => {
    try {
      const res: any = await api.get('/api/v1/auth/me');
      set({ user: res.data, loading: false });
    } catch {
      localStorage.removeItem('token');
      set({ user: null, token: null, loading: false });
    }
  },
}));
