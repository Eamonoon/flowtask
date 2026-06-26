'use client';

import { useState, useCallback } from 'react';
import { cn } from '@/lib/utils';
import type { Task } from '@/types/api';

interface TaskTreeProps {
  tasks: Task[];
  onTaskClick?: (task: Task) => void;
  onReorder?: (taskIds: string[]) => void;
  className?: string;
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

interface TreeNode {
  task: Task;
  children: TreeNode[];
}

function buildTree(tasks: Task[]): TreeNode[] {
  const map = new Map<string, TreeNode>();
  const roots: TreeNode[] = [];

  for (const task of tasks) {
    map.set(task.id, { task, children: [] });
  }

  for (const task of tasks) {
    const node = map.get(task.id)!;
    if (task.parent_task_id && map.has(task.parent_task_id)) {
      map.get(task.parent_task_id)!.children.push(node);
    } else {
      roots.push(node);
    }
  }

  return roots;
}

interface TreeNodeViewProps {
  node: TreeNode;
  depth: number;
  onTaskClick?: (task: Task) => void;
  expandedIds: Set<string>;
  onToggle: (id: string) => void;
  draggedId: string | null;
  onDragStart: (id: string) => void;
  onDragEnd: () => void;
  onDrop: (id: string) => void;
}

function TreeNodeView({
  node,
  depth,
  onTaskClick,
  expandedIds,
  onToggle,
  draggedId,
  onDragStart,
  onDragEnd,
  onDrop,
}: TreeNodeViewProps) {
  const { task, children } = node;
  const hasChildren = children.length > 0;
  const isExpanded = expandedIds.has(task.id);
  const isDragging = draggedId === task.id;
  const priority = PRIORITY_CONFIG[task.priority];
  const status = STATUS_CONFIG[task.status];

  return (
    <div>
      <div
        className={cn(
          'group flex items-center gap-2 rounded-md px-2 py-1.5',
          'cursor-pointer transition-colors hover:bg-accent',
          isDragging && 'opacity-50'
        )}
        style={{ paddingLeft: `${depth * 20 + 8}px` }}
        draggable
        onDragStart={() => onDragStart(task.id)}
        onDragEnd={onDragEnd}
        onDragOver={(e) => e.preventDefault()}
        onDrop={(e) => {
          e.preventDefault();
          onDrop(task.id);
        }}
        onClick={() => onTaskClick?.(task)}
      >
        {/* 展开/折叠 */}
        {hasChildren ? (
          <button
            onClick={(e) => {
              e.stopPropagation();
              onToggle(task.id);
            }}
            className="flex h-4 w-4 shrink-0 items-center justify-center rounded text-muted-foreground hover:bg-muted"
          >
            <svg
              className={cn('h-3 w-3 transition-transform', isExpanded && 'rotate-90')}
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={2}
            >
              <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 4.5l7.5 7.5-7.5 7.5" />
            </svg>
          </button>
        ) : (
          <div className="flex h-4 w-4 shrink-0 items-center justify-center">
            <div
              className={cn(
                'h-1.5 w-1.5 rounded-full',
                task.status === 'done' ? 'bg-emerald-500' : 'bg-muted-foreground/40'
              )}
            />
          </div>
        )}

        {/* 依赖箭头指示 */}
        {task.dependencies.length > 0 && (
          <svg className="h-3.5 w-3.5 shrink-0 text-muted-foreground/60" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M7.5 21L3 16.5m0 0L7.5 12M3 16.5h13.5m0-13.5L21 7.5m0 0L16.5 12M21 7.5H7.5" />
          </svg>
        )}

        {/* 状态 */}
        <span
          className={cn(
            'inline-flex shrink-0 items-center rounded-md px-1.5 py-0.5 text-[10px] font-medium leading-none',
            status.className
          )}
        >
          {status.label}
        </span>

        {/* 标题 */}
        <span
          className={cn(
            'flex-1 truncate text-sm',
            task.status === 'done'
              ? 'line-through text-muted-foreground'
              : 'text-foreground'
          )}
        >
          {task.title}
        </span>

        {/* 优先级 */}
        <span
          className={cn(
            'inline-flex shrink-0 items-center rounded-md px-1.5 py-0.5 text-[10px] font-medium leading-none',
            priority.className
          )}
        >
          {priority.label}
        </span>

        {/* 拖拽手柄 */}
        <div className="opacity-0 group-hover:opacity-100 transition-opacity shrink-0 cursor-grab active:cursor-grabbing">
          <svg className="h-4 w-4 text-muted-foreground" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5" />
          </svg>
        </div>
      </div>

      {/* 子节点 */}
      {hasChildren && isExpanded && (
        <div>
          {children.map((child) => (
            <TreeNodeView
              key={child.task.id}
              node={child}
              depth={depth + 1}
              onTaskClick={onTaskClick}
              expandedIds={expandedIds}
              onToggle={onToggle}
              draggedId={draggedId}
              onDragStart={onDragStart}
              onDragEnd={onDragEnd}
              onDrop={onDrop}
            />
          ))}
        </div>
      )}
    </div>
  );
}

export function TaskTree({ tasks, onTaskClick, onReorder, className }: TaskTreeProps) {
  const [expandedIds, setExpandedIds] = useState<Set<string>>(new Set());
  const [draggedId, setDraggedId] = useState<string | null>(null);

  const tree = buildTree(tasks);

  const toggleExpand = useCallback((id: string) => {
    setExpandedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  }, []);

  const handleDrop = useCallback(
    (targetId: string) => {
      if (!draggedId || draggedId === targetId || !onReorder) return;
      const ids = tasks.map((t) => t.id);
      const fromIdx = ids.indexOf(draggedId);
      const toIdx = ids.indexOf(targetId);
      if (fromIdx === -1 || toIdx === -1) return;

      const reordered = [...ids];
      reordered.splice(fromIdx, 1);
      reordered.splice(toIdx, 0, draggedId);
      onReorder(reordered);
    },
    [draggedId, tasks, onReorder]
  );

  if (tasks.length === 0) {
    return (
      <div className={cn('rounded-lg border bg-card p-8 text-center', className)}>
        <p className="text-sm text-muted-foreground">暂无任务</p>
      </div>
    );
  }

  return (
    <div className={cn('rounded-lg border bg-card', className)}>
      {tree.map((node) => (
        <TreeNodeView
          key={node.task.id}
          node={node}
          depth={0}
          onTaskClick={onTaskClick}
          expandedIds={expandedIds}
          onToggle={toggleExpand}
          draggedId={draggedId}
          onDragStart={setDraggedId}
          onDragEnd={() => setDraggedId(null)}
          onDrop={handleDrop}
        />
      ))}
    </div>
  );
}
