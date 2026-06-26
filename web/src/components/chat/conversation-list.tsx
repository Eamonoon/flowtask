'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import api from '@/lib/api';
import { cn } from '@/lib/utils';
import type { ApiResponse, AIConversation } from '@/types/api';
import { Button } from '@/components/ui/button';
import { Plus, MessageSquare, Trash2 } from 'lucide-react';

interface ConversationListProps {
  activeConversationId: string | null;
  onSelect: (id: string) => void;
  onNew: () => void;
  onDelete?: (id: string) => void;
}

export function ConversationList({
  activeConversationId,
  onSelect,
  onNew,
  onDelete,
}: ConversationListProps) {
  const queryClient = useQueryClient();

  const { data: conversations = [], isLoading } = useQuery({
    queryKey: ['ai', 'conversations'],
    queryFn: async () => {
      const { data } = await api.get<ApiResponse<AIConversation[]>>(
        '/ai/conversations'
      );
      return data.data;
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async (id: string) => {
      await api.delete(`/ai/conversations/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ai', 'conversations'] });
    },
  });

  const handleDelete = (e: React.MouseEvent, id: string) => {
    e.stopPropagation();
    deleteMutation.mutate(id);
    onDelete?.(id);
  };

  return (
    <aside className="w-64 border-r flex flex-col h-full shrink-0">
      <div className="p-3">
        <Button className="w-full" size="sm" onClick={onNew}>
          <Plus className="size-4 mr-1.5" />
          新对话
        </Button>
      </div>

      <div className="flex-1 overflow-y-auto px-2 pb-2 space-y-0.5">
        {isLoading && (
          <p className="text-sm text-muted-foreground text-center py-4">
            加载中...
          </p>
        )}

        {!isLoading && conversations.length === 0 && (
          <p className="text-sm text-muted-foreground text-center py-4">
            暂无对话
          </p>
        )}

        {conversations.map((conv) => (
          <button
            key={conv.id}
            onClick={() => onSelect(conv.id)}
            className={cn(
              'w-full text-left px-3 py-2 rounded-md text-sm transition-colors truncate group relative',
              'hover:bg-muted',
              activeConversationId === conv.id
                ? 'bg-muted font-medium text-foreground'
                : 'text-muted-foreground'
            )}
          >
            <div className="flex items-center gap-2">
              <MessageSquare className="size-3.5 shrink-0" />
              <span className="truncate flex-1">
                {conv.title || '新对话'}
              </span>
              <span
                role="button"
                tabIndex={0}
                onClick={(e) => handleDelete(e, conv.id)}
                onKeyDown={(e) => { if (e.key === 'Enter') handleDelete(e, conv.id); }}
                className="opacity-0 group-hover:opacity-100 text-muted-foreground hover:text-destructive transition-opacity shrink-0 cursor-pointer"
                title="删除对话"
              >
                <Trash2 className="size-3.5" />
              </span>
            </div>
            <span className="text-xs text-muted-foreground/70 ml-5.5 block">
              {formatTime(conv.updated_at)}
            </span>
          </button>
        ))}
      </div>
    </aside>
  );
}

function formatTime(iso: string): string {
  const date = new Date(iso);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60_000);
  const diffHours = Math.floor(diffMs / 3_600_000);
  const diffDays = Math.floor(diffMs / 86_400_000);

  if (diffMins < 1) return '刚刚';
  if (diffMins < 60) return `${diffMins} 分钟前`;
  if (diffHours < 24) return `${diffHours} 小时前`;
  if (diffDays < 7) return `${diffDays} 天前`;

  return date.toLocaleDateString('zh-CN', {
    month: 'short',
    day: 'numeric',
  });
}
