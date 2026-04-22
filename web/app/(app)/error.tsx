'use client';

import Link from 'next/link';

export default function AppError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <div style={{ padding: '24px 0', display: 'grid', gap: '12px' }}>
      <h2 style={{ margin: 0 }}>Something went wrong in the console.</h2>
      <p style={{ margin: 0, color: 'var(--muted-foreground)' }}>
        {error.message || 'Unexpected runtime error.'}
      </p>
      <div style={{ display: 'flex', gap: '10px', alignItems: 'center' }}>
        <button
          type="button"
          onClick={reset}
          style={{
            height: '36px',
            padding: '0 12px',
            borderRadius: '10px',
            border: '1px solid rgba(15,23,42,0.15)',
            background: '#fff',
            cursor: 'pointer',
          }}
        >
          Try again
        </button>
        <Link href="/dashboard" className="linkAccent">
          Back to dashboard
        </Link>
      </div>
    </div>
  );
}
