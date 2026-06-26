'use client';

import { useEffect, useState, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useAuthStore } from '@/stores/auth-store';
import { useGoalStore } from '@/stores/goal-store';
import { useGoalStream } from '@/hooks/use-goal-stream';
import { StreamingPlanViewer } from '@/components/goal/streaming-plan-viewer';
import { Button } from '@/components/ui/button';
import api from '@/lib/api';
import type { ApiResponse, LearningGoal, PaginatedResponse } from '@/types/api';

export default function GoalsPage() {
  const { isAuthenticated } = useAuthStore();
  const queryClient = useQueryClient();
  const {
    generationPhase,
    generatedTasks,
    errorMessage,
    taskCount,
    sessionId,
    learningGoalId,
  } = useGoalStore();
  const { createGoal, startStream, confirmSave, regenerate, restoreProgress, closeStream } =
    useGoalStream();
  const router = useRouter();

  const [input, setInput] = useState('');
  const [targetDuration, setTargetDuration] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [hasRestoredProgress, setHasRestoredProgress] = useState(false);

  const { data: goalsData } = useQuery({
    queryKey: ['learning-goals'],
    queryFn: async () => {
      const { data } = await api.get<ApiResponse<PaginatedResponse<LearningGoal>>>('/learning-goals');
      return data.data?.items ?? [];
    },
    enabled: isAuthenticated,
  });

  useEffect(() => {
    if (!isAuthenticated) router.push('/login');
  }, [isAuthenticated, router]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      closeStream();
    };
  }, [closeStream]);

  // Try to restore progress on mount
  useEffect(() => {
    if (learningGoalId && !hasRestoredProgress) {
      const restored = restoreProgress(learningGoalId);
      setHasRestoredProgress(true);
    }
  }, [learningGoalId, hasRestoredProgress, restoreProgress]);

  const handleGenerate = useCallback(async () => {
    if (!input.trim() || isSubmitting) return;

    setIsSubmitting(true);
    try {
      const { sessionId: newSessionId, learningGoalId: newGoalId } =
        await createGoal(input, targetDuration || undefined);

      // Start streaming after goal is created
      startStream(newGoalId, newSessionId);
    } catch (err) {
      // Error is handled in useGoalStream
    } finally {
      setIsSubmitting(false);
    }
  }, [input, targetDuration, isSubmitting, createGoal, startStream]);

  const handleConfirmSave = useCallback(async () => {
    if (!learningGoalId || !sessionId) return;

    try {
      await confirmSave(learningGoalId, sessionId);
      queryClient.invalidateQueries({ queryKey: ['learning-goals'] });
    } catch (err) {
      // Error is handled in useGoalStream
    }
  }, [learningGoalId, sessionId, confirmSave, queryClient]);

  const handleRegenerate = useCallback(async () => {
    if (!learningGoalId) return;

    try {
      const { sessionId: newSessionId, learningGoalId: newGoalId } =
        await regenerate(learningGoalId);

      startStream(newGoalId, newSessionId);
    } catch (err) {
      // Error is handled in useGoalStream
    }
  }, [learningGoalId, regenerate, startStream]);

  if (!isAuthenticated) return null;

  return (
    <div className="min-h-screen bg-background">
      <nav className="border-b px-6 py-4">
        <div className="max-w-7xl mx-auto flex justify-between items-center">
          <h1 className="text-xl font-bold">FlowTask</h1>
          <div className="flex gap-4 items-center">
            <Link href="/dashboard">Dashboard</Link>
            <Link href="/tasks">任务</Link>
            <Link href="/goals" className="font-semibold text-primary">
              学习目标
            </Link>
            <Link href="/chat">AI 助手</Link>
          </div>
        </div>
      </nav>

      <main className="max-w-4xl mx-auto p-6">
        <h2 className="text-2xl font-bold mb-6">创建学习目标</h2>

        {/* Input Form */}
        <div className="border rounded-lg p-6 mb-8">
          <label className="block text-sm font-medium mb-2">
            用自然语言描述你的学习目标
          </label>
          <textarea
            value={input}
            onChange={(e) => setInput(e.target.value)}
            className="w-full px-3 py-2 border rounded-md h-24"
            placeholder="例如：我想两个月学会 RAG"
            disabled={generationPhase === 'connecting' || generationPhase === 'streaming'}
          />
          <div className="flex gap-4 mt-4">
            <input
              type="text"
              value={targetDuration}
              onChange={(e) => setTargetDuration(e.target.value)}
              placeholder="目标时长（可选）"
              className="px-3 py-2 border rounded-md w-48"
              disabled={generationPhase === 'connecting' || generationPhase === 'streaming'}
            />
            <Button
              onClick={handleGenerate}
              disabled={!input.trim() || isSubmitting || generationPhase === 'connecting' || generationPhase === 'streaming'}
            >
              {generationPhase === 'connecting' ? '创建中...' : '生成学习计划'}
            </Button>
          </div>
          {errorMessage && (
            <p className="text-red-500 text-sm mt-2">{errorMessage}</p>
          )}
        </div>

        {/* Streaming Plan Viewer */}
        {generationPhase !== 'idle' && (
          <StreamingPlanViewer
            phase={generationPhase}
            tasks={generatedTasks}
            taskCount={taskCount}
            onConfirmSave={handleConfirmSave}
            onRegenerate={handleRegenerate}
          />
        )}

        {/* My Learning Goals */}
        <h3 className="text-xl font-semibold mb-4 mt-8">我的学习目标</h3>
        <div className="space-y-4">
          {!goalsData || goalsData.length === 0 ? (
            <p className="text-muted-foreground">暂无学习目标，创建一个开始吧！</p>
          ) : (
            goalsData.map((goal) => (
              <div
                key={goal.id}
                className="border rounded-lg p-4 hover:shadow-md transition-shadow"
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1 min-w-0">
                    <p className="font-medium truncate">{goal.description}</p>
                    <div className="flex items-center gap-4 mt-2 text-sm text-muted-foreground">
                      {goal.target_duration && (
                        <span>目标时长: {goal.target_duration}</span>
                      )}
                      <span>
                        任务: {goal.completed_task_count}/{goal.task_count} 已完成
                      </span>
                      <span
                        className={`inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium ${
                          goal.status === 'active'
                            ? 'bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-400'
                            : goal.status === 'completed'
                            ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-400'
                            : 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'
                        }`}
                      >
                        {goal.status === 'active'
                          ? '进行中'
                          : goal.status === 'completed'
                          ? '已完成'
                          : goal.status}
                      </span>
                    </div>
                    {goal.task_count > 0 && (
                      <div className="mt-2 h-1.5 w-full rounded-full bg-muted overflow-hidden">
                        <div
                          className="h-full rounded-full bg-primary transition-all"
                          style={{
                            width: `${Math.round((goal.completed_task_count / goal.task_count) * 100)}%`,
                          }}
                        />
                      </div>
                    )}
                  </div>
                  <span className="text-xs text-muted-foreground ml-4 shrink-0">
                    {new Date(goal.created_at).toLocaleDateString('zh-CN')}
                  </span>
                </div>
              </div>
            ))
          )}
        </div>
      </main>
    </div>
  );
}
