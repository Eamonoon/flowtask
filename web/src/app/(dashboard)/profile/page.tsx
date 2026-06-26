'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useForm, Controller } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { useAuthStore } from '@/stores/auth-store';
import { Button } from '@/components/ui/button';
import api from '@/lib/api';

const profileSchema = z.object({
  display_name: z.string().min(2, '显示名称至少 2 个字符').max(100, '显示名称最多 100 个字符'),
  avatar_url: z.string().url('请输入有效的 URL').optional().or(z.literal('')),
  preferences: z.object({
    theme: z.enum(['light', 'dark']).optional(),
    language: z.string().optional(),
    learning_style: z.string().optional(),
    weekly_study_hours: z.number().min(0, '不能为负数').max(168, '一周最多 168 小时').optional(),
    preferred_session_minutes: z.number().min(1, '至少 1 分钟').max(480, '最长 480 分钟').optional(),
  }),
});

type ProfileFormValues = z.infer<typeof profileSchema>;

export default function ProfilePage() {
  const { user, isAuthenticated, logout, setUser } = useAuthStore();
  const router = useRouter();

  useEffect(() => {
    if (!isAuthenticated) router.push('/login');
  }, [isAuthenticated, router]);

  const {
    register,
    handleSubmit,
    control,
    formState: { errors, isSubmitting },
  } = useForm<ProfileFormValues>({
    resolver: zodResolver(profileSchema),
    defaultValues: {
      display_name: user?.display_name || '',
      avatar_url: user?.avatar_url || '',
      preferences: {
        theme: (user?.preferences?.theme as 'light' | 'dark') || 'light',
        language: user?.preferences?.language || 'zh-CN',
        learning_style: user?.preferences?.learning_style || '',
        weekly_study_hours: user?.preferences?.weekly_study_hours || 10,
        preferred_session_minutes: user?.preferences?.preferred_session_minutes || 30,
      },
    },
  });

  const onSubmit = async (values: ProfileFormValues) => {
    try {
      const payload = {
        display_name: values.display_name,
        avatar_url: values.avatar_url || null,
        preferences: values.preferences,
      };
      const { data } = await api.put('/user/profile', payload);
      if (data.data) {
        setUser(data.data);
      }
    } catch {
      // error handling via interceptor
    }
  };

  if (!isAuthenticated) return null;

  return (
    <div className="min-h-screen bg-background">
      <nav className="border-b px-6 py-4">
        <div className="max-w-7xl mx-auto flex justify-between items-center">
          <h1 className="text-xl font-bold">FlowTask</h1>
          <div className="flex gap-4 items-center">
            <Link href="/dashboard">Dashboard</Link>
            <Link href="/tasks">任务</Link>
            <Link href="/goals">学习目标</Link>
            <Link href="/chat">AI 助手</Link>
          </div>
        </div>
      </nav>

      <main className="max-w-2xl mx-auto p-6">
        <h2 className="text-2xl font-bold mb-6">个人资料</h2>

        <form onSubmit={handleSubmit(onSubmit)} className="border rounded-lg p-6 space-y-5">
          {/* 邮箱 (只读) */}
          <div className="space-y-1.5">
            <label className="text-sm font-medium text-foreground">邮箱</label>
            <p className="text-muted-foreground text-sm">{user?.email}</p>
          </div>

          {/* 显示名称 */}
          <div className="space-y-1.5">
            <label htmlFor="display_name" className="text-sm font-medium text-foreground">
              显示名称 <span className="text-destructive">*</span>
            </label>
            <input
              id="display_name"
              type="text"
              className="flex h-9 w-full rounded-lg border border-input bg-background px-3 py-1 text-sm shadow-sm transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              {...register('display_name')}
            />
            {errors.display_name && (
              <p className="text-xs text-destructive">{errors.display_name.message}</p>
            )}
          </div>

          {/* 头像 URL */}
          <div className="space-y-1.5">
            <label htmlFor="avatar_url" className="text-sm font-medium text-foreground">
              头像 URL
            </label>
            <input
              id="avatar_url"
              type="text"
              placeholder="https://..."
              className="flex h-9 w-full rounded-lg border border-input bg-background px-3 py-1 text-sm shadow-sm transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              {...register('avatar_url')}
            />
            {errors.avatar_url && (
              <p className="text-xs text-destructive">{errors.avatar_url.message}</p>
            )}
          </div>

          {/* 偏好设置 */}
          <div className="border-t pt-5 space-y-4">
            <h3 className="text-lg font-semibold">偏好设置</h3>

            {/* 主题 */}
            <div className="space-y-1.5">
              <label htmlFor="pref-theme" className="text-sm font-medium text-foreground">
                主题
              </label>
              <Controller
                name="preferences.theme"
                control={control}
                render={({ field }) => (
                  <select
                    id="pref-theme"
                    className="flex h-9 w-full rounded-lg border border-input bg-background px-3 py-1 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                    value={field.value ?? 'light'}
                    onChange={field.onChange}
                    onBlur={field.onBlur}
                  >
                    <option value="light">浅色</option>
                    <option value="dark">深色</option>
                  </select>
                )}
              />
            </div>

            {/* 语言 */}
            <div className="space-y-1.5">
              <label htmlFor="pref-language" className="text-sm font-medium text-foreground">
                语言
              </label>
              <Controller
                name="preferences.language"
                control={control}
                render={({ field }) => (
                  <select
                    id="pref-language"
                    className="flex h-9 w-full rounded-lg border border-input bg-background px-3 py-1 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                    value={field.value ?? 'zh-CN'}
                    onChange={field.onChange}
                    onBlur={field.onBlur}
                  >
                    <option value="zh-CN">简体中文</option>
                    <option value="en">English</option>
                    <option value="ja">日本語</option>
                  </select>
                )}
              />
            </div>

            {/* 学习风格 */}
            <div className="space-y-1.5">
              <label htmlFor="pref-learning-style" className="text-sm font-medium text-foreground">
                学习风格
              </label>
              <Controller
                name="preferences.learning_style"
                control={control}
                render={({ field }) => (
                  <select
                    id="pref-learning-style"
                    className="flex h-9 w-full rounded-lg border border-input bg-background px-3 py-1 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                    value={field.value ?? ''}
                    onChange={field.onChange}
                    onBlur={field.onBlur}
                  >
                    <option value="">未选择</option>
                    <option value="visual">视觉型</option>
                    <option value="auditory">听觉型</option>
                    <option value="reading">阅读型</option>
                    <option value="kinesthetic">动手型</option>
                  </select>
                )}
              />
            </div>

            {/* 每周学习时长 */}
            <div className="space-y-1.5">
              <label htmlFor="pref-weekly-hours" className="text-sm font-medium text-foreground">
                每周学习目标（小时）
              </label>
              <input
                id="pref-weekly-hours"
                type="number"
                min={0}
                max={168}
                className="flex h-9 w-full rounded-lg border border-input bg-background px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                {...register('preferences.weekly_study_hours')}
              />
              {errors.preferences?.weekly_study_hours && (
                <p className="text-xs text-destructive">{errors.preferences.weekly_study_hours.message}</p>
              )}
            </div>

            {/* 每次学习时长 */}
            <div className="space-y-1.5">
              <label htmlFor="pref-session-mins" className="text-sm font-medium text-foreground">
                默认学习时长（分钟）
              </label>
              <input
                id="pref-session-mins"
                type="number"
                min={1}
                max={480}
                className="flex h-9 w-full rounded-lg border border-input bg-background px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                {...register('preferences.preferred_session_minutes')}
              />
              {errors.preferences?.preferred_session_minutes && (
                <p className="text-xs text-destructive">{errors.preferences.preferred_session_minutes.message}</p>
              )}
            </div>
          </div>

          {/* 操作按钮 */}
          <div className="flex gap-4 pt-4">
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? '保存中...' : '保存'}
            </Button>
            <Button
              type="button"
              variant="destructive"
              onClick={() => {
                logout();
                router.push('/login');
              }}
            >
              退出登录
            </Button>
          </div>
        </form>
      </main>
    </div>
  );
}
