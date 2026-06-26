'use client';

const STORAGE_PREFIX = 'goal-progress-';
const EXPIRATION_HOURS = 24;

interface SavedProgress {
  tasks: any[];
  timestamp: number;
  sessionId: string;
  learningGoalId: string;
}

/**
 * Save generation progress to localStorage
 */
export function saveProgress(
  goalId: string,
  tasks: any[],
  sessionId?: string,
  learningGoalId?: string
): void {
  if (typeof window === 'undefined') return;

  const progress: SavedProgress = {
    tasks,
    timestamp: Date.now(),
    sessionId: sessionId || '',
    learningGoalId: learningGoalId || goalId,
  };

  try {
    localStorage.setItem(
      `${STORAGE_PREFIX}${goalId}`,
      JSON.stringify(progress)
    );
  } catch (err) {
    console.error('Failed to save progress to localStorage:', err);
  }
}

/**
 * Load generation progress from localStorage
 */
export function loadProgress(goalId: string): SavedProgress | null {
  if (typeof window === 'undefined') return null;

  try {
    const saved = localStorage.getItem(`${STORAGE_PREFIX}${goalId}`);
    if (!saved) return null;

    const progress: SavedProgress = JSON.parse(saved);

    // Check if expired
    if (isProgressExpired(goalId)) {
      clearProgress(goalId);
      return null;
    }

    return progress;
  } catch (err) {
    console.error('Failed to load progress from localStorage:', err);
    return null;
  }
}

/**
 * Clear generation progress from localStorage
 */
export function clearProgress(goalId: string): void {
  if (typeof window === 'undefined') return;

  try {
    localStorage.removeItem(`${STORAGE_PREFIX}${goalId}`);
  } catch (err) {
    console.error('Failed to clear progress from localStorage:', err);
  }
}

/**
 * Check if saved progress has expired (older than 24 hours)
 */
export function isProgressExpired(goalId: string): boolean {
  if (typeof window === 'undefined') return true;

  try {
    const saved = localStorage.getItem(`${STORAGE_PREFIX}${goalId}`);
    if (!saved) return true;

    const progress: SavedProgress = JSON.parse(saved);
    const expirationMs = EXPIRATION_HOURS * 60 * 60 * 1000;
    const isExpired = Date.now() - progress.timestamp > expirationMs;

    return isExpired;
  } catch {
    return true;
  }
}

/**
 * Get all saved progress keys
 */
export function getAllProgressKeys(): string[] {
  if (typeof window === 'undefined') return [];

  const keys: string[] = [];
  for (let i = 0; i < localStorage.length; i++) {
    const key = localStorage.key(i);
    if (key && key.startsWith(STORAGE_PREFIX)) {
      keys.push(key.replace(STORAGE_PREFIX, ''));
    }
  }
  return keys;
}

/**
 * Clear all expired progress
 */
export function clearExpiredProgress(): void {
  const keys = getAllProgressKeys();
  keys.forEach((goalId) => {
    if (isProgressExpired(goalId)) {
      clearProgress(goalId);
    }
  });
}
