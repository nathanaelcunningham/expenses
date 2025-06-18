import React, { createContext, useContext, useEffect, useState } from 'react';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import { useQueryClient } from '@tanstack/react-query';
import { login, logout, refreshSession, validateSession, register } from '../gen/auth/v1/auth-AuthService_connectquery';
import type { User, Session, AuthError } from '../gen/auth/v1/auth_pb';

interface AuthState {
  user: User | null;
  session: Session | null;
  isLoading: boolean;
  isAuthenticated: boolean;
}

export interface AuthContextType extends AuthState {
  login: (email: string, password: string) => Promise<{ success: boolean; error?: AuthError }>;
  register: (email: string, name: string, password: string) => Promise<{ success: boolean; error?: AuthError }>;
  logout: () => Promise<void>;
  refreshSession: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

const SESSION_STORAGE_KEY = 'auth_session_id';

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [authState, setAuthState] = useState<AuthState>({
    user: null,
    session: null,
    isLoading: true,
    isAuthenticated: false,
  });

  const queryClient = useQueryClient();

  // Get stored session ID
  const getStoredSessionId = () => {
    try {
      return localStorage.getItem(SESSION_STORAGE_KEY);
    } catch {
      return null;
    }
  };

  // Store session ID
  const storeSessionId = (sessionId: string) => {
    try {
      localStorage.setItem(SESSION_STORAGE_KEY, sessionId);
    } catch {
      // Handle localStorage not available
    }
  };

  // Clear stored session ID
  const clearStoredSessionId = () => {
    try {
      localStorage.removeItem(SESSION_STORAGE_KEY);
    } catch {
      // Handle localStorage not available
    }
  };

  // Validate session query
  const sessionValidationQuery = useQuery(
    validateSession,
    { sessionId: getStoredSessionId() || '' },
    {
      enabled: !!getStoredSessionId(),
      retry: false,
      refetchOnWindowFocus: false,
    }
  );

  // Login mutation
  const loginMutation = useMutation(login);

  // Register mutation
  const registerMutation = useMutation(register);

  // Logout mutation
  const logoutMutation = useMutation(logout);

  // Refresh session mutation
  const refreshSessionMutation = useMutation(refreshSession);

  // Update auth state based on session validation
  useEffect(() => {
    if (sessionValidationQuery.data) {
      const result = sessionValidationQuery.data.result;
      if (result?.valid && result.user && result.session) {
        setAuthState({
          user: result.user,
          session: result.session,
          isLoading: false,
          isAuthenticated: true,
        });
      } else {
        // Invalid session, clear stored data
        clearStoredSessionId();
        setAuthState({
          user: null,
          session: null,
          isLoading: false,
          isAuthenticated: false,
        });
      }
    } else if (sessionValidationQuery.isError || !getStoredSessionId()) {
      // No session or validation failed
      setAuthState({
        user: null,
        session: null,
        isLoading: false,
        isAuthenticated: false,
      });
    }
  }, [sessionValidationQuery.data, sessionValidationQuery.isError]);

  // Auto-refresh session before expiry
  useEffect(() => {
    if (authState.session && authState.isAuthenticated) {
      const expiresAt = new Date(Number(authState.session.expiresAt) * 1000);
      const now = new Date();
      const timeUntilExpiry = expiresAt.getTime() - now.getTime();
      
      // Refresh 5 minutes before expiry
      const refreshTime = Math.max(timeUntilExpiry - 5 * 60 * 1000, 0);
      
      if (refreshTime > 0) {
        const refreshTimer = setTimeout(() => {
          handleRefreshSession();
        }, refreshTime);
        
        return () => clearTimeout(refreshTimer);
      }
    }
  }, [authState.session]);

  const handleLogin = async (email: string, password: string) => {
    try {
      const result = await loginMutation.mutateAsync({ email, password });
      
      if (result.error) {
        return { success: false, error: result.error };
      }
      
      if (result.session && result.user) {
        storeSessionId(result.session.id);
        setAuthState({
          user: result.user,
          session: result.session,
          isLoading: false,
          isAuthenticated: true,
        });
        
        // Invalidate and refetch queries
        queryClient.invalidateQueries();
        
        return { success: true };
      }
      
      return { success: false, error: { code: 'UNKNOWN_ERROR', message: 'Login failed' } as AuthError };
    } catch (error) {
      console.error('Login error:', error);
      return { success: false, error: { code: 'NETWORK_ERROR', message: 'Network error occurred' } as AuthError };
    }
  };

  const handleRegister = async (email: string, name: string, password: string) => {
    try {
      const result = await registerMutation.mutateAsync({ email, name, password });
      
      if (result.error) {
        return { success: false, error: result.error };
      }
      
      if (result.user) {
        // After successful registration, automatically log in
        return await handleLogin(email, password);
      }
      
      return { success: false, error: { code: 'UNKNOWN_ERROR', message: 'Registration failed' } as AuthError };
    } catch (error) {
      console.error('Registration error:', error);
      return { success: false, error: { code: 'NETWORK_ERROR', message: 'Network error occurred' } as AuthError };
    }
  };

  const handleLogout = async () => {
    try {
      const sessionId = authState.session?.id;
      if (sessionId) {
        await logoutMutation.mutateAsync({ sessionId });
      }
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      // Always clear local state regardless of server response
      clearStoredSessionId();
      setAuthState({
        user: null,
        session: null,
        isLoading: false,
        isAuthenticated: false,
      });
      
      // Clear all queries
      queryClient.clear();
    }
  };

  const handleRefreshSession = async () => {
    try {
      const sessionId = authState.session?.id;
      if (sessionId) {
        const result = await refreshSessionMutation.mutateAsync({ sessionId });
        
        if (result.session) {
          setAuthState(prev => ({
            ...prev,
            session: result.session!,
          }));
        }
      }
    } catch (error) {
      console.error('Refresh session error:', error);
      // If refresh fails, logout user
      await handleLogout();
    }
  };

  const contextValue: AuthContextType = {
    ...authState,
    login: handleLogin,
    register: handleRegister,
    logout: handleLogout,
    refreshSession: handleRefreshSession,
  };

  return (
    <AuthContext.Provider value={contextValue}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}