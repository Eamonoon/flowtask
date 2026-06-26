import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { saveProgress, loadProgress, clearProgress, isProgressExpired } from '@/lib/goal-progress';

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => {
      store[key] = value;
    },
    removeItem: (key: string) => {
      delete store[key];
    },
    clear: () => {
      store = {};
    },
    get length() {
      return Object.keys(store).length;
    },
    key: (index: number) => Object.keys(store)[index] || null,
  };
})();

Object.defineProperty(window, 'localStorage', { value: localStorageMock });

describe('Session Resumption Flow', () => {
  beforeEach(() => {
    localStorageMock.clear();
  });

  it('should save progress during generation', () => {
    const tasks = [
      { id: '1', title: 'Learn Go Basics' },
      { id: '2', title: 'Learn Go Concurrency' },
    ];

    saveProgress('goal-123', tasks, 'session-456', 'goal-123');

    const saved = localStorageMock.getItem('goal-progress-goal-123');
    expect(saved).not.toBeNull();

    const parsed = JSON.parse(saved!);
    expect(parsed.tasks).toHaveLength(2);
    expect(parsed.sessionId).toBe('session-456');
    expect(parsed.learningGoalId).toBe('goal-123');
  });

  it('should restore progress on page load', () => {
    const tasks = [
      { id: '1', title: 'Learn Go Basics' },
      { id: '2', title: 'Learn Go Concurrency' },
    ];

    saveProgress('goal-123', tasks, 'session-456', 'goal-123');

    const restored = loadProgress('goal-123');
    expect(restored).not.toBeNull();
    expect(restored!.tasks).toHaveLength(2);
    expect(restored!.sessionId).toBe('session-456');
  });

  it('should clear progress after successful save', () => {
    saveProgress('goal-123', [{ id: '1', title: 'Task 1' }], 'session-456', 'goal-123');

    clearProgress('goal-123');

    const restored = loadProgress('goal-123');
    expect(restored).toBeNull();
  });

  it('should detect expired progress', () => {
    saveProgress('goal-123', [{ id: '1', title: 'Task 1' }], 'session-456', 'goal-123');

    // Mock old timestamp (25 hours ago)
    const saved = JSON.parse(localStorageMock.getItem('goal-progress-goal-123')!);
    saved.timestamp = Date.now() - 25 * 60 * 60 * 1000;
    localStorageMock.setItem('goal-progress-goal-123', JSON.stringify(saved));

    expect(isProgressExpired('goal-123')).toBe(true);
  });

  it('should not load expired progress', () => {
    saveProgress('goal-123', [{ id: '1', title: 'Task 1' }], 'session-456', 'goal-123');

    // Mock old timestamp
    const saved = JSON.parse(localStorageMock.getItem('goal-progress-goal-123')!);
    saved.timestamp = Date.now() - 25 * 60 * 60 * 1000;
    localStorageMock.setItem('goal-progress-goal-123', JSON.stringify(saved));

    const restored = loadProgress('goal-123');
    expect(restored).toBeNull();
  });

  it('should handle multiple goals independently', () => {
    saveProgress('goal-1', [{ id: '1', title: 'Task 1' }], 'session-1', 'goal-1');
    saveProgress('goal-2', [{ id: '2', title: 'Task 2' }, { id: '3', title: 'Task 3' }], 'session-2', 'goal-2');

    const restored1 = loadProgress('goal-1');
    const restored2 = loadProgress('goal-2');

    expect(restored1!.tasks).toHaveLength(1);
    expect(restored2!.tasks).toHaveLength(2);

    // Clear one shouldn't affect the other
    clearProgress('goal-1');
    expect(loadProgress('goal-1')).toBeNull();
    expect(loadProgress('goal-2')).not.toBeNull();
  });
});
