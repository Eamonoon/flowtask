import { describe, it, expect, beforeEach } from 'vitest';
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
  };
})();

Object.defineProperty(window, 'localStorage', { value: localStorageMock });

describe('Goal Progress localStorage', () => {
  beforeEach(() => {
    localStorageMock.clear();
  });

  it('should save and load progress', () => {
    const tasks = [
      { id: '1', title: 'Task 1', description: 'Desc 1' },
      { id: '2', title: 'Task 2', description: 'Desc 2' },
    ];

    saveProgress('goal-123', tasks);
    const loaded = loadProgress('goal-123');

    expect(loaded).not.toBeNull();
    expect(loaded!.tasks).toHaveLength(2);
    expect(loaded!.tasks[0].title).toBe('Task 1');
  });

  it('should return null for non-existent progress', () => {
    const loaded = loadProgress('non-existent');
    expect(loaded).toBeNull();
  });

  it('should clear progress', () => {
    const tasks = [{ id: '1', title: 'Task 1' }];
    saveProgress('goal-123', tasks);

    clearProgress('goal-123');
    const loaded = loadProgress('goal-123');

    expect(loaded).toBeNull();
  });

  it('should detect expired progress', () => {
    const tasks = [{ id: '1', title: 'Task 1' }];
    saveProgress('goal-123', tasks);

    // Mock old timestamp (25 hours ago)
    const saved = JSON.parse(localStorageMock.getItem('goal-progress-goal-123')!);
    saved.timestamp = Date.now() - 25 * 60 * 60 * 1000;
    localStorageMock.setItem('goal-progress-goal-123', JSON.stringify(saved));

    const expired = isProgressExpired('goal-123');
    expect(expired).toBe(true);
  });

  it('should not detect fresh progress as expired', () => {
    const tasks = [{ id: '1', title: 'Task 1' }];
    saveProgress('goal-123', tasks);

    const expired = isProgressExpired('goal-123');
    expect(expired).toBe(false);
  });
});
