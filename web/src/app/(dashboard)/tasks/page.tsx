'use client';

import { useState, useEffect, useCallback, useMemo } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useAuthStore } from '@/stores/auth-store';
import api from '@/lib/api';
import type { Task, LearningGoal, ApiResponse, PaginatedResponse } from '@/types/api';

import { KanbanColumn } from '@/components/task/kanban-column';
import { TaskCreateForm, type TaskFormValues } from '@/components/task/task-create-form';
import { TaskFiltersBar, type TaskFilters } from '@/components/task/task-filters';
import { LabelManager } from '@/components/task/label-manager';
import { SubtaskList } from '@/components/task/subtask-list';
import { useInfiniteTasks } from '@/hooks/use-infinite-tasks';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';

const DEFAULT_FILTERS: TaskFilters = {
  keyword: '',
  statuses: [],
  priorities: [],
  labelIds: [],
  deadlineFrom: '',
  deadlineTo: '',
  sortBy: 'sort_order',
  sortOrder: 'asc',
};

const STATUS_COLUMNS = [
  { status: 'todo' as const, title: '待办' },
  { status: 'doing' as const, title: '进行中' },
  { status: 'done' as const, title: '已完成' },
];

export default function TasksPage() {
  const { isAuthenticated } = useAuthStore();
  const router = useRouter();
  const queryClient = useQueryClient();

  const [filters, setFilters] = useState<TaskFilters>(DEFAULT_FILTERS);
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [showLabelManager, setShowLabelManager] = useState(false);
  const [selectedTask, setSelectedTask] = useState<Task | null>(null);

  useEffect(() => {
    if (!isAuthenticated) router.push('/login');
  }, [isAuthenticated, router]);

  // 使用无限滚动查询
  const {
    data: infiniteData,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    isLoading: isLoadingTasks,
  } = useInfiniteTasks({ filters });

  // 获取标签列表
  const { data: labelsData } = useQuery({
    queryKey: ['labels'],
    queryFn: async () => {
      const { data } = await api.get<ApiResponse<import('@/types/api').Label[]>>('/labels');
      return data.data;
    },
  });

  // 获取学习目标
  const { data: goalsData } = useQuery({
    queryKey: ['learning-goals'],
    queryFn: async () => {
      const { data } = await api.get<ApiResponse<PaginatedResponse<LearningGoal>>>('/learning-goals');
      return data.data?.items ?? [];
    },
  });

  // 合并所有分页数据
  const allTasks: Task[] = useMemo(() => {
    if (!infiniteData?.pages) return [];
    return infiniteData.pages.flatMap((page) => page.items);
  }, [infiniteData]);

  // 按状态分组
  const tasksByStatus = useMemo(() => {
    const grouped: Record<'todo' | 'doing' | 'done', Task[]> = {
      todo: [],
      doing: [],
      done: [],
    };
    allTasks.forEach((task) => {
      if (grouped[task.status]) {
        grouped[task.status].push(task);
      }
    });
    return grouped;
  }, [allTasks]);

  // 创建任务
  const createMutation = useMutation({
    mutationFn: async (values: TaskFormValues) => {
      const payload = {
        title: values.title,
        description: values.description || undefined,
        priority: values.priority,
        deadline: values.deadline || undefined,
        learning_goal_id: values.learning_goal_id || undefined,
      };
      const { data } = await api.post<ApiResponse<Task>>('/tasks', payload);
      return data.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
      setShowCreateDialog(false);
    },
  });

  // 更新任务状态（拖拽）
  const updateStatusMutation = useMutation({
    mutationFn: async ({ taskId, status }: { taskId: string; status: string }) => {
      const { data } = await api.put<ApiResponse<Task>>(`/tasks/${taskId}`, {
        status,
      });
      return data.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
    },
  });

  const handleStatusChange = useCallback(
    (taskId: string, newStatus: 'todo' | 'doing' | 'done') => {
      updateStatusMutation.mutate({ taskId, status: newStatus });
    },
    [updateStatusMutation]
  );

  if (!isAuthenticated) return null;

  return (
    <div className="min-h-screen bg-background">
      {/* 顶部导航 */}
      <nav className="border-b px-6 py-4">
        <div className="max-w-7xl mx-auto flex justify-between items-center">
          <h1 className="text-xl font-bold">FlowTask</h1>
          <div className="flex gap-4 items-center">
            <Link href="/dashboard" className="hover:text-primary">
              Dashboard
            </Link>
            <Link href="/tasks" className="font-semibold text-primary">
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
        {/* 页面标题与操作栏 */}
        <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-6">
          <h2 className="text-2xl font-bold">任务管理</h2>
          <div className="flex gap-2">
            <Button
              variant="outline"
              onClick={() => setShowLabelManager(!showLabelManager)}
            >
              <svg className="h-4 w-4 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M9.568 3H5.25A2.25 2.25 0 003 5.25v4.318c0 .597.237 1.17.659 1.591l9.581 9.581c.699.699 1.78.872 2.607.33a18.095 18.095 0 005.223-5.223c.542-.827.369-1.908-.33-2.607L11.16 3.66A2.25 2.25 0 009.568 3z"
                />
                <path strokeLinecap="round" strokeLinejoin="round" d="M6 6h.008v.008H6V6z" />
              </svg>
              标签管理
            </Button>
            <Button onClick={() => setShowCreateDialog(true)}>
              <svg className="h-4 w-4 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
              </svg>
              新建任务
            </Button>
          </div>
        </div>

        {/* 筛选栏 */}
        <div className="mb-6">
          <TaskFiltersBar
            filters={filters}
            onFiltersChange={setFilters}
            labels={labelsData}
          />
        </div>

        {/* 标签管理面板 */}
        {showLabelManager && (
          <div className="mb-6 rounded-xl border bg-card p-4 max-w-sm">
            <LabelManager />
          </div>
        )}

        {/* 任务详情侧边面板 */}
        {selectedTask && (
          <div className="mb-6 rounded-xl border bg-card p-5">
            <div className="flex items-start justify-between mb-4">
              <div>
                <h3 className="text-lg font-semibold">{selectedTask.title}</h3>
                {selectedTask.description && (
                  <p className="text-sm text-muted-foreground mt-1">
                    {selectedTask.description}
                  </p>
                )}
              </div>
              <Button
                variant="ghost"
                size="icon"
                onClick={() => setSelectedTask(null)}
              >
                <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </Button>
            </div>
            <SubtaskList taskId={selectedTask.id} />
          </div>
        )}

        {/* 看板视图 */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {STATUS_COLUMNS.map((col) => (
            <KanbanColumn
              key={col.status}
              title={col.title}
              status={col.status}
              tasks={tasksByStatus[col.status]}
              onStatusChange={handleStatusChange}
              labels={labelsData}
              isLoading={isLoadingTasks}
            />
          ))}
        </div>

        {/* 加载更多 */}
        {hasNextPage && (
          <div className="flex justify-center mt-6">
            <Button
              variant="outline"
              onClick={() => fetchNextPage()}
              disabled={isFetchingNextPage}
            >
              {isFetchingNextPage ? '加载中...' : '加载更多'}
            </Button>
          </div>
        )}

        {/* 创建任务对话框 */}
        {showCreateDialog && (
          <div className="fixed inset-0 z-50 flex items-center justify-center">
            <div
              className="fixed inset-0 bg-black/50"
              onClick={() => setShowCreateDialog(false)}
            />
            <div className="relative z-10 w-full max-w-lg rounded-xl border bg-card p-6 shadow-lg mx-4">
              <h3 className="text-lg font-semibold mb-4">新建任务</h3>
              <TaskCreateForm
                onSubmit={(values) => createMutation.mutate(values)}
                onCancel={() => setShowCreateDialog(false)}
                learningGoals={goalsData}
                isSubmitting={createMutation.isPending}
              />
            </div>
          </div>
        )}
      </main>
    </div>
  );
}
