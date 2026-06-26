import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import GoalsPage from '@/app/(dashboard)/goals/page';

// Mock dependencies
vi.mock('@/stores/auth-store', () => ({
  useAuthStore: () => ({
    isAuthenticated: true,
    accessToken: 'test-token',
  }),
}));

vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: vi.fn(),
  }),
}));

describe('Goals Page Integration', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render the goals page', () => {
    render(<GoalsPage />);

    expect(screen.getByText('创建学习目标')).toBeInTheDocument();
    expect(screen.getByText('生成学习计划')).toBeInTheDocument();
  });

  it('should show input fields', () => {
    render(<GoalsPage />);

    expect(screen.getByPlaceholderText('例如：我想两个月学会 RAG')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('目标时长（可选）')).toBeInTheDocument();
  });

  it('should disable generate button when input is empty', () => {
    render(<GoalsPage />);

    const button = screen.getByText('生成学习计划');
    expect(button).toBeDisabled();
  });

  it('should enable generate button when input has text', async () => {
    const user = userEvent.setup();
    render(<GoalsPage />);

    const textarea = screen.getByPlaceholderText('例如：我想两个月学会 RAG');
    await user.type(textarea, '我想学 Go');

    const button = screen.getByText('生成学习计划');
    expect(button).not.toBeDisabled();
  });

  it('should show my learning goals section', () => {
    render(<GoalsPage />);

    expect(screen.getByText('我的学习目标')).toBeInTheDocument();
  });
});
