'use client';

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { cn } from '@/lib/utils';
import api from '@/lib/api';
import type { ApiResponse } from '@/types/api';

export interface Subtask {
  id: string;
  title: string;
  is_completed: boolean;
  sort_order: number;
}

interface SubtaskListProps {
  taskId: string;
  className?: string;
}

export function SubtaskList({ taskId, className }: SubtaskListProps) {
  const queryClient = useQueryClient();
  const [newTitle, setNewTitle] = useState('');

  // 获取子任务列表
  const { data: subtasksData, isLoading } = useQuery({
    queryKey: ['subtasks', taskId],
    queryFn: async () => {
      const { data } = await api.get<ApiResponse<Subtask[]>>(
        `/tasks/${taskId}/subtasks`
      );
      return data.data;
    },
    enabled: !!taskId,
  });

  const subtasks = subtasksData ?? [];
  const completedCount = subtasks.filter((s) => s.is_completed).length;

  // 创建子任务
  const createMutation = useMutation({
    mutationFn: async (title: string) => {
      const { data } = await api.post<ApiResponse<Subtask>>(
        `/tasks/${taskId}/subtasks`,
        { title }
      );
      return data.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['subtasks', taskId] });
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
      setNewTitle('');
    },
  });

  // 切换子任务完成状态
  const toggleMutation = useMutation({
    mutationFn: async (subtask: Subtask) => {
      const { data } = await api.patch<ApiResponse<Subtask>>(
        `/tasks/${taskId}/subtasks/${subtask.id}`,
        { is_completed: !subtask.is_completed }
      );
      return data.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['subtasks', taskId] });
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
    },
  });

  // 删除子任务
  const deleteMutation = useMutation({
    mutationFn: async (subtaskId: string) => {
      await api.delete(`/tasks/${taskId}/subtasks/${subtaskId}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['subtasks', taskId] });
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
    },
  });

  const handleAdd = () => {
    const trimmed = newTitle.trim();
    if (!trimmed) return;
    createMutation.mutate(trimmed);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      handleAdd();
    }
  };

  return (
    <div className={cn('space-y-3', className)}>
      {/* 头部 */}
      <div className="flex items-center justify-between">
        <h4 className="text-sm font-semibold text-foreground">子任务</h4>
        {subtasks.length > 0 && (
          <span className="text-xs text-muted-foreground">
            {completedCount}/{subtasks.length} 已完成
          </span>
        )}
      </div>

      {/* 进度条 */}
      {subtasks.length > 0 && (
        <div className="h-1.5 w-full rounded-full bg-muted overflow-hidden">
          <div
            className="h-full rounded-full bg-primary transition-all duration-300"
            style={{
              width: `${(completedCount / subtasks.length) * 100}%`,
            }}
          />
        </div>
      )}

      {/* 子任务列表 */}
      {isLoading ? (
        <div className="flex items-center justify-center py-4">
          <div className="h-4 w-4 animate-spin rounded-full border-2 border-primary border-t-transparent" />
        </div>
      ) : (
        <div className="space-y-1">
          {subtasks.map((subtask) => (
            <div
              key={subtask.id}
              className="flex items-center gap-2 rounded-md px-2 py-1.5 group hover:bg-accent"
            >
              <input
                type="checkbox"
                checked={subtask.is_completed}
                onChange={() => toggleMutation.mutate(subtask)}
                disabled={toggleMutation.isPending}
                className="h-3.5 w-3.5 rounded border-input accent-primary shrink-0 cursor-pointer"
              />
              <span
                className={cn(
                  'text-sm flex-1 transition-colors',
                  subtask.is_completed && 'line-through text-muted-foreground'
                )}
              >
                {subtask.title}
              </span>
              <button
                onClick={() => deleteMutation.mutate(subtask.id)}
                disabled={deleteMutation.isPending}
                className="opacity-0 group-hover:opacity-100 text-muted-foreground hover:text-destructive transition-opacity shrink-0"
                title="删除"
              >
                <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          ))}
        </div>
      )}

      {/* 添加子任务 */}
      <div className="flex gap-2">
        <input
          type="text"
          placeholder="添加子任务..."
          value={newTitle}
          onChange={(e) => setNewTitle(e.target.value)}
          onKeyDown={handleKeyDown}
          className="flex h-8 flex-1 rounded-md border border-input bg-background px-2.5 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        />
        <button
          type="button"
          onClick={handleAdd}
          disabled={!newTitle.trim() || createMutation.isPending}
          className="inline-flex h-8 items-center justify-center rounded-md bg-primary px-2.5 text-sm font-medium text-primary-foreground shadow-sm hover:bg-primary/80 disabled:opacity-50 transition-colors"
        >
          <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
          </svg>
        </button>
      </div>
    </div>
  );
}
