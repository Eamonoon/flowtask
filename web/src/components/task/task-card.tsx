'use client';

import { cn } from '@/lib/utils';
import type { Task, Label } from '@/types/api';

const PRIORITY_CONFIG = {
  low: { label: '低', className: 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400' },
  medium: { label: '中', className: 'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-400' },
  high: { label: '高', className: 'bg-orange-100 text-orange-700 dark:bg-orange-900/40 dark:text-orange-400' },
  urgent: { label: '紧急', className: 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-400' },
} as const;

interface TaskCardProps {
  task: Task;
  labels?: Label[];
  onDragStart?: (e: React.DragEvent) => void;
  onClick?: (task: Task) => void;
}

export function TaskCard({ task, labels, onDragStart, onClick }: TaskCardProps) {
  const priority = PRIORITY_CONFIG[task.priority] || PRIORITY_CONFIG.medium;

  const taskLabels = task.labels && task.labels.length > 0
    ? task.labels
    : [];

  const hasSubtasks = task.subtask_count && task.subtask_count > 0;
  const subtaskProgress = hasSubtasks
    ? `${task.completed_subtask_count || 0}/${task.subtask_count}`
    : null;

  const isOverdue =
    task.deadline &&
    task.status !== 'done' &&
    new Date(task.deadline) < new Date();

  const deadlineDate = task.deadline
    ? new Date(task.deadline).toLocaleDateString('zh-CN', {
        month: 'numeric',
        day: 'numeric',
      })
    : null;

  return (
    <div
      draggable
      onDragStart={onDragStart}
      onClick={() => onClick?.(task)}
      className={cn(
        'group cursor-grab rounded-lg border bg-card p-3',
        'shadow-sm hover:shadow-md transition-shadow',
        'active:cursor-grabbing active:opacity-80'
      )}
    >
      {/* 优先级 + 标题 */}
      <div className="flex items-start gap-2 mb-2">
        <span
          className={cn(
            'inline-flex items-center rounded-md px-1.5 py-0.5 text-[10px] font-medium leading-none shrink-0',
            priority.className
          )}
        >
          {priority.label}
        </span>
        <p className="text-sm font-medium text-foreground leading-snug line-clamp-2">
          {task.title || 'Untitled Task'}
        </p>
      </div>

      {/* 标签 */}
      {taskLabels.length > 0 && (
        <div className="flex flex-wrap gap-1 mb-2">
          {taskLabels.map((label) => (
            <span
              key={label.id}
              className="inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-medium leading-none"
              style={{
                backgroundColor: `${label.color}20`,
                color: label.color,
              }}
            >
              {label.name}
            </span>
          ))}
        </div>
      )}

      {/* 底部信息行 */}
      <div className="flex items-center gap-3 text-muted-foreground">
        {/* 截止日期 */}
        {deadlineDate && (
          <span
            className={cn(
              'inline-flex items-center gap-1 text-[11px]',
              isOverdue && 'text-red-500 font-medium'
            )}
          >
            <svg className="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 012.25-2.25h13.5A2.25 2.25 0 0121 7.5v11.25m-18 0A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75m-18 0v-7.5A2.25 2.25 0 015.25 9h13.5A2.25 2.25 0 0121 11.25v7.5"
              />
            </svg>
            {deadlineDate}
          </span>
        )}

        {/* 子任务进度 */}
        {subtaskProgress && (
          <span className="inline-flex items-center gap-1 text-[11px]">
            <svg className="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            {subtaskProgress}
          </span>
        )}
      </div>
    </div>
  );
}
