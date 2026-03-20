import { createContext, useContext, useState, useCallback, type ReactNode } from 'react';
import { api } from '../api/client';

interface User {
  id: number;
  username: string;
  email: string;
  admin: boolean;
  first_name: string;
  last_name: string;
}

interface LoginResponse {
  token: string;
  user: User;
}

interface AuthState {
  token: string | null;
  user: User | null;
}

interface AuthContextValue extends AuthState {
  login: (login: string, password: string) => Promise<void>;
  logout: () => void;
  isAuthenticated: boolean;
}

const AuthContext = createContext<AuthContextValue | null>(null);

function loadInitialState(): AuthState {
  const token = localStorage.getItem('token');
  const userJson = localStorage.getItem('user');
  if (token && userJson) {
    try {
      return { token, user: JSON.parse(userJson) as User };
    } catch {
      return { token: null, user: null };
    }
  }
  return { token: null, user: null };
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>(loadInitialState);

  const login = useCallback(async (loginValue: string, password: string) => {
    const resp = await api.post<LoginResponse>('/auth/login', {
      login: loginValue,
      password,
    });
    localStorage.setItem('token', resp.token);
    localStorage.setItem('user', JSON.stringify(resp.user));
    setState({ token: resp.token, user: resp.user });
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    setState({ token: null, user: null });
  }, []);

  return (
    <AuthContext.Provider
      value={{ ...state, login, logout, isAuthenticated: !!state.token }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
