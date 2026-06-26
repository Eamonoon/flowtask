'use client';

import { useState } from 'react';
import { cn } from '@/lib/utils';
import type { Task } from '@/types/api';
import type { GenerationPhase } from '@/stores/goal-store';
import { Button } from '@/components/ui/button';

interface StreamingPlanViewerProps {
  phase: GenerationPhase;
  tasks: Task[];
  taskCount: number;
  onConfirmSave: () => Promise<void>;
  onRegenerate: () => Promise<void>;
  className?: string;
}

export function StreamingPlanViewer({
  phase,
  tasks,
  taskCount,
  onConfirmSave,
  onRegenerate,
  className,
}: StreamingPlanViewerProps) {
  const [isConfirming, setIsConfirming] = useState(false);
  const [isSaved, setIsSaved] = useState(false);

  const handleConfirm = async () => {
    setIsConfirming(true);
    try {
      await onConfirmSave();
      setIsSaved(true);
    } catch (err) {
      // Error is handled in parent
    } finally {
      setIsConfirming(false);
    }
  };

  const renderTaskItem = (task: Task, index: number) => (
    <div
      key={task.id || index}
      className="flex items-start gap-3 p-3 bg-muted/50 rounded-lg animate-in fade-in slide-in-from-left-2 duration-300"
      style={{ animationDelay: `${index * 50}ms` }}
    >
      <div className="h-2 w-2 mt-2 rounded-full bg-primary shrink-0" />
      <div className="flex-1 min-w-0">
        <h4 className="font-medium text-sm">{task.title}</h4>
        {task.description && (
          <p className="text-xs text-muted-foreground mt-1 line-clamp-2">
            {task.description}
          </p>
        )}
        {task.estimated_duration && (
          <span className="inline-block mt-1 text-xs bg-primary/10 text-primary px-2 py-0.5 rounded">
            {task.estimated_duration}
          </span>
        )}
      </div>
    </div>
  );

  return (
    <div className={cn('rounded-lg border bg-card p-6', className)}>
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-semibold">AI 生成学习计划</h3>
        {phase === 'streaming' && (
          <span className="text-sm text-muted-foreground">
            已生成 {taskCount} 个任务...
          </span>
        )}
      </div>

      {/* Connecting State */}
      {phase === 'connecting' && (
        <div className="flex flex-col items-center justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
          <p className="mt-3 text-muted-foreground">正在连接 AI 服务...</p>
        </div>
      )}

      {/* Streaming State */}
      {phase === 'streaming' && tasks.length === 0 && (
        <div className="flex flex-col items-center justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
          <p className="mt-3 text-muted-foreground">AI 正在生成任务...</p>
        </div>
      )}

      {/* Task List */}
      {tasks.length > 0 && (
        <div className="space-y-2 mb-4">
          {tasks.map((task, index) => renderTaskItem(task, index))}
          {phase === 'streaming' && (
            <div className="flex items-center gap-2 py-2 text-muted-foreground">
              <div className="h-4 w-4 animate-pulse rounded-full bg-primary/30" />
              <span className="text-sm">正在生成更多任务...</span>
            </div>
          )}
        </div>
      )}

      {/* Preview State - Confirm/Regenerate */}
      {phase === 'preview' && !isSaved && (
        <div className="mt-4 p-4 bg-muted rounded-lg">
          <p className="text-sm mb-3">
            已生成 {taskCount} 个任务，请确认是否保存到任务列表？
          </p>
          <div className="flex gap-3">
            <Button onClick={handleConfirm} disabled={isConfirming}>
              {isConfirming ? '保存中...' : '确认保存'}
            </Button>
            <Button variant="outline" onClick={onRegenerate}>
              重新生成
            </Button>
          </div>
        </div>
      )}

      {/* Done State */}
      {phase === 'done' || isSaved ? (
        <div className="mt-4 p-4 bg-emerald-50 dark:bg-emerald-900/20 rounded-lg">
          <p className="text-sm text-emerald-700 dark:text-emerald-400">
            ✓ 学习计划已保存，共 {taskCount} 个任务
          </p>
        </div>
      ) : null}

      {/* Error State */}
      {phase === 'error' && (
        <div className="mt-4 p-4 bg-red-50 dark:bg-red-900/20 rounded-lg">
          <p className="text-sm text-red-700 dark:text-red-400">
            生成失败，请重试
          </p>
          <Button
            variant="outline"
            size="sm"
            onClick={onRegenerate}
            className="mt-2"
          >
            重试
          </Button>
        </div>
      )}
    </div>
  );
}
