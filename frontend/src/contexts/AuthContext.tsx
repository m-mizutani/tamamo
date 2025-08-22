import { createContext, useContext, useState, useEffect, ReactNode } from 'react';

interface User {
  id: string;
  name: string;
  display_name: string;
  email: string;
  teamId: string;
  teamName: string;
}

interface AuthContextType {
  user: User | null;
  loading: boolean;
  error: string | null;
  login: () => void;
  logout: () => Promise<void>;
  checkAuth: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const checkAuth = async () => {
    try {
      const response = await fetch('/api/auth/check', {
        credentials: 'include',
      });

      if (response.ok) {
        const data = await response.json();
        if (data.authenticated && data.user) {
          setUser(data.user);
        } else {
          setUser(null);
        }
      } else {
        setUser(null);
      }
    } catch (err) {
      console.error('Auth check failed:', err);
      setError('Failed to check authentication status');
      setUser(null);
    } finally {
      setLoading(false);
    }
  };

  const login = () => {
    // Redirect to login endpoint
    window.location.href = '/api/auth/login';
  };

  const logout = async () => {
    try {
      const response = await fetch('/api/auth/logout', {
        method: 'POST',
        credentials: 'include',
      });

      if (response.ok) {
        setUser(null);
        // Redirect to login page
        window.location.href = '/';
      } else {
        throw new Error('Logout failed');
      }
    } catch (err) {
      console.error('Logout failed:', err);
      setError('Failed to logout');
    }
  };

  useEffect(() => {
    checkAuth();
  }, []);

  // Check for auth errors in URL params (from OAuth callback)
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const authError = params.get('error');
    
    if (authError) {
      let errorMessage = 'Authentication failed';
      switch (authError) {
        case 'invalid_state':
          errorMessage = 'Security validation failed. Your authentication session may have expired or been tampered with. Please try logging in again. If this persists, clear your browser cookies and cache.';
          break;
        case 'auth_failed':
          errorMessage = 'Slack authentication failed. This could be due to insufficient permissions, cancelled authorization, or a temporary issue with Slack. Please ensure you have the necessary permissions and try again.';
          break;
        case 'invalid_request':
          errorMessage = 'The authentication request was malformed or missing required parameters. This might be a configuration issue. Please contact your administrator if this problem continues.';
          break;
        case 'invalid_scope':
          errorMessage = 'The requested Slack permissions are invalid or not configured correctly. Please contact your administrator to fix the OAuth scope configuration.';
          break;
        case 'access_denied':
          errorMessage = 'Access was denied during Slack authentication. You may not have permission to use this application with your Slack workspace. Please contact your workspace administrator.';
          break;
        default:
          errorMessage = `Authentication error: ${authError}. Please try logging in again or contact support if the issue persists.`;
          break;
      }
      setError(errorMessage);
      
      // Clear the error param from URL
      params.delete('error');
      const newUrl = params.toString() 
        ? `${window.location.pathname}?${params.toString()}`
        : window.location.pathname;
      window.history.replaceState({}, '', newUrl);
    }
  }, []);

  return (
    <AuthContext.Provider value={{ user, loading, error, login, logout, checkAuth }}>
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