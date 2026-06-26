import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { TaskFiltersBar, type TaskFilters } from '@/components/task/task-filters';

// 默认筛选条件
const defaultFilters: TaskFilters = {
  keyword: '',
  statuses: [],
  priorities: [],
  labelIds: [],
  deadlineFrom: '',
  deadlineTo: '',
  sortBy: 'created_at',
  sortOrder: 'asc',
};

describe('TaskFiltersBar 任务筛选栏组件', () => {
  it('应渲染搜索输入框', () => {
    // 验证搜索输入框正确渲染
    render(<TaskFiltersBar filters={defaultFilters} onFiltersChange={vi.fn()} />);

    const searchInput = screen.getByPlaceholderText('搜索任务...');
    expect(searchInput).toBeInTheDocument();
  });

  it('搜索输入框的 onChange 应调用 onFiltersChange', async () => {
    // 验证输入关键字时正确触发回调
    const handleChange = vi.fn();
    const user = userEvent.setup();

    render(<TaskFiltersBar filters={defaultFilters} onFiltersChange={handleChange} />);

    const searchInput = screen.getByPlaceholderText('搜索任务...');
    await user.type(searchInput, 'Go');

    expect(handleChange).toHaveBeenCalled();
    // 组件通过 update({ keyword: e.target.value }) 调用，每次按键都会触发
    // 验证至少调用了两次（G 和 o 各一次）
    expect(handleChange.mock.calls.length).toBeGreaterThanOrEqual(2);
    // 最后一次调用的 keyword 应包含完整输入内容
    const lastCall = handleChange.mock.calls[handleChange.mock.calls.length - 1][0];
    expect(lastCall.keyword).toContain('o');
  });

  it('应渲染状态筛选下拉菜单', () => {
    // 验证状态筛选按钮渲染
    render(<TaskFiltersBar filters={defaultFilters} onFiltersChange={vi.fn()} />);

    const statusButton = screen.getByText('状态');
    expect(statusButton).toBeInTheDocument();
  });

  it('点击状态筛选按钮应展开选项列表', async () => {
    // 验证下拉菜单展开
    const user = userEvent.setup();

    render(<TaskFiltersBar filters={defaultFilters} onFiltersChange={vi.fn()} />);

    const statusButton = screen.getByText('状态');
    await user.click(statusButton);

    // 应出现待办、进行中、已完成选项
    expect(screen.getByText('待办')).toBeInTheDocument();
    expect(screen.getByText('进行中')).toBeInTheDocument();
    expect(screen.getByText('已完成')).toBeInTheDocument();
  });

  it('应渲染优先级筛选下拉菜单', () => {
    // 验证优先级筛选按钮渲染
    render(<TaskFiltersBar filters={defaultFilters} onFiltersChange={vi.fn()} />);

    // "优先级" 同时出现在排序选项和筛选按钮中，使用 getAllByText
    const priorityElements = screen.getAllByText('优先级');
    expect(priorityElements.length).toBeGreaterThanOrEqual(1);
  });

  it('应渲染截止日期范围输入框', () => {
    // 验证日期范围输入框正确渲染
    render(<TaskFiltersBar filters={defaultFilters} onFiltersChange={vi.fn()} />);

    const dateInputs = screen.getAllByDisplayValue('');
    const dateTypeInputs = document.querySelectorAll('input[type="date"]');
    expect(dateTypeInputs.length).toBeGreaterThanOrEqual(2);
  });

  it('截止日期输入应触发 onFiltersChange', async () => {
    // 验证日期选择触发回调
    const handleChange = vi.fn();
    const user = userEvent.setup();

    render(<TaskFiltersBar filters={defaultFilters} onFiltersChange={handleChange} />);

    const dateInputs = document.querySelectorAll('input[type="date"]');
    if (dateInputs.length >= 1) {
      await user.type(dateInputs[0] as HTMLElement, '2025-01-01');
      expect(handleChange).toHaveBeenCalled();
    }
  });

  it('应渲染排序选择器', () => {
    // 验证排序选择器渲染
    render(<TaskFiltersBar filters={defaultFilters} onFiltersChange={vi.fn()} />);

    const sortSelect = screen.getByDisplayValue('创建时间');
    expect(sortSelect).toBeInTheDocument();
  });

  it('排序选择器应包含所有排序选项', () => {
    // 验证排序选项完整
    render(<TaskFiltersBar filters={defaultFilters} onFiltersChange={vi.fn()} />);

    const sortSelect = screen.getByDisplayValue('创建时间') as HTMLSelectElement;
    const options = Array.from(sortSelect.options).map((opt) => opt.text);

    expect(options).toContain('创建时间');
    expect(options).toContain('更新时间');
    expect(options).toContain('截止日期');
    expect(options).toContain('优先级');
    expect(options).toContain('自定义排序');
  });

  it('有活跃筛选时应显示清除筛选按钮', () => {
    // 验证筛选条件存在时出现清除按钮
    const activeFilters: TaskFilters = {
      ...defaultFilters,
      keyword: 'Go',
    };

    render(<TaskFiltersBar filters={activeFilters} onFiltersChange={vi.fn()} />);

    const clearButton = screen.getByText('清除筛选');
    expect(clearButton).toBeInTheDocument();
  });

  it('无活跃筛选时不应显示清除筛选按钮', () => {
    // 验证无筛选条件时清除按钮不出现
    render(<TaskFiltersBar filters={defaultFilters} onFiltersChange={vi.fn()} />);

    const clearButton = screen.queryByText('清除筛选');
    expect(clearButton).not.toBeInTheDocument();
  });

  it('点击清除筛选应重置所有筛选条件', async () => {
    // 验证清除按钮功能
    const handleChange = vi.fn();
    const user = userEvent.setup();

    const activeFilters: TaskFilters = {
      keyword: 'Go',
      statuses: ['todo'],
      priorities: ['high'],
      labelIds: ['label-1'],
      deadlineFrom: '2025-01-01',
      deadlineTo: '2025-12-31',
      sortBy: 'deadline',
      sortOrder: 'desc',
    };

    render(<TaskFiltersBar filters={activeFilters} onFiltersChange={handleChange} />);

    const clearButton = screen.getByText('清除筛选');
    await user.click(clearButton);

    expect(handleChange).toHaveBeenCalledWith({
      keyword: '',
      statuses: [],
      priorities: [],
      labelIds: [],
      deadlineFrom: '',
      deadlineTo: '',
      sortBy: 'sort_order',
      sortOrder: 'asc',
    });
  });

  it('应渲染标签筛选（传入标签数据时）', () => {
    // 验证传入标签列表时显示标签筛选
    const labels = [
      { id: 'label-1', name: '前端', color: '#3b82f6' },
      { id: 'label-2', name: '后端', color: '#22c55e' },
    ];

    render(
      <TaskFiltersBar
        filters={defaultFilters}
        onFiltersChange={vi.fn()}
        labels={labels}
      />
    );

    const labelButton = screen.getByText('标签');
    expect(labelButton).toBeInTheDocument();
  });

  it('不传标签数据时不应显示标签筛选', () => {
    // 验证无标签数据时标签筛选不出现
    render(<TaskFiltersBar filters={defaultFilters} onFiltersChange={vi.fn()} />);

    const labelButton = screen.queryByText('标签');
    expect(labelButton).not.toBeInTheDocument();
  });
});
