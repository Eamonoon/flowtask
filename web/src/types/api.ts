export interface ApiResponse<T = unknown> {
  code: number;
  data: T;
  message: string;
}

export interface User {
  id: string;
  email: string;
  display_name: string;
  avatar_url: string | null;
  preferences: UserPreferences;
  created_at: string;
}

export interface UserPreferences {
  theme?: 'light' | 'dark';
  language?: string;
  learning_style?: string;
  weekly_study_hours?: number;
  preferred_session_minutes?: number;
}

export interface AuthData {
  user: User;
  access_token: string;
  refresh_token: string;
}

export interface LearningGoal {
  id: string;
  description: string;
  target_duration: string | null;
  status: 'active' | 'paused' | 'completed' | 'archived';
  task_count: number;
  completed_task_count: number;
  created_at: string;
  updated_at: string;
}

export interface Task {
  id: string;
  title: string;
  description: string | null;
  status: 'todo' | 'doing' | 'done';
  priority: 'low' | 'medium' | 'high' | 'urgent';
  deadline: string | null;
  estimated_duration: string | null;
  recommended_resources: Resource[];
  labels: Label[];
  subtask_count: number;
  completed_subtask_count: number;
  learning_goal_id: string | null;
  parent_task_id: string | null;
  dependencies: string[];
  sort_order: number;
  created_at: string;
  updated_at: string;
}

export interface Resource {
  title: string;
  url?: string;
  description?: string;
}

export interface Label {
  id: string;
  name: string;
  color: string;
}

export interface PaginatedResponse<T> {
  items: T[];
  total?: number;
  page?: number;
  page_size?: number;
  next_cursor?: string;
  has_more: boolean;
}

export interface StudySession {
  id: string;
  task_id: string | null;
  duration: number;
  date: string;
  notes: string | null;
  created_at: string;
}

export interface AIConversation {
  id: string;
  title: string | null;
  learning_goal_id: string | null;
  created_at: string;
  updated_at: string;
}

export interface AIMessage {
  id: string;
  conversation_id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  created_at: string;
}

export interface DashboardStats {
  today_tasks: {
    total: number;
    completed: number;
    items: Task[];
  };
  overall: {
    total_tasks: number;
    completed_tasks: number;
    completion_rate: number;
  };
  study_time: {
    today_minutes: number;
    week_minutes: number;
    month_minutes: number;
  };
  upcoming_deadlines: Task[];
  recent_activity: ActivityItem[];
}

export interface ActivityItem {
  type: string;
  description: string;
  timestamp: string;
}

export interface ChartData {
  labels: string[];
  values: number[];
}
