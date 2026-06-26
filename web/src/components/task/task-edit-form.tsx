'use client';

import { useForm, Controller } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { Button } from '@/components/ui/button';
import type { Task, LearningGoal } from '@/types/api';

const taskEditSchema = z.object({
  title: z.string().min(1, '请输入任务标题'),
  description: z.string().optional(),
  estimated_duration: z.string().optional(),
  status: z.enum(['todo', 'doing', 'done']),
  priority: z.enum(['low', 'medium', 'high', 'urgent']),
  learning_goal_id: z.string().optional().nullable(),
});

export type TaskEditFormValues = z.infer<typeof taskEditSchema>;

interface TaskEditFormProps {
  task: Task;
  onSubmit: (values: TaskEditFormValues) => void;
  onCancel: () => void;
  learningGoals?: LearningGoal[];
  isSubmitting?: boolean;
}

const PRIORITY_OPTIONS = [
  { value: 'low', label: '低优先级' },
  { value: 'medium', label: '中优先级' },
  { value: 'high', label: '高优先级' },
  { value: 'urgent', label: '紧急' },
] as const;

const STATUS_OPTIONS = [
  { value: 'todo', label: '待办' },
  { value: 'doing', label: '进行中' },
  { value: 'done', label: '已完成' },
] as const;

export function TaskEditForm({
  task,
  onSubmit,
  onCancel,
  learningGoals,
  isSubmitting,
}: TaskEditFormProps) {
  const {
    register,
    handleSubmit,
    control,
    formState: { errors },
  } = useForm<TaskEditFormValues>({
    resolver: zodResolver(taskEditSchema),
    defaultValues: {
      title: task.title,
      description: task.description ?? '',
      estimated_duration: task.estimated_duration ?? '',
      status: task.status,
      priority: task.priority,
      learning_goal_id: task.learning_goal_id ?? '',
    },
  });

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      {/* 任务标题 */}
      <div className="space-y-1.5">
        <label htmlFor="edit-title" className="text-sm font-medium text-foreground">
          任务标题 <span className="text-destructive">*</span>
        </label>
        <input
          id="edit-title"
          type="text"
          placeholder="输入任务标题..."
          className="flex h-9 w-full rounded-lg border border-input bg-background px-3 py-1 text-sm shadow-sm transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          {...register('title')}
        />
        {errors.title && (
          <p className="text-xs text-destructive">{errors.title.message}</p>
        )}
      </div>

      {/* 任务描述 */}
      <div className="space-y-1.5">
        <label htmlFor="edit-desc" className="text-sm font-medium text-foreground">
          描述
        </label>
        <textarea
          id="edit-desc"
          rows={3}
          placeholder="输入任务描述..."
          className="flex w-full rounded-lg border border-input bg-background px-3 py-2 text-sm shadow-sm transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring resize-none"
          {...register('description')}
        />
      </div>

      {/* 预计时长 */}
      <div className="space-y-1.5">
        <label htmlFor="edit-duration" className="text-sm font-medium text-foreground">
          预计时长
        </label>
        <input
          id="edit-duration"
          type="text"
          placeholder="例如: 2小时、30分钟..."
          className="flex h-9 w-full rounded-lg border border-input bg-background px-3 py-1 text-sm shadow-sm transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          {...register('estimated_duration')}
        />
      </div>

      {/* 状态 */}
      <div className="space-y-1.5">
        <label htmlFor="edit-status" className="text-sm font-medium text-foreground">
          状态
        </label>
        <Controller
          name="status"
          control={control}
          render={({ field }) => (
            <select
              id="edit-status"
              className="flex h-9 w-full rounded-lg border border-input bg-background px-3 py-1 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              value={field.value}
              onChange={field.onChange}
              onBlur={field.onBlur}
            >
              {STATUS_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          )}
        />
      </div>

      {/* 优先级 */}
      <div className="space-y-1.5">
        <label htmlFor="edit-priority" className="text-sm font-medium text-foreground">
          优先级
        </label>
        <Controller
          name="priority"
          control={control}
          render={({ field }) => (
            <select
              id="edit-priority"
              className="flex h-9 w-full rounded-lg border border-input bg-background px-3 py-1 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              value={field.value}
              onChange={field.onChange}
              onBlur={field.onBlur}
            >
              {PRIORITY_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          )}
        />
      </div>

      {/* 关联学习目标 */}
      {learningGoals && learningGoals.length > 0 && (
        <div className="space-y-1.5">
          <label htmlFor="edit-goal" className="text-sm font-medium text-foreground">
            关联学习目标
          </label>
          <Controller
            name="learning_goal_id"
            control={control}
            render={({ field }) => (
              <select
                id="edit-goal"
                className="flex h-9 w-full rounded-lg border border-input bg-background px-3 py-1 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                value={field.value ?? ''}
                onChange={(e) => field.onChange(e.target.value || null)}
                onBlur={field.onBlur}
              >
                <option value="">不关联</option>
                {learningGoals.map((goal) => (
                  <option key={goal.id} value={goal.id}>
                    {goal.description}
                  </option>
                ))}
              </select>
            )}
          />
        </div>
      )}

      {/* 操作按钮 */}
      <div className="flex justify-end gap-2 pt-2">
        <Button type="button" variant="outline" onClick={onCancel}>
          取消
        </Button>
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? '保存中...' : '保存修改'}
        </Button>
      </div>
    </form>
  );
}
