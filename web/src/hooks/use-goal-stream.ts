'use client';

import { useCallback, useRef, useEffect } from 'react';
import { useGoalStore } from '@/stores/goal-store';
import { useAuthStore } from '@/stores/auth-store';
import api from '@/lib/api';
import { saveProgress, loadProgress, clearProgress } from '@/lib/goal-progress';
import type { ApiResponse } from '@/types/api';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';

export function useGoalStream() {
  const {
    setPhase,
    addTask,
    setTasks,
    setError,
    setSessionId,
    setLearningGoalId,
    reset,
  } = useGoalStore();

  const { accessToken } = useAuthStore();
  const abortControllerRef = useRef<AbortController | null>(null);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      closeStream();
    };
  }, []);

  const closeStream = useCallback(() => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
      abortControllerRef.current = null;
    }
  }, []);

  const createGoal = useCallback(
    async (description: string, targetDuration?: string) => {
      try {
        setPhase('connecting');
        reset();

        const { data } = await api.post<
          ApiResponse<{ session_id: string; learning_goal_id: string }>
        >('/learning-goals', {
          description,
          target_duration: targetDuration,
        });

        const { session_id, learning_goal_id } = data.data;
        setSessionId(session_id);
        setLearningGoalId(learning_goal_id);

        return { sessionId: session_id, learningGoalId: learning_goal_id };
      } catch (err: unknown) {
        const axiosErr = err as { response?: { data?: { message?: string } } };
        const message = axiosErr.response?.data?.message || '创建学习目标失败';
        setError(message);
        throw err;
      }
    },
    [setPhase, reset, setSessionId, setLearningGoalId, setError]
  );

  const startStream = useCallback(
    async (learningGoalId: string, sessionId: string) => {
      closeStream();
      setPhase('streaming');

      const url = `${API_URL}/learning-goals/${learningGoalId}/generate-stream?session_id=${sessionId}`;
      const token = useAuthStore.getState().accessToken;

      const abortController = new AbortController();
      abortControllerRef.current = abortController;

      try {
        const response = await fetch(url, {
          method: 'GET',
          headers: {
            'Authorization': token ? `Bearer ${token}` : '',
            'Accept': 'text/event-stream',
          },
          signal: abortController.signal,
        });

        if (!response.ok) {
          const errorData = await response.json();
          setError(errorData.message || '连接失败，请重试');
          return;
        }

        const reader = response.body?.getReader();
        if (!reader) {
          setError('无法读取响应流');
          return;
        }

        const decoder = new TextDecoder();
        let buffer = '';

        while (true) {
          const { done, value } = await reader.read();
          if (done) break;

          buffer += decoder.decode(value, { stream: true });
          const lines = buffer.split('\n');
          buffer = lines.pop() || '';

          for (const line of lines) {
            const trimmed = line.trim();
            if (!trimmed) continue;

            // Parse SSE event
            if (trimmed.startsWith('event:')) {
              const eventType = trimmed.slice(6).trim();
              // Next line should be data
              continue;
            }

            if (trimmed.startsWith('data:')) {
              const dataStr = trimmed.slice(5).trim();

              try {
                const data = JSON.parse(dataStr);

                // Determine event type from previous event line or data structure
                if (data.id && data.title) {
                  // Task event
                  addTask(data);

                  // Save progress to localStorage after each task
                  const { generatedTasks } = useGoalStore.getState();
                  saveProgress(learningGoalId, generatedTasks, sessionId, learningGoalId);
                } else if (data.task_count !== undefined && !data.learning_goal_id) {
                  // Progress event - handled by taskCount in store
                } else if (data.learning_goal_id && data.task_count !== undefined) {
                  // Done event
                  setPhase('preview');
                  closeStream();
                  return;
                } else if (data.code && data.message) {
                  // Error event
                  setError(data.message || '生成失败，请重试');
                  closeStream();
                  return;
                }
              } catch (parseErr) {
                console.error('Failed to parse SSE data:', parseErr);
              }
            }
          }
        }

        // Stream ended normally
        setPhase('preview');
      } catch (err: unknown) {
        if (err instanceof Error && err.name === 'AbortError') {
          // Stream was aborted, ignore
          return;
        }
        console.error('Stream error:', err);
        setError('连接中断，请重试');
      } finally {
        abortControllerRef.current = null;
      }
    },
    [closeStream, setPhase, addTask, setError]
  );

  const confirmSave = useCallback(
    async (learningGoalId: string, sessionId: string) => {
      try {
        setPhase('connecting');

        const { generatedTasks } = useGoalStore.getState();
        const { data } = await api.post<ApiResponse<{ saved_task_count: number }>>(
          `/learning-goals/${learningGoalId}/tasks/confirm`,
          {
            session_id: sessionId,
            tasks: generatedTasks,
          }
        );

        // Clear progress from localStorage after successful save
        clearProgress(learningGoalId);

        setPhase('done');
        return data.data;
      } catch (err: unknown) {
        const axiosErr = err as { response?: { data?: { message?: string } } };
        const message = axiosErr.response?.data?.message || '保存失败，请重试';
        setError(message);
        throw err;
      }
    },
    [setPhase, setError]
  );

  const regenerate = useCallback(
    async (learningGoalId: string) => {
      try {
        closeStream();
        clearProgress(learningGoalId);
        reset();
        setPhase('connecting');

        const { data } = await api.post<
          ApiResponse<{ session_id: string; learning_goal_id: string }>
        >(`/learning-goals/${learningGoalId}/regenerate`);

        const { session_id, learning_goal_id } = data.data;
        setSessionId(session_id);
        setLearningGoalId(learning_goal_id);

        return { sessionId: session_id, learningGoalId: learning_goal_id };
      } catch (err: unknown) {
        const axiosErr = err as { response?: { data?: { message?: string } } };
        const message = axiosErr.response?.data?.message || '重新生成失败';
        setError(message);
        throw err;
      }
    },
    [closeStream, reset, setPhase, setSessionId, setLearningGoalId, setError]
  );

  const continueStream = useCallback(
    (learningGoalId: string, sessionId: string) => {
      startStream(learningGoalId, sessionId);
    },
    [startStream]
  );

  const restoreProgress = useCallback(
    (goalId: string) => {
      const saved = loadProgress(goalId);
      if (saved && saved.tasks.length > 0) {
        setTasks(saved.tasks);
        setSessionId(saved.sessionId);
        setLearningGoalId(saved.learningGoalId);
        setPhase('preview');
        return true;
      }
      return false;
    },
    [setTasks, setSessionId, setLearningGoalId, setPhase]
  );

  return {
    createGoal,
    startStream,
    confirmSave,
    regenerate,
    continueStream,
    restoreProgress,
    closeStream,
  };
}
