import React from 'react';
import ReactDOM from 'react-dom/client';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Toaster } from 'react-hot-toast';
import App from './App';
import './index.css';

// Create a client
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 3,
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
      staleTime: 1000 * 60 * 5, // 5 minutes
      cacheTime: 1000 * 60 * 10, // 10 minutes
    },
    mutations: {
      retry: 1,
    },
  },
});

// Error boundary component
class ErrorBoundary extends React.Component<
  { children: React.ReactNode },
  { hasError: boolean; error?: Error }
> {
  constructor(props: { children: React.ReactNode }) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): { hasError: boolean; error: Error } {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('React Error Boundary caught an error:', error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="min-h-screen bg-secondary-900 text-secondary-100 flex items-center justify-center">
          <div className="text-center p-8">
            <h1 className="text-4xl font-bold text-error-500 mb-4">
              🚨 Something went wrong
            </h1>
            <p className="text-lg text-secondary-400 mb-6">
              The StackAgent GUI encountered an unexpected error
            </p>
            <div className="bg-secondary-800 border border-secondary-700 rounded-lg p-4 mb-6 text-left">
              <h2 className="text-sm font-semibold text-secondary-300 mb-2">Error Details:</h2>
              <pre className="text-xs text-error-400 overflow-auto">
                {this.state.error?.message}
              </pre>
              <pre className="text-xs text-secondary-500 mt-2 overflow-auto">
                {this.state.error?.stack}
              </pre>
            </div>
            <button
              onClick={() => window.location.reload()}
              className="px-6 py-2 bg-primary-600 hover:bg-primary-700 text-white rounded-lg transition-colors"
            >
              Reload Application
            </button>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

// Toast configuration
const toastOptions = {
  duration: 4000,
  position: 'top-right' as const,
  style: {
    background: '#1e293b',
    color: '#f1f5f9',
    border: '1px solid #334155',
  },
  success: {
    iconTheme: {
      primary: '#22c55e',
      secondary: '#f1f5f9',
    },
  },
  error: {
    iconTheme: {
      primary: '#ef4444',
      secondary: '#f1f5f9',
    },
  },
};

// Render the app
ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <ErrorBoundary>
      <QueryClientProvider client={queryClient}>
        <App />
        <Toaster toastOptions={toastOptions} />
      </QueryClientProvider>
    </ErrorBoundary>
  </React.StrictMode>
); 