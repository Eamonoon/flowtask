import { describe, it, expect, vi, beforeEach, type Mock } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

// Mock next/navigation
vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: vi.fn(),
    replace: vi.fn(),
  }),
}));

// Mock API
vi.mock('@/lib/api', () => ({
  default: {
    put: vi.fn(),
  },
}));

// Mock auth store
vi.mock('@/stores/auth-store', () => ({
  useAuthStore: vi.fn(),
}));

import api from '@/lib/api';
import { useAuthStore } from '@/stores/auth-store';

// 模拟用户数据
const mockUser = {
  id: 'user-1',
  email: 'test@example.com',
  display_name: '测试用户',
  avatar_url: 'https://example.com/avatar.png',
  preferences: {
    theme: 'light' as const,
    language: 'zh-CN',
    learning_style: 'visual',
    weekly_study_hours: 10,
    preferred_session_minutes: 30,
  },
  created_at: '2025-01-01T00:00:00Z',
};

// 导入 ProfilePage 组件
import ProfilePage from '@/app/(dashboard)/profile/page';

describe('ProfilePage 个人资料表单组件', () => {
  beforeEach(() => {
    vi.clearAllMocks();

    // 默认 mock auth store
    (useAuthStore as unknown as Mock).mockReturnValue({
      user: mockUser,
      isAuthenticated: true,
      logout: vi.fn(),
      setUser: vi.fn(),
    });
  });

  it('应正确渲染用户数据', () => {
    // 验证表单中显示用户当前数据
    render(<ProfilePage />);

    // 邮箱应以只读文本显示
    expect(screen.getByText('test@example.com')).toBeInTheDocument();

    // 显示名称输入框应有默认值
    const displayNameInput = screen.getByDisplayValue('测试用户');
    expect(displayNameInput).toBeInTheDocument();
  });

  it('显示名称最少应为2个字符', async () => {
    // 验证 display_name 验证规则
    const user = userEvent.setup();

    render(<ProfilePage />);

    // 清空显示名称
    const displayNameInput = screen.getByDisplayValue('测试用户');
    await user.clear(displayNameInput);
    await user.type(displayNameInput, 'A');

    // 提交表单
    const submitButton = screen.getByText('保存');
    await user.click(submitButton);

    // 应显示验证错误
    await waitFor(() => {
      expect(screen.getByText('显示名称至少 2 个字符')).toBeInTheDocument();
    });
  });

  it('显示名称应接受有效输入', async () => {
    // 验证有效显示名称不触发错误
    const user = userEvent.setup();

    (api.put as Mock).mockResolvedValue({
      data: { code: 0, data: mockUser, message: 'success' },
    });

    render(<ProfilePage />);

    // 修改显示名称
    const displayNameInput = screen.getByDisplayValue('测试用户');
    await user.clear(displayNameInput);
    await user.type(displayNameInput, '新名称');

    // 提交表单
    const submitButton = screen.getByText('保存');
    await user.click(submitButton);

    // 不应显示验证错误
    await waitFor(() => {
      expect(screen.queryByText('显示名称至少 2 个字符')).not.toBeInTheDocument();
    });
  });

  it('应正确渲染偏好设置字段', () => {
    // 验证偏好设置字段渲染
    render(<ProfilePage />);

    // 偏好设置标题
    expect(screen.getByText('偏好设置')).toBeInTheDocument();

    // 主题选择
    expect(screen.getByLabelText('主题')).toBeInTheDocument();

    // 语言选择
    expect(screen.getByLabelText('语言')).toBeInTheDocument();

    // 学习风格选择
    expect(screen.getByLabelText('学习风格')).toBeInTheDocument();

    // 每周学习目标
    expect(screen.getByLabelText('每周学习目标（小时）')).toBeInTheDocument();

    // 默认学习时长
    expect(screen.getByLabelText('默认学习时长（分钟）')).toBeInTheDocument();
  });

  it('主题选择应包含浅色和深色选项', () => {
    // 验证主题选项
    render(<ProfilePage />);

    const themeSelect = screen.getByLabelText('主题') as HTMLSelectElement;
    const options = Array.from(themeSelect.options).map((opt) => opt.text);

    expect(options).toContain('浅色');
    expect(options).toContain('深色');
  });

  it('语言选择应包含中文、英文和日文选项', () => {
    // 验证语言选项
    render(<ProfilePage />);

    const langSelect = screen.getByLabelText('语言') as HTMLSelectElement;
    const options = Array.from(langSelect.options).map((opt) => opt.text);

    expect(options).toContain('简体中文');
    expect(options).toContain('English');
    expect(options).toContain('日本語');
  });

  it('学习风格选择应包含多种风格', () => {
    // 验证学习风格选项
    render(<ProfilePage />);

    const styleSelect = screen.getByLabelText('学习风格') as HTMLSelectElement;
    const options = Array.from(styleSelect.options).map((opt) => opt.text);

    expect(options).toContain('视觉型');
    expect(options).toContain('听觉型');
    expect(options).toContain('阅读型');
    expect(options).toContain('动手型');
  });

  it('表单提交应调用 API', async () => {
    // 验证表单提交正确调用 API
    const user = userEvent.setup();
    const setUserMock = vi.fn();

    (useAuthStore as unknown as Mock).mockReturnValue({
      user: mockUser,
      isAuthenticated: true,
      logout: vi.fn(),
      setUser: setUserMock,
    });

    (api.put as Mock).mockResolvedValue({
      data: {
        code: 0,
        data: { ...mockUser, display_name: '更新后的名称' },
        message: 'success',
      },
    });

    render(<ProfilePage />);

    // 修改显示名称
    const displayNameInput = screen.getByDisplayValue('测试用户');
    await user.clear(displayNameInput);
    await user.type(displayNameInput, '更新后的名称');

    // 提交表单
    const submitButton = screen.getByText('保存');
    await user.click(submitButton);

    // 验证 API 调用
    await waitFor(() => {
      expect(api.put).toHaveBeenCalledWith('/user/profile', expect.objectContaining({
        display_name: '更新后的名称',
      }));
    });
  });

  it('应渲染退出登录按钮', () => {
    // 验证退出登录按钮存在
    render(<ProfilePage />);

    const logoutButton = screen.getByText('退出登录');
    expect(logoutButton).toBeInTheDocument();
  });

  it('未认证用户应跳转到登录页', () => {
    // 验证未认证时的重定向行为
    (useAuthStore as unknown as Mock).mockReturnValue({
      user: null,
      isAuthenticated: false,
      logout: vi.fn(),
      setUser: vi.fn(),
    });

    const { container } = render(<ProfilePage />);

    // 未认证时不应渲染表单内容
    expect(container.innerHTML).toBe('');
  });

  it('应渲染头像 URL 输入框', () => {
    // 验证头像 URL 字段
    render(<ProfilePage />);

    const avatarInput = screen.getByLabelText('头像 URL');
    expect(avatarInput).toBeInTheDocument();
  });
});
