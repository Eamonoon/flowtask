'use client';

import { useState, useCallback } from 'react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import type { Label } from '@/types/api';

export interface TaskFilters {
  keyword: string;
  statuses: string[];
  priorities: string[];
  labelIds: string[];
  deadlineFrom: string;
  deadlineTo: string;
  sortBy: 'created_at' | 'updated_at' | 'deadline' | 'priority' | 'sort_order';
  sortOrder: 'asc' | 'desc';
}

interface TaskFiltersBarProps {
  filters: TaskFilters;
  onFiltersChange: (filters: TaskFilters) => void;
  labels?: Label[];
}

const STATUS_OPTIONS = [
  { value: 'todo', label: '待办' },
  { value: 'doing', label: '进行中' },
  { value: 'done', label: '已完成' },
] as const;

const PRIORITY_OPTIONS = [
  { value: 'low', label: '低' },
  { value: 'medium', label: '中' },
  { value: 'high', label: '高' },
  { value: 'urgent', label: '紧急' },
] as const;

const SORT_OPTIONS = [
  { value: 'created_at', label: '创建时间' },
  { value: 'updated_at', label: '更新时间' },
  { value: 'deadline', label: '截止日期' },
  { value: 'priority', label: '优先级' },
  { value: 'sort_order', label: '自定义排序' },
] as const;

function MultiSelectDropdown({
  label,
  options,
  selected,
  onChange,
}: {
  label: string;
  options: readonly { value: string; label: string }[];
  selected: string[];
  onChange: (values: string[]) => void;
}) {
  const [open, setOpen] = useState(false);

  const toggle = (value: string) => {
    onChange(
      selected.includes(value)
        ? selected.filter((v) => v !== value)
        : [...selected, value]
    );
  };

  return (
    <div className="relative">
      <button
        type="button"
        onClick={() => setOpen(!open)}
        className={cn(
          'flex h-9 items-center gap-1.5 rounded-lg border border-input bg-background px-3 text-sm shadow-sm',
          'hover:bg-accent transition-colors',
          selected.length > 0 && 'border-primary/50'
        )}
      >
        <span className="text-muted-foreground">{label}</span>
        {selected.length > 0 && (
          <span className="inline-flex items-center justify-center h-4 min-w-4 rounded-full bg-primary text-primary-foreground text-[10px] px-1">
            {selected.length}
          </span>
        )}
        <svg
          className={cn('h-3.5 w-3.5 text-muted-foreground transition-transform', open && 'rotate-180')}
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          strokeWidth={2}
        >
          <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" />
        </svg>
      </button>

      {open && (
        <>
          <div className="fixed inset-0 z-10" onClick={() => setOpen(false)} />
          <div className="absolute z-20 mt-1 w-48 rounded-lg border bg-popover p-1 shadow-md">
            {options.map((opt) => (
              <label
                key={opt.value}
                className="flex items-center gap-2 rounded-md px-2 py-1.5 text-sm cursor-pointer hover:bg-accent"
              >
                <input
                  type="checkbox"
                  checked={selected.includes(opt.value)}
                  onChange={() => toggle(opt.value)}
                  className="h-3.5 w-3.5 rounded border-input accent-primary"
                />
                {opt.label}
              </label>
            ))}
          </div>
        </>
      )}
    </div>
  );
}

export function TaskFiltersBar({ filters, onFiltersChange, labels }: TaskFiltersBarProps) {
  const update = useCallback(
    (partial: Partial<TaskFilters>) => {
      onFiltersChange({ ...filters, ...partial });
    },
    [filters, onFiltersChange]
  );

  const hasActiveFilters =
    filters.keyword ||
    filters.statuses.length > 0 ||
    filters.priorities.length > 0 ||
    filters.labelIds.length > 0 ||
    filters.deadlineFrom ||
    filters.deadlineTo;

  const clearFilters = () => {
    onFiltersChange({
      keyword: '',
      statuses: [],
      priorities: [],
      labelIds: [],
      deadlineFrom: '',
      deadlineTo: '',
      sortBy: 'sort_order',
      sortOrder: 'asc',
    });
  };

  return (
    <div className="space-y-3">
      {/* 第一行：搜索 + 排序 */}
      <div className="flex flex-wrap items-center gap-2">
        {/* 搜索框 */}
        <div className="relative flex-1 min-w-[200px] max-w-md">
          <svg
            className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            strokeWidth={2}
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z"
            />
          </svg>
          <input
            type="text"
            placeholder="搜索任务..."
            value={filters.keyword}
            onChange={(e) => update({ keyword: e.target.value })}
            className="flex h-9 w-full rounded-lg border border-input bg-background pl-9 pr-3 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          />
        </div>

        {/* 排序 */}
        <select
          value={filters.sortBy}
          onChange={(e) => update({ sortBy: e.target.value as TaskFilters['sortBy'] })}
          className="flex h-9 rounded-lg border border-input bg-background px-3 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          {SORT_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </select>

        {/* 排序方向 */}
        <Button
          variant="outline"
          size="icon"
          onClick={() => update({ sortOrder: filters.sortOrder === 'asc' ? 'desc' : 'asc' })}
          title={filters.sortOrder === 'asc' ? '升序' : '降序'}
        >
          {filters.sortOrder === 'asc' ? (
            <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M3 4.5h14.25M3 9h9.75M3 13.5h5.25m5.25-.75L17.25 9m0 0L21 12.75M17.25 9v12" />
            </svg>
          ) : (
            <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M3 4.5h14.25M3 9h9.75M3 13.5h5.25m5.25.75L17.25 15m0 0L21 12.25M17.25 15V3" />
            </svg>
          )}
        </Button>

        {/* 清除筛选 */}
        {hasActiveFilters && (
          <Button variant="ghost" size="sm" onClick={clearFilters}>
            清除筛选
          </Button>
        )}
      </div>

      {/* 第二行：筛选器 */}
      <div className="flex flex-wrap items-center gap-2">
        <MultiSelectDropdown
          label="状态"
          options={STATUS_OPTIONS}
          selected={filters.statuses}
          onChange={(statuses) => update({ statuses })}
        />

        <MultiSelectDropdown
          label="优先级"
          options={PRIORITY_OPTIONS}
          selected={filters.priorities}
          onChange={(priorities) => update({ priorities })}
        />

        {labels && labels.length > 0 && (
          <MultiSelectDropdown
            label="标签"
            options={labels.map((l) => ({ value: l.id, label: l.name }))}
            selected={filters.labelIds}
            onChange={(labelIds) => update({ labelIds })}
          />
        )}

        {/* 截止日期范围 */}
        <div className="flex items-center gap-1.5">
          <span className="text-xs text-muted-foreground">截止日期</span>
          <input
            type="date"
            value={filters.deadlineFrom}
            onChange={(e) => update({ deadlineFrom: e.target.value })}
            className="flex h-9 rounded-lg border border-input bg-background px-2 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          />
          <span className="text-xs text-muted-foreground">至</span>
          <input
            type="date"
            value={filters.deadlineTo}
            onChange={(e) => update({ deadlineTo: e.target.value })}
            className="flex h-9 rounded-lg border border-input bg-background px-2 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          />
        </div>
      </div>
    </div>
  );
}
