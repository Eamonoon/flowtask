'use client';

import { cn } from '@/lib/utils';

function SkeletonLine({ className }: { className?: string }) {
  return (
    <div className={cn('rounded-md bg-muted animate-pulse', className)} />
  );
}

export function DashboardSkeleton() {
  return (
    <div className="space-y-6 p-6">
      {/* Header */}
      <SkeletonLine className="h-8 w-48" />

      {/* Stats cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <div key={i} className="p-4 border rounded-lg space-y-2">
            <SkeletonLine className="h-4 w-20" />
            <SkeletonLine className="h-8 w-16" />
          </div>
        ))}
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {Array.from({ length: 2 }).map((_, i) => (
          <div key={i} className="border rounded-lg p-4 space-y-4">
            <SkeletonLine className="h-5 w-32" />
            <SkeletonLine className="h-48 w-full" />
          </div>
        ))}
      </div>
    </div>
  );
}

export function TaskListSkeleton() {
  return (
    <div className="space-y-4 p-6">
      {/* Header bar */}
      <div className="flex items-center justify-between">
        <SkeletonLine className="h-8 w-32" />
        <SkeletonLine className="h-9 w-24" />
      </div>

      {/* Filters */}
      <div className="flex gap-2">
        <SkeletonLine className="h-9 w-20" />
        <SkeletonLine className="h-9 w-20" />
        <SkeletonLine className="h-9 w-32" />
      </div>

      {/* Task items */}
      {Array.from({ length: 6 }).map((_, i) => (
        <div
          key={i}
          className="border rounded-lg p-4 flex items-center gap-4"
        >
          <SkeletonLine className="size-4 rounded-full shrink-0" />
          <div className="flex-1 space-y-2">
            <SkeletonLine className="h-4 w-3/5" />
            <SkeletonLine className="h-3 w-2/5" />
          </div>
          <SkeletonLine className="h-6 w-16 rounded-full" />
        </div>
      ))}
    </div>
  );
}

export function ChatSkeleton() {
  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="border-b p-4">
        <SkeletonLine className="h-6 w-36" />
      </div>

      {/* Messages */}
      <div className="flex-1 p-4 space-y-4">
        {/* User message */}
        <div className="flex justify-end">
          <SkeletonLine className="h-10 w-48 rounded-2xl" />
        </div>

        {/* AI messages */}
        {Array.from({ length: 3 }).map((_, i) => (
          <div key={i} className="flex justify-start">
            <div className="space-y-1">
              <SkeletonLine className="h-4 w-64" />
              <SkeletonLine className="h-4 w-48" />
            </div>
          </div>
        ))}

        {/* Typing indicator */}
        <div className="flex justify-start gap-1">
          <SkeletonLine className="size-2 rounded-full" />
          <SkeletonLine className="size-2 rounded-full" />
          <SkeletonLine className="size-2 rounded-full" />
        </div>
      </div>

      {/* Input */}
      <div className="border-t p-4">
        <SkeletonLine className="h-10 w-full rounded-lg" />
      </div>
    </div>
  );
}
