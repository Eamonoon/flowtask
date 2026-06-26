'use client';

import { useQuery } from '@tanstack/react-query';
import {
  PieChart,
  Pie,
  Cell,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';
import { cn } from '@/lib/utils';
import api from '@/lib/api';
import type { ApiResponse } from '@/types/api';

interface CategoryChartProps {
  className?: string;
}

interface CategoryDataPoint {
  name: string;
  value: number;
}

const COLORS = [
  'hsl(var(--primary))',
  '#6366f1',
  '#f59e0b',
  '#10b981',
  '#ef4444',
  '#8b5cf6',
  '#06b6d4',
  '#ec4899',
  '#84cc16',
  '#f97316',
];

export function CategoryChart({ className }: CategoryChartProps) {
  const { data, isLoading } = useQuery<CategoryDataPoint[]>({
    queryKey: ['dashboard', 'category-stats'],
    queryFn: async () => {
      const { data } = await api.get<ApiResponse<CategoryDataPoint[]>>(
        '/dashboard/charts/category-stats'
      );
      return data.data;
    },
  });

  return (
    <div className={cn('rounded-lg border bg-card p-4 shadow-sm', className)}>
      <h3 className="mb-4 text-sm font-semibold text-foreground">任务分类分布</h3>

      {isLoading ? (
        <div className="flex h-[280px] items-center justify-center">
          <div className="h-5 w-5 animate-spin rounded-full border-2 border-primary border-t-transparent" />
        </div>
      ) : data && data.length > 0 ? (
        <ResponsiveContainer width="100%" height={280}>
          <PieChart>
            <Pie
              data={data}
              cx="50%"
              cy="50%"
              innerRadius={55}
              outerRadius={90}
              paddingAngle={3}
              dataKey="value"
              nameKey="name"
            >
              {data.map((_, index) => (
                <Cell
                  key={`cell-${index}`}
                  fill={COLORS[index % COLORS.length]}
                />
              ))}
            </Pie>
            <Tooltip
              contentStyle={{
                borderRadius: '8px',
                border: '1px solid hsl(var(--border))',
                background: 'hsl(var(--card))',
                fontSize: '12px',
              }}
              formatter={(value, name) => [`${value} 个任务`, name]}
            />
            <Legend
              iconType="circle"
              iconSize={8}
              formatter={(value: string) => (
                <span className="text-xs text-muted-foreground">{value}</span>
              )}
            />
          </PieChart>
        </ResponsiveContainer>
      ) : (
        <div className="flex h-[280px] items-center justify-center">
          <p className="text-sm text-muted-foreground">暂无数据</p>
        </div>
      )}
    </div>
  );
}
