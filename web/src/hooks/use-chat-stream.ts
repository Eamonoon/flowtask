'use client';

import { useState, useRef, useCallback } from 'react';
import { useAuthStore } from '@/stores/auth-store';
import api from '@/lib/api';
import type { ApiResponse } from '@/types/api';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';

export interface TaskTreeNode {
  title: string;
  description?: string;
  priority?: 'low' | 'medium' | 'high' | 'urgent';
  estimated_duration?: string;
  children?: TaskTreeNode[];
}

export interface UseChatStreamOptions {
  conversationId: string | null;
  learningGoalId?: string | null;
  mode?: 'chat' | 'task-breakdown';
  onConversationCreated?: (id: string) => void;
  onDone?: (fullContent: string) => void;
}

export interface UseChatStreamReturn {
  sendMessage: (message: string) => void;
  isStreaming: boolean;
  currentResponse: string;
  taskTree: TaskTreeNode[] | null;
  saveAsTasks: (goalId: string) => Promise<void>;
}

export function useChatStream({
  conversationId,
  learningGoalId,
  mode = 'chat',
  onConversationCreated,
  onDone,
}: UseChatStreamOptions): UseChatStreamReturn {
  const [isStreaming, setIsStreaming] = useState(false);
  const [currentResponse, setCurrentResponse] = useState('');
  const [taskTree, setTaskTree] = useState<TaskTreeNode[] | null>(null);
  const abortControllerRef = useRef<AbortController | null>(null);

  const sendMessage = useCallback(
    async (message: string) => {
      if (isStreaming) return;

      setIsStreaming(true);
      setCurrentResponse('');
      setTaskTree(null);

      const abortController = new AbortController();
      abortControllerRef.current = abortController;

      try {
        const endpoint =
          mode === 'task-breakdown' ? '/ai/task-breakdown' : '/ai/chat';
        const url = `${API_URL}${endpoint}`;

        const token = useAuthStore.getState().accessToken;

        const body: Record<string, unknown> = { message };
        if (conversationId) body.conversation_id = conversationId;
        if (learningGoalId) body.learning_goal_id = learningGoalId;
        if (mode === 'task-breakdown') body.save_to_goal = false;

        const response = await fetch(url, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            ...(token ? { Authorization: `Bearer ${token}` } : {}),
          },
          body: JSON.stringify(body),
          signal: abortController.signal,
        });

        if (!response.ok) {
          throw new Error(`HTTP ${response.status}`);
        }

        const reader = response.body?.getReader();
        if (!reader) throw new Error('No readable stream');

        const decoder = new TextDecoder();
        let buffer = '';
        let fullContent = '';

        while (true) {
          const { done, value } = await reader.read();
          if (done) break;

          buffer += decoder.decode(value, { stream: true });

          // Process complete SSE events
          const events = buffer.split('\n\n');
          buffer = events.pop() || ''; // keep incomplete chunk

          for (const eventBlock of events) {
            const lines = eventBlock.split('\n');
            let eventType = 'message';
            let data = '';

            for (const line of lines) {
              if (line.startsWith('event: ')) {
                eventType = line.slice(7);
              } else if (line.startsWith('data: ')) {
                data = line.slice(6);
              }
            }

            if (!data) continue;

            try {
              const parsed = JSON.parse(data);

              switch (eventType) {
                case 'conversation':
                  if (parsed.conversation_id && onConversationCreated) {
                    onConversationCreated(parsed.conversation_id);
                  }
                  break;

                case 'delta':
                  fullContent += parsed.content || '';
                  setCurrentResponse(fullContent);
                  break;

                case 'done':
                  fullContent = parsed.full_content || fullContent;
                  setCurrentResponse(fullContent);
                  // Notify conversation created (backend sends conversation_id in done event)
                  if (parsed.conversation_id && onConversationCreated) {
                    onConversationCreated(parsed.conversation_id);
                  }
                  // Try to parse task tree from response
                  const parsedTasks = tryParseTaskTree(fullContent);
                  if (parsedTasks) {
                    setTaskTree(parsedTasks);
                  }
                  onDone?.(fullContent);
                  break;
              }
            } catch {
              // Skip malformed JSON
            }
          }
        }
      } catch (err) {
        if ((err as Error).name === 'AbortError') return;
        console.error('Chat stream error:', err);
        setCurrentResponse((prev) =>
          prev || '请求失败，请稍后重试。'
        );
      } finally {
        setIsStreaming(false);
        abortControllerRef.current = null;
      }
    },
    [
      isStreaming,
      conversationId,
      learningGoalId,
      mode,
      onConversationCreated,
      onDone,
    ]
  );

  const saveAsTasks = useCallback(
    async (goalId: string) => {
      if (!taskTree || taskTree.length === 0) {
        throw new Error('没有可保存的任务');
      }

      // Flatten the tree and POST each task
      const tasks = flattenTaskTree(taskTree, goalId);

      for (const task of tasks) {
        await api.post<ApiResponse<unknown>>(
          `/learning-goals/${goalId}/tasks`,
          task
        );
      }
    },
    [taskTree]
  );

  return {
    sendMessage,
    isStreaming,
    currentResponse,
    taskTree,
    saveAsTasks,
  };
}

/**
 * Try to parse a JSON task tree from the AI response content.
 * Looks for a JSON block wrapped in ```json ... ``` or standalone JSON array.
 */
function tryParseTaskTree(content: string): TaskTreeNode[] | null {
  // Try to find JSON in code block
  const codeBlockMatch = content.match(/```json\s*([\s\S]*?)```/);
  const jsonStr = codeBlockMatch
    ? codeBlockMatch[1].trim()
    : content.trim();

  try {
    const parsed = JSON.parse(jsonStr);
    if (Array.isArray(parsed) && parsed.length > 0 && parsed[0].title) {
      return parsed as TaskTreeNode[];
    }
  } catch {
    // Not valid JSON
  }
  return null;
}

interface FlatTask {
  title: string;
  description: string;
  priority: string;
  estimated_duration: string | null;
  parent_task_id: string | null;
  learning_goal_id: string;
}

/**
 * Flatten a nested task tree into an array suitable for sequential API calls.
 * Note: since we don't know parent IDs before creation, we keep the tree structure
 * and rely on the backend to handle parent_task_id linking, or we pass the raw tree.
 */
function flattenTaskTree(
  nodes: TaskTreeNode[],
  goalId: string,
  parentId: string | null = null
): FlatTask[] {
  const tasks: FlatTask[] = [];
  for (const node of nodes) {
    tasks.push({
      title: node.title,
      description: node.description || '',
      priority: node.priority || 'medium',
      estimated_duration: node.estimated_duration || null,
      parent_task_id: parentId,
      learning_goal_id: goalId,
    });
    if (node.children && node.children.length > 0) {
      // Children will be saved with parent_task_id = null (needs backend support)
      // The task tree is posted as a flat list; backend should link them
      tasks.push(...flattenTaskTree(node.children, goalId, null));
    }
  }
  return tasks;
}
