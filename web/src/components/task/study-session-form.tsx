'use client';

import { useForm, Controller } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { Button } from '@/components/ui/button';
import type { Task } from '@/types/api';

const studySessionSchema = z.object({
  task_id: z.string().optional(),
  duration: z
    .number({ message: '请输入有效的时长' })
    .min(1, '时长至少为 1 分钟')
    .max(600, '时长不能超过 600 分钟'),
  date: z.string().min(1, '请选择日期'),
  notes: z.string().optional(),
});

export type StudySessionFormValues = z.infer<typeof studySessionSchema>;

interface StudySessionFormProps {
  onSubmit: (values: StudySessionFormValues) => void;
  onCancel: () => void;
  tasks?: Task[];
  isSubmitting?: boolean;
}

export function StudySessionForm({
  onSubmit,
  onCancel,
  tasks,
  isSubmitting,
}: StudySessionFormProps) {
  const today = new Date().toISOString().split('T')[0];

  const {
    register,
    handleSubmit,
    control,
    formState: { errors },
  } = useForm<StudySessionFormValues>({
    resolver: zodResolver(studySessionSchema),
    defaultValues: {
      task_id: '',
      duration: 30,
      date: today,
      notes: '',
    },
  });

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      {/* 关联任务 */}
      {tasks && tasks.length > 0 && (
        <div className="space-y-1.5">
          <label htmlFor="session-task" className="text-sm font-medium text-foreground">
            关联任务
          </label>
          <Controller
            name="task_id"
            control={control}
            render={({ field }) => (
              <select
                id="session-task"
                className="flex h-9 w-full rounded-lg border border-input bg-background px-3 py-1 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                value={field.value ?? ''}
                onChange={field.onChange}
                onBlur={field.onBlur}
              >
                <option value="">不关联任务</option>
                {tasks.map((task) => (
                  <option key={task.id} value={task.id}>
                    {task.title}
                  </option>
                ))}
              </select>
            )}
          />
        </div>
      )}

      {/* 学习时长 */}
      <div className="space-y-1.5">
        <label htmlFor="session-duration" className="text-sm font-medium text-foreground">
          学习时长（分钟） <span className="text-destructive">*</span>
        </label>
        <input
          id="session-duration"
          type="number"
          min={1}
          max={600}
          placeholder="30"
          className="flex h-9 w-full rounded-lg border border-input bg-background px-3 py-1 text-sm shadow-sm transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          {...register('duration')}
        />
        {errors.duration && (
          <p className="text-xs text-destructive">{errors.duration.message}</p>
        )}
      </div>

      {/* 学习日期 */}
      <div className="space-y-1.5">
        <label htmlFor="session-date" className="text-sm font-medium text-foreground">
          学习日期 <span className="text-destructive">*</span>
        </label>
        <input
          id="session-date"
          type="date"
          className="flex h-9 w-full rounded-lg border border-input bg-background px-3 py-1 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          {...register('date')}
        />
        {errors.date && (
          <p className="text-xs text-destructive">{errors.date.message}</p>
        )}
      </div>

      {/* 学习笔记 */}
      <div className="space-y-1.5">
        <label htmlFor="session-notes" className="text-sm font-medium text-foreground">
          学习笔记
        </label>
        <textarea
          id="session-notes"
          rows={3}
          placeholder="记录本次学习的内容和感想..."
          className="flex w-full rounded-lg border border-input bg-background px-3 py-2 text-sm shadow-sm transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring resize-none"
          {...register('notes')}
        />
      </div>

      {/* 操作按钮 */}
      <div className="flex justify-end gap-2 pt-2">
        <Button type="button" variant="outline" onClick={onCancel}>
          取消
        </Button>
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? '记录中...' : '记录学习'}
        </Button>
      </div>
    </form>
  );
}
