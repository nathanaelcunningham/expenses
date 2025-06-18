import React from 'react';
import { useAuth } from '../contexts/AuthContext';
import { Navigate } from '@tanstack/react-router';

interface ProtectedRouteProps {
  children: React.ReactNode;
}

export function ProtectedRoute({ children }: ProtectedRouteProps) {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" />;
  }

  return <>{children}</>;
}

// Hook for easier use in route definitions
export function useRequireAuth() {
  const { isAuthenticated, isLoading } = useAuth();
  
  if (isLoading) {
    return { isLoading: true, isAuthenticated: false };
  }
  
  return { isLoading: false, isAuthenticated };
}