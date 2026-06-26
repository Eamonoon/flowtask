'use client';

import { useCallback, useRef } from 'react';
import { cn } from '@/lib/utils';
import type { Task, Label } from '@/types/api';
import { TaskCard } from './task-card';

const STATUS_CONFIG = {
  todo: { label: '待办', color: 'bg-muted' },
  doing: { label: '进行中', color: 'bg-blue-500/10' },
  done: { label: '已完成', color: 'bg-green-500/10' },
} as const;

interface KanbanColumnProps {
  title: string;
  status: 'todo' | 'doing' | 'done';
  tasks: Task[];
  onStatusChange: (taskId: string, newStatus: 'todo' | 'doing' | 'done') => void;
  labels?: Label[];
  isLoading?: boolean;
}

export function KanbanColumn({
  title,
  status,
  tasks,
  onStatusChange,
  labels,
  isLoading,
}: KanbanColumnProps) {
  const columnRef = useRef<HTMLDivElement>(null);

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
  }, []);

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      const taskId = e.dataTransfer.getData('text/plain');
      if (taskId) {
        onStatusChange(taskId, status);
      }
    },
    [onStatusChange, status]
  );

  const config = STATUS_CONFIG[status];

  return (
    <div
      ref={columnRef}
      onDragOver={handleDragOver}
      onDrop={handleDrop}
      className={cn(
        'flex flex-col rounded-xl border bg-card min-w-[280px] flex-1',
        'transition-colors'
      )}
    >
      {/* 列头部 */}
      <div className="flex items-center gap-2 px-4 py-3 border-b">
        <span className={cn('h-2.5 w-2.5 rounded-full', config.color)} />
        <h3 className="text-sm font-semibold text-foreground">{title}</h3>
        <span className="ml-auto text-xs font-medium text-muted-foreground bg-muted rounded-full px-2 py-0.5">
          {tasks.length}
        </span>
      </div>

      {/* 任务列表 */}
      <div className="flex-1 overflow-y-auto p-3 space-y-2.5 min-h-[200px]">
        {isLoading ? (
          <div className="flex items-center justify-center py-8">
            <div className="h-5 w-5 animate-spin rounded-full border-2 border-primary border-t-transparent" />
          </div>
        ) : tasks.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-8 text-muted-foreground">
            <svg
              className="h-8 w-8 mb-2 opacity-40"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={1.5}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M9 12h3.75M9 15h3.75M9 18h3.75m3 .75H18a2.25 2.25 0 002.25-2.25V6.108c0-1.135-.845-2.098-1.976-2.192a48.424 48.424 0 00-1.123-.08m-5.801 0c-.065.21-.1.433-.1.664 0 .414.336.75.75.75h4.5a.75.75 0 00.75-.75 2.25 2.25 0 00-.1-.664m-5.8 0A2.251 2.251 0 0113.5 2.25H15c1.012 0 1.867.668 2.15 1.586m-5.8 0c-.376.023-.75.05-1.124.08C9.095 4.01 8.25 4.973 8.25 6.108V8.25m0 0H4.875c-.621 0-1.125.504-1.125 1.125v11.25c0 .621.504 1.125 1.125 1.125h9.75c.621 0 1.125-.504 1.125-1.125V9.375c0-.621-.504-1.125-1.125-1.125H8.25z"
              />
            </svg>
            <p className="text-xs">暂无任务</p>
          </div>
        ) : (
          tasks.map((task) => (
            <TaskCard
              key={task.id}
              task={task}
              labels={labels}
              onDragStart={(e) => {
                e.dataTransfer.setData('text/plain', task.id);
                e.dataTransfer.effectAllowed = 'move';
              }}
            />
          ))
        )}
      </div>
    </div>
  );
}
