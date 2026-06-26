'use client';

import { cn } from '@/lib/utils';
import type { Task } from '@/types/api';

interface TaskDetailDialogProps {
  task: Task | null;
  open: boolean;
  onClose: () => void;
  onStatusChange?: (task: Task, status: Task['status']) => void;
  onEdit?: (task: Task) => void;
}

const PRIORITY_CONFIG = {
  low: { label: '低', className: 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400' },
  medium: { label: '中', className: 'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-400' },
  high: { label: '高', className: 'bg-orange-100 text-orange-700 dark:bg-orange-900/40 dark:text-orange-400' },
  urgent: { label: '紧急', className: 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-400' },
} as const;

const STATUS_CONFIG = {
  todo: { label: '待办', className: 'bg-muted text-muted-foreground' },
  doing: { label: '进行中', className: 'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-400' },
  done: { label: '已完成', className: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-400' },
} as const;

export function TaskDetailDialog({
  task,
  open,
  onClose,
  onStatusChange,
  onEdit,
}: TaskDetailDialogProps) {
  if (!open || !task) return null;

  const priority = PRIORITY_CONFIG[task.priority];
  const status = STATUS_CONFIG[task.status];

  const hasSubtasks = task.subtask_count > 0;
  const subtaskProgress = hasSubtasks
    ? Math.round((task.completed_subtask_count / task.subtask_count) * 100)
    : 0;

  const deadlineDate = task.deadline
    ? new Date(task.deadline).toLocaleDateString('zh-CN', {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
      })
    : null;

  const isOverdue =
    task.deadline &&
    task.status !== 'done' &&
    new Date(task.deadline) < new Date();

  const handleOverlayClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      onClose();
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Escape') {
      onClose();
    }
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
      onClick={handleOverlayClick}
      onKeyDown={handleKeyDown}
      role="dialog"
      aria-modal="true"
      tabIndex={-1}
    >
      <div className="relative w-full max-w-lg max-h-[85vh] overflow-y-auto rounded-xl border bg-card shadow-lg">
        {/* 头部 */}
        <div className="sticky top-0 z-10 flex items-start justify-between border-b bg-card px-5 py-4">
          <div className="min-w-0 flex-1 pr-4">
            <h2 className="text-lg font-semibold text-foreground">{task.title}</h2>
            <div className="mt-1.5 flex flex-wrap items-center gap-2">
              <span
                className={cn(
                  'inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium',
                  status.className
                )}
              >
                {status.label}
              </span>
              <span
                className={cn(
                  'inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium',
                  priority.className
                )}
              >
                {priority.label}优先级
              </span>
              {isOverdue && (
                <span className="inline-flex items-center rounded-md bg-red-100 px-2 py-0.5 text-xs font-medium text-red-700 dark:bg-red-900/40 dark:text-red-400">
                  已逾期
                </span>
              )}
            </div>
          </div>
          <button
            onClick={onClose}
            className="shrink-0 rounded-md p-1 text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
          >
            <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        {/* 内容 */}
        <div className="space-y-4 px-5 py-4">
          {/* 描述 */}
          {task.description && (
            <div>
              <h3 className="mb-1 text-xs font-medium text-muted-foreground">描述</h3>
              <p className="text-sm text-foreground whitespace-pre-wrap">{task.description}</p>
            </div>
          )}

          {/* 元信息 */}
          <div className="grid grid-cols-2 gap-3">
            {task.estimated_duration && (
              <div>
                <h3 className="mb-0.5 text-xs font-medium text-muted-foreground">预计时长</h3>
                <p className="text-sm text-foreground">{task.estimated_duration}</p>
              </div>
            )}
            {deadlineDate && (
              <div>
                <h3 className="mb-0.5 text-xs font-medium text-muted-foreground">截止日期</h3>
                <p className={cn('text-sm', isOverdue ? 'text-red-500 font-medium' : 'text-foreground')}>
                  {deadlineDate}
                </p>
              </div>
            )}
          </div>

          {/* 子任务进度 */}
          {hasSubtasks && (
            <div>
              <div className="mb-1.5 flex items-center justify-between">
                <h3 className="text-xs font-medium text-muted-foreground">子任务进度</h3>
                <span className="text-xs text-muted-foreground">
                  {task.completed_subtask_count}/{task.subtask_count}
                </span>
              </div>
              <div className="h-2 w-full rounded-full bg-muted overflow-hidden">
                <div
                  className="h-full rounded-full bg-primary transition-all duration-300"
                  style={{ width: `${subtaskProgress}%` }}
                />
              </div>
            </div>
          )}

          {/* 标签 */}
          {task.labels.length > 0 && (
            <div>
              <h3 className="mb-1.5 text-xs font-medium text-muted-foreground">标签</h3>
              <div className="flex flex-wrap gap-1.5">
                {task.labels.map((label) => (
                  <span
                    key={label.id}
                    className="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium"
                    style={{
                      backgroundColor: `${label.color}20`,
                      color: label.color,
                    }}
                  >
                    {label.name}
                  </span>
                ))}
              </div>
            </div>
          )}

          {/* 推荐资源 */}
          {task.recommended_resources.length > 0 && (
            <div>
              <h3 className="mb-1.5 text-xs font-medium text-muted-foreground">推荐资源</h3>
              <div className="space-y-2">
                {task.recommended_resources.map((resource, index) => (
                  <div key={index} className="rounded-md border p-2.5">
                    {resource.url ? (
                      <a
                        href={resource.url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-sm font-medium text-primary hover:underline"
                      >
                        {resource.title}
                        <svg className="ml-1 inline h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                          <path strokeLinecap="round" strokeLinejoin="round" d="M13.5 6H5.25A2.25 2.25 0 003 8.25v10.5A2.25 2.25 0 005.25 21h10.5A2.25 2.25 0 0018 18.75V10.5m-10.5 6L21 3m0 0h-5.25M21 3v5.25" />
                        </svg>
                      </a>
                    ) : (
                      <span className="text-sm font-medium text-foreground">{resource.title}</span>
                    )}
                    {resource.description && (
                      <p className="mt-0.5 text-xs text-muted-foreground">{resource.description}</p>
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* 底部操作 */}
        <div className="sticky bottom-0 flex items-center justify-between border-t bg-card px-5 py-3">
          <div className="flex gap-1">
            {(['todo', 'doing', 'done'] as const).map((s) => (
              <button
                key={s}
                onClick={() => onStatusChange?.(task, s)}
                disabled={task.status === s}
                className={cn(
                  'rounded-md px-2.5 py-1 text-xs font-medium transition-colors',
                  task.status === s
                    ? STATUS_CONFIG[s].className
                    : 'text-muted-foreground hover:bg-muted'
                )}
              >
                {STATUS_CONFIG[s].label}
              </button>
            ))}
          </div>
          <button
            onClick={() => onEdit?.(task)}
            className="inline-flex items-center gap-1 rounded-md bg-primary px-3 py-1.5 text-xs font-medium text-primary-foreground shadow-sm hover:bg-primary/80 transition-colors"
          >
            <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L6.832 19.82a4.5 4.5 0 01-1.897 1.13l-2.685.8.8-2.685a4.5 4.5 0 011.13-1.897L16.863 4.487z" />
            </svg>
            编辑
          </button>
        </div>
      </div>
    </div>
  );
}
