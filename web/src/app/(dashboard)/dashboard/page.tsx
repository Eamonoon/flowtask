'use client';

import Link from 'next/link';
import { useAuthStore } from '@/stores/auth-store';
import { useRouter } from 'next/navigation';
import { useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import api from '@/lib/api';
import type { ApiResponse, DashboardStats, Task } from '@/types/api';
import { StatsCard } from '@/components/dashboard/stats-card';
import { StudyTimeChart } from '@/components/dashboard/study-time-chart';
import { CategoryChart } from '@/components/dashboard/category-chart';
import { CompletionRateChart } from '@/components/dashboard/completion-rate-chart';
import { RecentActivity } from '@/components/dashboard/recent-activity';
import { UpcomingDeadlines } from '@/components/dashboard/upcoming-deadlines';

export default function DashboardPage() {
  const { user, isAuthenticated } = useAuthStore();
  const router = useRouter();

  useEffect(() => {
    if (!isAuthenticated) {
      router.push('/login');
    }
  }, [isAuthenticated, router]);

  const { data: stats, isLoading } = useQuery({
    queryKey: ['dashboard', 'stats'],
    queryFn: async () => {
      const { data } = await api.get<ApiResponse<DashboardStats>>('/dashboard/stats');
      return data.data;
    },
    enabled: isAuthenticated,
  });

  if (!isAuthenticated) return null;

  const completionRate = stats
    ? `${Math.round(stats.overall.completion_rate * 100)}%`
    : '-';

  return (
    <div className="min-h-screen bg-background">
      <nav className="border-b px-6 py-4">
        <div className="max-w-7xl mx-auto flex justify-between items-center">
          <h1 className="text-xl font-bold">FlowTask</h1>
          <div className="flex gap-4 items-center">
            <Link href="/dashboard" className="font-semibold text-primary">
              Dashboard
            </Link>
            <Link href="/tasks" className="hover:text-primary">
              任务
            </Link>
            <Link href="/goals" className="hover:text-primary">
              学习目标
            </Link>
            <Link href="/chat" className="hover:text-primary">
              AI 助手
            </Link>
          </div>
        </div>
      </nav>

      <main className="max-w-7xl mx-auto p-6">
        <h2 className="text-2xl font-bold mb-6">
          欢迎回来，{user?.display_name}
        </h2>

        {/* 统计卡片 */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
          <StatsCard
            title="今日任务"
            value={isLoading ? '-' : stats?.today_tasks.total ?? 0}
            subtitle="个任务"
          />
          <StatsCard
            title="已完成"
            value={isLoading ? '-' : stats?.today_tasks.completed ?? 0}
            subtitle="今日完成"
          />
          <StatsCard
            title="完成率"
            value={isLoading ? '-' : completionRate}
            subtitle={`共 ${stats?.overall.total_tasks ?? 0} 个任务`}
          />
          <StatsCard
            title="今日学习"
            value={isLoading ? '-' : stats?.study_time.today_minutes ?? 0}
            subtitle="分钟"
          />
        </div>

        {/* 图表区域 */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
          <StudyTimeChart />
          <CategoryChart />
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
          <CompletionRateChart />
          <UpcomingDeadlines tasks={stats?.upcoming_deadlines ?? []} />
        </div>

        {/* 最近动态 */}
        <RecentActivity activities={stats?.recent_activity ?? []} />
      </main>
    </div>
  );
}
