import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type Role = 'secretary' | 'teacher';

export interface User {
  id: number;
  name: string;
  email: string;
}

export interface AuthState {
  token: string | null;
  role: Role | null;
  user: User | null;
  isAuthenticated: boolean;

  // Actions
  setAuth: (token: string, role: Role, user: User) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      role: null,
      user: null,
      isAuthenticated: false,

      setAuth: (token: string, role: Role, user: User) => {
        set({
          token,
          role,
          user,
          isAuthenticated: true,
        });
      },

      logout: () => {
        set({
          token: null,
          role: null,
          user: null,
          isAuthenticated: false,
        });
      },
    }),
    {
      name: 'edutrack-auth',
      partialize: (state) => ({
        token: state.token,
        role: state.role,
        user: state.user,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);
