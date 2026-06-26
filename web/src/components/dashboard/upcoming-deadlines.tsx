'use client';

import { cn } from '@/lib/utils';
import type { Task } from '@/types/api';

interface UpcomingDeadlinesProps {
  tasks: Task[];
  onTaskClick?: (task: Task) => void;
  className?: string;
}

const PRIORITY_CONFIG = {
  low: { label: '低', className: 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400' },
  medium: { label: '中', className: 'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-400' },
  high: { label: '高', className: 'bg-orange-100 text-orange-700 dark:bg-orange-900/40 dark:text-orange-400' },
  urgent: { label: '紧急', className: 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-400' },
} as const;

function isOverdue(deadline: string | null): boolean {
  if (!deadline) return false;
  return new Date(deadline) < new Date();
}

function formatDeadline(deadline: string): string {
  const date = new Date(deadline);
  const now = new Date();
  const diffMs = date.getTime() - now.getTime();
  const diffDays = Math.ceil(diffMs / (1000 * 60 * 60 * 24));

  if (diffDays < 0) {
    return `已逾期 ${Math.abs(diffDays)} 天`;
  }
  if (diffDays === 0) {
    return '今天截止';
  }
  if (diffDays === 1) {
    return '明天截止';
  }
  if (diffDays <= 7) {
    return `${diffDays} 天后截止`;
  }
  return date.toLocaleDateString('zh-CN', { month: 'numeric', day: 'numeric' });
}

export function UpcomingDeadlines({ tasks, onTaskClick, className }: UpcomingDeadlinesProps) {
  const sortedTasks = [...tasks]
    .filter((t) => t.deadline)
    .sort((a, b) => new Date(a.deadline!).getTime() - new Date(b.deadline!).getTime());

  return (
    <div className={cn('rounded-lg border bg-card p-4 shadow-sm', className)}>
      <h3 className="mb-3 text-sm font-semibold text-foreground">即将到期</h3>

      {sortedTasks.length === 0 ? (
        <p className="py-6 text-center text-sm text-muted-foreground">暂无即将到期的任务</p>
      ) : (
        <div className="space-y-2">
          {sortedTasks.map((task) => {
            const priority = PRIORITY_CONFIG[task.priority];
            const overdue = isOverdue(task.deadline);

            return (
              <div
                key={task.id}
                onClick={() => onTaskClick?.(task)}
                className={cn(
                  'flex items-center gap-3 rounded-md px-3 py-2',
                  'cursor-pointer transition-colors hover:bg-accent',
                  overdue && 'border-l-2 border-l-red-500'
                )}
              >
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <p className="truncate text-sm font-medium text-foreground">
                      {task.title}
                    </p>
                    <span
                      className={cn(
                        'inline-flex shrink-0 items-center rounded-md px-1.5 py-0.5 text-[10px] font-medium leading-none',
                        priority.className
                      )}
                    >
                      {priority.label}
                    </span>
                  </div>
                  <p
                    className={cn(
                      'mt-0.5 text-xs',
                      overdue
                        ? 'font-medium text-red-500'
                        : 'text-muted-foreground'
                    )}
                  >
                    {formatDeadline(task.deadline!)}
                  </p>
                </div>

                {overdue && (
                  <span className="inline-flex shrink-0 items-center rounded-md bg-red-100 px-1.5 py-0.5 text-[10px] font-medium text-red-700 dark:bg-red-900/40 dark:text-red-400">
                    逾期
                  </span>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
