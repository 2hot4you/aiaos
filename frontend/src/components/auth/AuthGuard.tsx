'use client';

import { useEffect } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { useAuthStore } from '@/stores/authStore';
import { Spin } from 'antd';

export default function AuthGuard({ children }: { children: React.ReactNode }) {
  const { user, token, loading, fetchMe, init } = useAuthStore();
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    init();
  }, [init]);

  useEffect(() => {
    if (token && !user && !loading) {
      fetchMe();
    }
    if (!loading && !token && pathname !== '/login') {
      router.push('/login');
    }
  }, [token, user, loading, pathname, fetchMe, router]);

  if (pathname === '/login') return <>{children}</>;

  if (loading || (!user && token)) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh', background: '#0A0E1A' }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!user) return null;

  return <>{children}</>;
}
