import { describe, it, expect, vi, beforeEach, type Mock } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { LabelManager } from '@/components/task/label-manager';

// Mock API 模块
vi.mock('@/lib/api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    delete: vi.fn(),
  },
}));

import api from '@/lib/api';

// 创建测试用 QueryClient
function createTestQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        gcTime: 0,
      },
    },
  });
}

// 包装组件提供 QueryClient 上下文
function renderWithQueryClient(ui: React.ReactElement) {
  const queryClient = createTestQueryClient();
  return render(
    <QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>
  );
}

// 模拟标签数据
const mockLabels = [
  { id: 'label-1', name: '前端', color: '#3b82f6' },
  { id: 'label-2', name: '后端', color: '#22c55e' },
  { id: 'label-3', name: '设计', color: '#ef4444' },
];

describe('LabelManager 标签管理组件', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // 默认 mock API 返回标签列表
    (api.get as Mock).mockResolvedValue({
      data: { code: 0, data: mockLabels, message: 'success' },
    });
  });

  it('应渲染标签管理标题', async () => {
    // 验证组件正确渲染标题
    renderWithQueryClient(<LabelManager />);

    expect(screen.getByText('标签管理')).toBeInTheDocument();
  });

  it('应渲染标签列表', async () => {
    // 验证标签列表正确渲染
    renderWithQueryClient(<LabelManager />);

    await waitFor(() => {
      expect(screen.getByText('前端')).toBeInTheDocument();
      expect(screen.getByText('后端')).toBeInTheDocument();
      expect(screen.getByText('设计')).toBeInTheDocument();
    });
  });

  it('标签应显示对应的色点', async () => {
    // 验证每个标签旁边有对应颜色的圆点
    renderWithQueryClient(<LabelManager />);

    await waitFor(() => {
      const colorDots = document.querySelectorAll('span[style*="background-color"]');
      expect(colorDots.length).toBeGreaterThanOrEqual(mockLabels.length);
    });
  });

  it('点击新建标签按钮应展开创建表单', async () => {
    // 验证创建表单的显示/隐藏
    const user = userEvent.setup();

    renderWithQueryClient(<LabelManager />);

    // 点击新建标签按钮
    const createButton = screen.getByText('新建标签');
    await user.click(createButton);

    // 应出现标签名称输入框
    expect(screen.getByPlaceholderText('标签名称')).toBeInTheDocument();
    // 应出现创建按钮
    expect(screen.getByText('创建')).toBeInTheDocument();
  });

  it('创建新标签应调用 API', async () => {
    // 验证创建标签的完整流程
    const user = userEvent.setup();

    (api.post as Mock).mockResolvedValue({
      data: {
        code: 0,
        data: { id: 'label-new', name: '测试标签', color: '#3b82f6' },
        message: 'created',
      },
    });

    renderWithQueryClient(<LabelManager />);

    // 展开创建表单
    await user.click(screen.getByText('新建标签'));

    // 输入标签名称
    const nameInput = screen.getByPlaceholderText('标签名称');
    await user.type(nameInput, '测试标签');

    // 点击创建
    const submitButton = screen.getByText('创建');
    await user.click(submitButton);

    // 验证 API 被调用
    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith('/labels', {
        name: '测试标签',
        color: expect.any(String),
      });
    });
  });

  it('标签名为空时不应调用 API', async () => {
    // 验证空标签名不提交
    const user = userEvent.setup();

    renderWithQueryClient(<LabelManager />);

    // 展开创建表单
    await user.click(screen.getByText('新建标签'));

    // 不输入名称直接点击创建
    const submitButton = screen.getByText('创建');
    await user.click(submitButton);

    // API 不应被调用
    expect(api.post).not.toHaveBeenCalled();
  });

  it('点击取消应关闭创建表单', async () => {
    // 验证取消功能
    const user = userEvent.setup();

    renderWithQueryClient(<LabelManager />);

    // 展开创建表单
    await user.click(screen.getByText('新建标签'));
    expect(screen.getByPlaceholderText('标签名称')).toBeInTheDocument();

    // 点击取消
    await user.click(screen.getByText('取消'));
    expect(screen.queryByPlaceholderText('标签名称')).not.toBeInTheDocument();
  });

  it('应渲染颜色选择预设', async () => {
    // 验证颜色选择预设正确渲染
    const user = userEvent.setup();

    renderWithQueryClient(<LabelManager />);

    // 展开创建表单
    await user.click(screen.getByText('新建标签'));

    // 检查预设颜色按钮（12 个预设颜色）
    const colorButtons = document.querySelectorAll('button[style*="background-color"]');
    expect(colorButtons.length).toBe(12);
  });

  it('应渲染自定义颜色选择器', async () => {
    // 验证自定义颜色输入
    const user = userEvent.setup();

    renderWithQueryClient(<LabelManager />);

    // 展开创建表单
    await user.click(screen.getByText('新建标签'));

    // 检查颜色输入
    const colorInput = document.querySelector('input[type="color"]');
    expect(colorInput).toBeInTheDocument();
  });

  it('点击删除按钮应调用删除 API', async () => {
    // 验证删除标签功能
    const user = userEvent.setup();

    (api.delete as Mock).mockResolvedValue({ data: { code: 0, message: 'success' } });

    renderWithQueryClient(<LabelManager />);

    await waitFor(() => {
      expect(screen.getByText('前端')).toBeInTheDocument();
    });

    // 找到第一个标签的删除按钮
    const labelItems = document.querySelectorAll('[title="删除标签"]');
    expect(labelItems.length).toBeGreaterThanOrEqual(1);

    // 点击第一个删除按钮
    await user.click(labelItems[0]);

    // 验证 API 被调用
    await waitFor(() => {
      expect(api.delete).toHaveBeenCalledWith('/labels/label-1');
    });
  });

  it('标签列表为空时应显示空状态提示', async () => {
    // 验证空列表提示
    (api.get as Mock).mockResolvedValue({
      data: { code: 0, data: [], message: 'success' },
    });

    renderWithQueryClient(<LabelManager />);

    await waitFor(() => {
      expect(screen.getByText('暂无标签')).toBeInTheDocument();
    });
  });

  it('加载中应显示加载动画', () => {
    // 验证加载状态
    // 让 API 一直 pending
    (api.get as Mock).mockReturnValue(new Promise(() => {}));

    renderWithQueryClient(<LabelManager />);

    // 检查加载动画元素
    const spinner = document.querySelector('.animate-spin');
    expect(spinner).toBeInTheDocument();
  });
});
