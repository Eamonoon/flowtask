import { create } from 'zustand';
import type { Task } from '@/types/api';

export type GenerationPhase = 'idle' | 'connecting' | 'streaming' | 'preview' | 'done' | 'error';

interface GoalState {
  // Generation state
  generationPhase: GenerationPhase;
  generatedTasks: Task[];
  errorMessage: string;
  taskCount: number;
  sessionId: string | null;
  learningGoalId: string | null;

  // Actions
  setPhase: (phase: GenerationPhase) => void;
  addTask: (task: Task) => void;
  setTasks: (tasks: Task[]) => void;
  setError: (error: string) => void;
  setSessionId: (id: string | null) => void;
  setLearningGoalId: (id: string | null) => void;
  reset: () => void;
  clearError: () => void;
}

const initialState = {
  generationPhase: 'idle' as GenerationPhase,
  generatedTasks: [],
  errorMessage: '',
  taskCount: 0,
  sessionId: null,
  learningGoalId: null,
};

export const useGoalStore = create<GoalState>((set) => ({
  ...initialState,

  setPhase: (phase) => set({ generationPhase: phase }),

  addTask: (task) =>
    set((state) => ({
      generatedTasks: [...state.generatedTasks, task],
      taskCount: state.taskCount + 1,
    })),

  setTasks: (tasks) =>
    set({
      generatedTasks: tasks,
      taskCount: tasks.length,
    }),

  setError: (error) =>
    set({
      errorMessage: error,
      generationPhase: 'error',
    }),

  setSessionId: (id) => set({ sessionId: id }),

  setLearningGoalId: (id) => set({ learningGoalId: id }),

  reset: () => set(initialState),

  clearError: () => set({ errorMessage: '' }),
}));
