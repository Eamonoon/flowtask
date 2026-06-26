'use client';

import { cn } from '@/lib/utils';

type Trend = 'up' | 'down' | 'neutral';

interface StatsCardProps {
  title: string;
  value: string | number;
  subtitle?: string;
  icon?: React.ReactNode;
  trend?: Trend;
  trendValue?: string;
  className?: string;
}

const TREND_CONFIG: Record<Trend, { color: string; icon: React.ReactNode }> = {
  up: {
    color: 'text-emerald-600 dark:text-emerald-400',
    icon: (
      <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M4.5 19.5l15-15m0 0H8.25m11.25 0v11.25" />
      </svg>
    ),
  },
  down: {
    color: 'text-red-500 dark:text-red-400',
    icon: (
      <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M4.5 4.5l15 15m0 0V8.25m0 11.25H8.25" />
      </svg>
    ),
  },
  neutral: {
    color: 'text-muted-foreground',
    icon: (
      <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 12h16.5" />
      </svg>
    ),
  },
};

export function StatsCard({
  title,
  value,
  subtitle,
  icon,
  trend,
  trendValue,
  className,
}: StatsCardProps) {
  const trendConfig = trend ? TREND_CONFIG[trend] : null;

  return (
    <div
      className={cn(
        'rounded-lg border bg-card p-4 shadow-sm',
        'transition-shadow hover:shadow-md',
        className
      )}
    >
      <div className="flex items-center justify-between">
        <span className="text-sm font-medium text-muted-foreground">{title}</span>
        {icon && (
          <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-muted text-muted-foreground">
            {icon}
          </div>
        )}
      </div>

      <div className="mt-2">
        <p className="text-2xl font-bold text-foreground">{value}</p>
        <div className="mt-1 flex items-center gap-1.5">
          {trendConfig && (
            <span className={cn('inline-flex items-center gap-0.5 text-xs font-medium', trendConfig.color)}>
              {trendConfig.icon}
              {trendValue && <span>{trendValue}</span>}
            </span>
          )}
          {subtitle && (
            <span className="text-xs text-muted-foreground">{subtitle}</span>
          )}
        </div>
      </div>
    </div>
  );
}
