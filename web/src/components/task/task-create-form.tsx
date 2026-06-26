'use client';

import { useForm, Controller } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { Button } from '@/components/ui/button';
import type { LearningGoal } from '@/types/api';

const taskSchema = z.object({
  title: z.string().min(1, '请输入任务标题'),
  description: z.string().optional(),
  priority: z.enum(['low', 'medium', 'high', 'urgent']),
  deadline: z.string().optional(),
  learning_goal_id: z.string().optional(),
});

export type TaskFormValues = z.infer<typeof taskSchema>;

interface TaskCreateFormProps {
  onSubmit: (values: TaskFormValues) => void;
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

export function TaskCreateForm({
  onSubmit,
  onCancel,
  learningGoals,
  isSubmitting,
}: TaskCreateFormProps) {
  const {
    register,
    handleSubmit,
    control,
    formState: { errors },
  } = useForm<TaskFormValues>({
    resolver: zodResolver(taskSchema),
    defaultValues: {
      title: '',
      description: '',
      priority: 'medium',
      deadline: '',
      learning_goal_id: '',
    },
  });

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      {/* 任务标题 */}
      <div className="space-y-1.5">
        <label htmlFor="task-title" className="text-sm font-medium text-foreground">
          任务标题 <span className="text-destructive">*</span>
        </label>
        <input
          id="task-title"
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
        <label htmlFor="task-desc" className="text-sm font-medium text-foreground">
          描述
        </label>
        <textarea
          id="task-desc"
          rows={3}
          placeholder="输入任务描述..."
          className="flex w-full rounded-lg border border-input bg-background px-3 py-2 text-sm shadow-sm transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring resize-none"
          {...register('description')}
        />
      </div>

      {/* 优先级 */}
      <div className="space-y-1.5">
        <label htmlFor="task-priority" className="text-sm font-medium text-foreground">
          优先级
        </label>
        <Controller
          name="priority"
          control={control}
          render={({ field }) => (
            <select
              id="task-priority"
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

      {/* 截止日期 */}
      <div className="space-y-1.5">
        <label htmlFor="task-deadline" className="text-sm font-medium text-foreground">
          截止日期
        </label>
        <input
          id="task-deadline"
          type="date"
          className="flex h-9 w-full rounded-lg border border-input bg-background px-3 py-1 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          {...register('deadline')}
        />
      </div>

      {/* 学习目标 */}
      {learningGoals && learningGoals.length > 0 && (
        <div className="space-y-1.5">
          <label htmlFor="task-goal" className="text-sm font-medium text-foreground">
            关联学习目标
          </label>
          <Controller
            name="learning_goal_id"
            control={control}
            render={({ field }) => (
              <select
                id="task-goal"
                className="flex h-9 w-full rounded-lg border border-input bg-background px-3 py-1 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                value={field.value ?? ''}
                onChange={field.onChange}
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
          {isSubmitting ? '创建中...' : '创建任务'}
        </Button>
      </div>
    </form>
  );
}
