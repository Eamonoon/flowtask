import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { StreamingPlanViewer } from '@/components/goal/streaming-plan-viewer';

describe('Error Scenario Coverage', () => {
  const mockConfirmSave = vi.fn();
  const mockRegenerate = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should display error message when phase is error', () => {
    render(
      <StreamingPlanViewer
        phase="error"
        tasks={[]}
        taskCount={0}
        onConfirmSave={mockConfirmSave}
        onRegenerate={mockRegenerate}
      />
    );

    expect(screen.getByText('生成失败，请重试')).toBeInTheDocument();
  });

  it('should show retry button on error', () => {
    render(
      <StreamingPlanViewer
        phase="error"
        tasks={[]}
        taskCount={0}
        onConfirmSave={mockConfirmSave}
        onRegenerate={mockRegenerate}
      />
    );

    const retryButton = screen.getByText('重试');
    expect(retryButton).toBeInTheDocument();
  });

  it('should call onRegenerate when retry button is clicked', async () => {
    mockRegenerate.mockResolvedValue(undefined);

    render(
      <StreamingPlanViewer
        phase="error"
        tasks={[]}
        taskCount={0}
        onConfirmSave={mockConfirmSave}
        onRegenerate={mockRegenerate}
      />
    );

    const retryButton = screen.getByText('重试');
    fireEvent.click(retryButton);

    await waitFor(() => {
      expect(mockRegenerate).toHaveBeenCalledTimes(1);
    });
  });

  it('should display connecting state with loading indicator', () => {
    render(
      <StreamingPlanViewer
        phase="connecting"
        tasks={[]}
        taskCount={0}
        onConfirmSave={mockConfirmSave}
        onRegenerate={mockRegenerate}
      />
    );

    expect(screen.getByText('正在连接 AI 服务...')).toBeInTheDocument();
  });

  it('should display streaming state with task count', () => {
    const tasks = [
      { id: '1', title: 'Task 1', description: 'Desc 1' },
      { id: '2', title: 'Task 2', description: 'Desc 2' },
    ];

    render(
      <StreamingPlanViewer
        phase="streaming"
        tasks={tasks}
        taskCount={2}
        onConfirmSave={mockConfirmSave}
        onRegenerate={mockRegenerate}
      />
    );

    expect(screen.getByText('已生成 2 个任务...')).toBeInTheDocument();
    expect(screen.getByText('Task 1')).toBeInTheDocument();
    expect(screen.getByText('Task 2')).toBeInTheDocument();
  });

  it('should display preview state with confirm and regenerate buttons', () => {
    const tasks = [{ id: '1', title: 'Task 1' }];

    render(
      <StreamingPlanViewer
        phase="preview"
        tasks={tasks}
        taskCount={1}
        onConfirmSave={mockConfirmSave}
        onRegenerate={mockRegenerate}
      />
    );

    expect(screen.getByText('确认保存')).toBeInTheDocument();
    expect(screen.getByText('重新生成')).toBeInTheDocument();
  });

  it('should call onConfirmSave when confirm button is clicked', async () => {
    mockConfirmSave.mockResolvedValue(undefined);
    const tasks = [{ id: '1', title: 'Task 1' }];

    render(
      <StreamingPlanViewer
        phase="preview"
        tasks={tasks}
        taskCount={1}
        onConfirmSave={mockConfirmSave}
        onRegenerate={mockRegenerate}
      />
    );

    const confirmButton = screen.getByText('确认保存');
    fireEvent.click(confirmButton);

    await waitFor(() => {
      expect(mockConfirmSave).toHaveBeenCalledTimes(1);
    });
  });

  it('should display success state after save', () => {
    const tasks = [{ id: '1', title: 'Task 1' }];

    render(
      <StreamingPlanViewer
        phase="done"
        tasks={tasks}
        taskCount={1}
        onConfirmSave={mockConfirmSave}
        onRegenerate={mockRegenerate}
      />
    );

    expect(screen.getByText(/学习计划已保存/)).toBeInTheDocument();
  });

  it('should handle empty tasks array', () => {
    render(
      <StreamingPlanViewer
        phase="streaming"
        tasks={[]}
        taskCount={0}
        onConfirmSave={mockConfirmSave}
        onRegenerate={mockRegenerate}
      />
    );

    expect(screen.getByText('AI 正在生成任务...')).toBeInTheDocument();
  });
});
