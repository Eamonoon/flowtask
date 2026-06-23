'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/stores/auth-store';
import { Providers } from '@/components/providers';

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const router = useRouter();
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // 等待 Zustand persist 完成恢复
    const unsub = useAuthStore.persist.onFinishHydration(() => {
      setIsLoading(false);
    });

    // 如果已经完成恢复，立即设置加载完成
    if (useAuthStore.persist.hasHydrated()) {
      setIsLoading(false);
    }

    return unsub;
  }, []);

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.push('/login');
    }
  }, [isAuthenticated, isLoading, router]);

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-muted-foreground">加载中...</div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return null;
  }

  return (
    <div className="min-h-screen bg-background">
      <Providers>{children}</Providers>
    </div>
  );
}
