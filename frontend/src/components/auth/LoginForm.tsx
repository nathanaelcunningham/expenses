import { useState } from 'react';
import { Link, useNavigate } from '@tanstack/react-router';
import { useAppForm } from '../../hooks/form';
import { useAuth } from '../../contexts/AuthContext';

export function LoginForm() {
  const navigate = useNavigate();
  const { login } = useAuth();
  const [serverError, setServerError] = useState<string | null>(null);

  const form = useAppForm({
    defaultValues: {
      email: '',
      password: '',
    },
    validators: {
      onBlur: ({ value }) => {
        const errors = {
          fields: {},
        } as {
          fields: Record<string, string>;
        };

        if (!value.email.trim()) {
          errors.fields.email = 'Email is required';
        } else if (!value.email.includes('@')) {
          errors.fields.email = 'Please enter a valid email address';
        }

        if (!value.password.trim()) {
          errors.fields.password = 'Password is required';
        }

        return errors;
      },
    },
    onSubmit: async ({ value }) => {
      setServerError(null);
      
      try {
        const result = await login(value.email, value.password);
        
        if (result.success) {
          navigate({ to: '/' });
        } else if (result.error) {
          setServerError(result.error.message);
        }
      } catch (error) {
        console.error('Login error:', error);
        setServerError('An unexpected error occurred. Please try again.');
      }
    },
  });

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        e.stopPropagation();
        form.handleSubmit();
      }}
      className="space-y-6"
    >
        {serverError && (
          <div className="p-3 text-sm text-red-700 bg-red-100 border border-red-300 rounded-md">
            {serverError}
          </div>
        )}

        <form.Field name="email">
          {(field) => (
            <div>
              <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-2">
                Email Address
              </label>
              <input
                id="email"
                type="email"
                value={field.state.value}
                placeholder="Enter your email"
                onBlur={field.handleBlur}
                onChange={(e) => field.handleChange(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                autoComplete="email"
              />
              {field.state.meta.isTouched && field.state.meta.errors.length > 0 && (
                <div className="text-sm text-red-600 mt-1">
                  {field.state.meta.errors[0]}
                </div>
              )}
            </div>
          )}
        </form.Field>

        <form.Field name="password">
          {(field) => (
            <div>
              <label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-2">
                Password
              </label>
              <input
                id="password"
                type="password"
                value={field.state.value}
                placeholder="Enter your password"
                onBlur={field.handleBlur}
                onChange={(e) => field.handleChange(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                autoComplete="current-password"
              />
              {field.state.meta.isTouched && field.state.meta.errors.length > 0 && (
                <div className="text-sm text-red-600 mt-1">
                  {field.state.meta.errors[0]}
                </div>
              )}
            </div>
          )}
        </form.Field>

        <div>
          <button
            type="submit"
            disabled={form.state.isSubmitting}
            className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {form.state.isSubmitting ? 'Signing in...' : 'Sign in'}
          </button>
        </div>

        <div className="text-center">
          <span className="text-sm text-gray-600">
            Don't have an account?{' '}
            <Link
              to="/register"
              className="font-medium text-blue-600 hover:text-blue-500"
            >
              Sign up
            </Link>
          </span>
        </div>
      </form>
  );
}