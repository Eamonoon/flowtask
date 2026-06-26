'use client';

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import api from '@/lib/api';
import type { Label, ApiResponse } from '@/types/api';
import { cn } from '@/lib/utils';

const PRESET_COLORS = [
  '#ef4444', '#f97316', '#eab308', '#22c55e', '#14b8a6',
  '#3b82f6', '#6366f1', '#8b5cf6', '#a855f7', '#ec4899',
  '#6b7280', '#78716c',
];

interface LabelManagerProps {
  className?: string;
}

export function LabelManager({ className }: LabelManagerProps) {
  const queryClient = useQueryClient();
  const [newName, setNewName] = useState('');
  const [newColor, setNewColor] = useState(PRESET_COLORS[5]);
  const [showCreate, setShowCreate] = useState(false);

  // 获取标签列表
  const { data: labelsData, isLoading } = useQuery({
    queryKey: ['labels'],
    queryFn: async () => {
      const { data } = await api.get<ApiResponse<Label[]>>('/labels');
      return data.data;
    },
  });

  const labels = labelsData ?? [];

  // 创建标签
  const createMutation = useMutation({
    mutationFn: async (payload: { name: string; color: string }) => {
      const { data } = await api.post<ApiResponse<Label>>('/labels', payload);
      return data.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['labels'] });
      setNewName('');
      setNewColor(PRESET_COLORS[5]);
      setShowCreate(false);
    },
  });

  // 删除标签
  const deleteMutation = useMutation({
    mutationFn: async (labelId: string) => {
      await api.delete(`/labels/${labelId}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['labels'] });
    },
  });

  const handleCreate = () => {
    const trimmed = newName.trim();
    if (!trimmed) return;
    createMutation.mutate({ name: trimmed, color: newColor });
  };

  return (
    <div className={cn('space-y-3', className)}>
      <div className="flex items-center justify-between">
        <h4 className="text-sm font-semibold text-foreground">标签管理</h4>
        <Button
          variant="ghost"
          size="sm"
          onClick={() => setShowCreate(!showCreate)}
        >
          {showCreate ? '取消' : '新建标签'}
        </Button>
      </div>

      {/* 创建表单 */}
      {showCreate && (
        <div className="rounded-lg border p-3 space-y-3 bg-muted/30">
          <input
            type="text"
            placeholder="标签名称"
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            maxLength={20}
            className="flex h-8 w-full rounded-md border border-input bg-background px-2.5 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          />
          <div className="flex flex-wrap gap-1.5">
            {PRESET_COLORS.map((color) => (
              <button
                key={color}
                type="button"
                onClick={() => setNewColor(color)}
                className={cn(
                  'h-6 w-6 rounded-full border-2 transition-transform',
                  newColor === color
                    ? 'border-foreground scale-110'
                    : 'border-transparent hover:scale-110'
                )}
                style={{ backgroundColor: color }}
              />
            ))}
          </div>
          <div className="flex items-center gap-2">
            <input
              type="color"
              value={newColor}
              onChange={(e) => setNewColor(e.target.value)}
              className="h-6 w-8 cursor-pointer rounded border-0 bg-transparent p-0"
            />
            <span className="text-xs text-muted-foreground">{newColor}</span>
          </div>
          <Button
            size="sm"
            onClick={handleCreate}
            disabled={!newName.trim() || createMutation.isPending}
          >
            {createMutation.isPending ? '创建中...' : '创建'}
          </Button>
        </div>
      )}

      {/* 标签列表 */}
      {isLoading ? (
        <div className="flex items-center justify-center py-4">
          <div className="h-4 w-4 animate-spin rounded-full border-2 border-primary border-t-transparent" />
        </div>
      ) : labels.length === 0 ? (
        <p className="text-xs text-muted-foreground text-center py-4">暂无标签</p>
      ) : (
        <div className="space-y-1">
          {labels.map((label) => (
            <div
              key={label.id}
              className="flex items-center gap-2 rounded-md px-2 py-1.5 hover:bg-accent group"
            >
              <span
                className="h-3 w-3 rounded-full shrink-0"
                style={{ backgroundColor: label.color }}
              />
              <span className="text-sm text-foreground flex-1">{label.name}</span>
              <button
                onClick={() => deleteMutation.mutate(label.id)}
                disabled={deleteMutation.isPending}
                className="opacity-0 group-hover:opacity-100 text-muted-foreground hover:text-destructive transition-opacity"
                title="删除标签"
              >
                <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
