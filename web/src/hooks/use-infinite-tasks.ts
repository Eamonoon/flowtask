'use client';

import { useInfiniteQuery } from '@tanstack/react-query';
import api from '@/lib/api';
import type { Task, ApiResponse, PaginatedResponse } from '@/types/api';
import type { TaskFilters } from '@/components/task/task-filters';

interface UseInfiniteTasksOptions {
  filters?: TaskFilters;
  pageSize?: number;
}

export function useInfiniteTasks({ filters, pageSize = 20 }: UseInfiniteTasksOptions = {}) {
  return useInfiniteQuery({
    queryKey: ['tasks', 'infinite', filters],
    queryFn: async ({ pageParam }) => {
      const params: Record<string, string | number | string[]> = {
        limit: pageSize,
      };

      if (pageParam) {
        params.cursor = pageParam;
      }

      if (filters?.keyword) {
        params.search = filters.keyword;
      }
      if (filters?.statuses && filters.statuses.length > 0) {
        params.status = filters.statuses;
      }
      if (filters?.priorities && filters.priorities.length > 0) {
        params.priority = filters.priorities;
      }
      if (filters?.labelIds && filters.labelIds.length > 0) {
        params.label_ids = filters.labelIds;
      }
      if (filters?.deadlineFrom) {
        params.deadline_from = filters.deadlineFrom;
      }
      if (filters?.deadlineTo) {
        params.deadline_to = filters.deadlineTo;
      }
      if (filters?.sortBy) {
        params.sort_by = filters.sortBy;
      }
      if (filters?.sortOrder) {
        params.sort_order = filters.sortOrder;
      }

      const { data } = await api.get<ApiResponse<PaginatedResponse<Task>>>('/tasks', {
        params,
      });
      return data.data;
    },
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (lastPage) => {
      if (!lastPage.has_more || !lastPage.next_cursor) {
        return undefined;
      }
      return lastPage.next_cursor;
    },
  });
}
