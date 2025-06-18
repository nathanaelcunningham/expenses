import { createFileRoute } from '@tanstack/react-router';
import { AuthLayout } from '../components/auth/AuthLayout';
import { RegisterForm } from '../components/auth/RegisterForm';
import { useAuth } from '../contexts/AuthContext';

export const Route = createFileRoute('/register')({
  component: Register,
});

function Register() {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  if (isAuthenticated) {
    window.location.href = '/';
    return null;
  }

  return (
    <AuthLayout
      title="Create your account"
      subtitle="Join us to start tracking your expenses."
    >
      <RegisterForm />
    </AuthLayout>
  );
}