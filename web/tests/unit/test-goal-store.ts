import { describe, it, expect, beforeEach } from 'vitest';
import { useGoalStore } from '@/stores/goal-store';

describe('Goal Store', () => {
  beforeEach(() => {
    // Reset store before each test
    useGoalStore.getState().reset();
  });

  it('should have correct initial state', () => {
    const state = useGoalStore.getState();

    expect(state.generationPhase).toBe('idle');
    expect(state.generatedTasks).toEqual([]);
    expect(state.errorMessage).toBe('');
    expect(state.taskCount).toBe(0);
    expect(state.sessionId).toBeNull();
    expect(state.learningGoalId).toBeNull();
  });

  it('should set phase correctly', () => {
    const { setPhase } = useGoalStore.getState();

    setPhase('connecting');
    expect(useGoalStore.getState().generationPhase).toBe('connecting');

    setPhase('streaming');
    expect(useGoalStore.getState().generationPhase).toBe('streaming');

    setPhase('preview');
    expect(useGoalStore.getState().generationPhase).toBe('preview');

    setPhase('done');
    expect(useGoalStore.getState().generationPhase).toBe('done');
  });

  it('should add tasks and increment count', () => {
    const { addTask } = useGoalStore.getState();

    addTask({ id: '1', title: 'Task 1' });
    expect(useGoalStore.getState().generatedTasks).toHaveLength(1);
    expect(useGoalStore.getState().taskCount).toBe(1);

    addTask({ id: '2', title: 'Task 2' });
    expect(useGoalStore.getState().generatedTasks).toHaveLength(2);
    expect(useGoalStore.getState().taskCount).toBe(2);
  });

  it('should set tasks array', () => {
    const { setTasks } = useGoalStore.getState();
    const tasks = [
      { id: '1', title: 'Task 1' },
      { id: '2', title: 'Task 2' },
      { id: '3', title: 'Task 3' },
    ];

    setTasks(tasks);
    expect(useGoalStore.getState().generatedTasks).toHaveLength(3);
    expect(useGoalStore.getState().taskCount).toBe(3);
  });

  it('should set error and change phase to error', () => {
    const { setError } = useGoalStore.getState();

    setError('Something went wrong');
    expect(useGoalStore.getState().errorMessage).toBe('Something went wrong');
    expect(useGoalStore.getState().generationPhase).toBe('error');
  });

  it('should set session ID', () => {
    const { setSessionId } = useGoalStore.getState();

    setSessionId('session-123');
    expect(useGoalStore.getState().sessionId).toBe('session-123');

    setSessionId(null);
    expect(useGoalStore.getState().sessionId).toBeNull();
  });

  it('should set learning goal ID', () => {
    const { setLearningGoalId } = useGoalStore.getState();

    setLearningGoalId('goal-123');
    expect(useGoalStore.getState().learningGoalId).toBe('goal-123');

    setLearningGoalId(null);
    expect(useGoalStore.getState().learningGoalId).toBeNull();
  });

  it('should reset state completely', () => {
    const { setPhase, addTask, setError, setSessionId, setLearningGoalId, reset } =
      useGoalStore.getState();

    // Modify state
    setPhase('streaming');
    addTask({ id: '1', title: 'Task 1' });
    setError('error');
    setSessionId('session-123');
    setLearningGoalId('goal-123');

    // Reset
    reset();

    const state = useGoalStore.getState();
    expect(state.generationPhase).toBe('idle');
    expect(state.generatedTasks).toEqual([]);
    expect(state.errorMessage).toBe('');
    expect(state.taskCount).toBe(0);
    expect(state.sessionId).toBeNull();
    expect(state.learningGoalId).toBeNull();
  });

  it('should clear error', () => {
    const { setError, clearError } = useGoalStore.getState();

    setError('error message');
    expect(useGoalStore.getState().errorMessage).toBe('error message');

    clearError();
    expect(useGoalStore.getState().errorMessage).toBe('');
  });
});
