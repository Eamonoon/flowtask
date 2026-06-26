'use client';

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import api from '@/lib/api';
import type { ApiResponse, ChartData } from '@/types/api';

type TimeRange = 'week' | 'month';

interface StudyTimeChartProps {
  className?: string;
}

interface ChartPoint {
  label: string;
  minutes: number;
}

export function StudyTimeChart({ className }: StudyTimeChartProps) {
  const [range, setRange] = useState<TimeRange>('week');

  const { data, isLoading } = useQuery<ChartPoint[]>({
    queryKey: ['dashboard', 'study-time', range],
    queryFn: async () => {
      const { data } = await api.get<ApiResponse<ChartData>>(
        `/dashboard/charts/study-time`,
        { params: { range } }
      );
      return data.data.labels.map((label, i) => ({
        label,
        minutes: data.data.values[i],
      }));
    },
  });

  return (
    <div className={cn('rounded-lg border bg-card p-4 shadow-sm', className)}>
      <div className="mb-4 flex items-center justify-between">
        <h3 className="text-sm font-semibold text-foreground">学习时长</h3>
        <div className="flex gap-1 rounded-lg bg-muted p-0.5">
          <Button
            variant={range === 'week' ? 'default' : 'ghost'}
            size="xs"
            onClick={() => setRange('week')}
          >
            本周
          </Button>
          <Button
            variant={range === 'month' ? 'default' : 'ghost'}
            size="xs"
            onClick={() => setRange('month')}
          >
            本月
          </Button>
        </div>
      </div>

      {isLoading ? (
        <div className="flex h-[250px] items-center justify-center">
          <div className="h-5 w-5 animate-spin rounded-full border-2 border-primary border-t-transparent" />
        </div>
      ) : (
        <ResponsiveContainer width="100%" height={250}>
          <LineChart data={data ?? []}>
            <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
            <XAxis
              dataKey="label"
              tick={{ fontSize: 12 }}
              className="text-muted-foreground"
              tickLine={false}
              axisLine={false}
            />
            <YAxis
              tick={{ fontSize: 12 }}
              className="text-muted-foreground"
              tickLine={false}
              axisLine={false}
              tickFormatter={(value: number) => `${value}分钟`}
            />
            <Tooltip
              contentStyle={{
                borderRadius: '8px',
                border: '1px solid hsl(var(--border))',
                background: 'hsl(var(--card))',
                fontSize: '12px',
              }}
              formatter={(value) => [`${value} 分钟`, '学习时长']}
              labelFormatter={(label) => String(label)}
            />
            <Line
              type="monotone"
              dataKey="minutes"
              stroke="hsl(var(--primary))"
              strokeWidth={2}
              dot={{ r: 4, fill: 'hsl(var(--primary))' }}
              activeDot={{ r: 6 }}
            />
          </LineChart>
        </ResponsiveContainer>
      )}
    </div>
  );
}
