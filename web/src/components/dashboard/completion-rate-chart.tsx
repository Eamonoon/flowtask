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

interface CompletionRateChartProps {
  className?: string;
}

interface ChartPoint {
  label: string;
  rate: number;
}

export function CompletionRateChart({ className }: CompletionRateChartProps) {
  const [range, setRange] = useState<TimeRange>('week');

  const { data, isLoading } = useQuery<ChartPoint[]>({
    queryKey: ['dashboard', 'completion-rate', range],
    queryFn: async () => {
      const { data } = await api.get<ApiResponse<ChartData>>(
        '/dashboard/charts/completion-rate',
        { params: { range } }
      );
      return data.data.labels.map((label, i) => ({
        label,
        rate: data.data.values[i],
      }));
    },
  });

  return (
    <div className={cn('rounded-lg border bg-card p-4 shadow-sm', className)}>
      <div className="mb-4 flex items-center justify-between">
        <h3 className="text-sm font-semibold text-foreground">完成率趋势</h3>
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
              domain={[0, 1]}
              tickFormatter={(value: number) => `${Math.round(value * 100)}%`}
            />
            <Tooltip
              contentStyle={{
                borderRadius: '8px',
                border: '1px solid hsl(var(--border))',
                background: 'hsl(var(--card))',
                fontSize: '12px',
              }}
              formatter={(value) => [`${Math.round((value as number) * 100)}%`, '完成率']}
              labelFormatter={(label) => String(label)}
            />
            <Line
              type="monotone"
              dataKey="rate"
              stroke="#10b981"
              strokeWidth={2}
              dot={{ r: 4, fill: '#10b981' }}
              activeDot={{ r: 6 }}
            />
          </LineChart>
        </ResponsiveContainer>
      )}
    </div>
  );
}
