'use client';

import { useState, useEffect, useRef } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useQueryClient } from '@tanstack/react-query';
import { useAuthStore } from '@/stores/auth-store';
import { ConversationList } from '@/components/chat/conversation-list';
import { ChatMessage } from '@/components/chat/chat-message';
import { ChatInput } from '@/components/chat/chat-input';
import { useChatStream } from '@/hooks/use-chat-stream';
import api from '@/lib/api';
import type { ApiResponse, AIMessage } from '@/types/api';
import { Button } from '@/components/ui/button';
import { Bot, Sparkles, Loader2, Save } from 'lucide-react';

type ChatMode = 'chat' | 'task-breakdown';

export default function ChatPage() {
  const { isAuthenticated } = useAuthStore();
  const router = useRouter();
  const queryClient = useQueryClient();

  const [conversationId, setConversationId] = useState<string | null>(null);
  const [messages, setMessages] = useState<AIMessage[]>([]);
  const [mode, setMode] = useState<ChatMode>('chat');
  const [learningGoalId, setLearningGoalId] = useState<string | null>(null);
  const [saveStatus, setSaveStatus] = useState<
    'idle' | 'saving' | 'saved' | 'error'
  >('idle');

  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!isAuthenticated) router.push('/login');
  }, [isAuthenticated, router]);

  // Scroll to bottom on new messages
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, conversationId]);

  const {
    sendMessage,
    isStreaming,
    currentResponse,
    taskTree,
    saveAsTasks,
  } = useChatStream({
    conversationId,
    learningGoalId,
    mode,
    onConversationCreated: (id) => {
      setConversationId(id);
      queryClient.invalidateQueries({ queryKey: ['ai', 'conversations'] });
    },
    onDone: (fullContent) => {
      // Append AI message to the list
      setMessages((prev) => [
        ...prev,
        {
          id: `stream-${Date.now()}`,
          conversation_id: conversationId || '',
          role: 'assistant',
          content: fullContent,
          created_at: new Date().toISOString(),
        },
      ]);
    },
  });

  // Load conversation messages when selecting a conversation
  const handleSelectConversation = async (id: string) => {
    setConversationId(id);
    setMessages([]);
    setSaveStatus('idle');
    try {
      const { data } = await api.get<ApiResponse<AIMessage[]>>(
        `/ai/conversations/${id}/messages`
      );
      setMessages(data.data);
    } catch {
      // Ignore load errors
    }
  };

  const handleNewConversation = () => {
    setConversationId(null);
    setMessages([]);
    setSaveStatus('idle');
  };

  const handleSend = (content: string) => {
    // Add user message immediately
    const userMsg: AIMessage = {
      id: `user-${Date.now()}`,
      conversation_id: conversationId || '',
      role: 'user',
      content,
      created_at: new Date().toISOString(),
    };
    setMessages((prev) => [...prev, userMsg]);
    sendMessage(content);
  };

  const handleSaveAsTasks = async () => {
    if (!taskTree || !learningGoalId) return;
    setSaveStatus('saving');
    try {
      await saveAsTasks(learningGoalId);
      setSaveStatus('saved');
    } catch {
      setSaveStatus('error');
    }
  };

  if (!isAuthenticated) return null;

  return (
    <div className="min-h-screen bg-background flex">
      {/* Left sidebar: conversation list */}
      <ConversationList
        activeConversationId={conversationId}
        onSelect={handleSelectConversation}
        onNew={handleNewConversation}
        onDelete={(id) => {
          if (id === conversationId) {
            setConversationId(null);
            setMessages([]);
          }
        }}
      />

      {/* Right panel: chat area */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* Top nav */}
        <nav className="border-b px-6 py-4">
          <div className="flex justify-between items-center">
            <h1 className="text-xl font-bold">FlowTask</h1>
            <div className="flex gap-4 items-center">
              <Link href="/dashboard">Dashboard</Link>
              <Link href="/tasks">任务</Link>
              <Link href="/goals">学习目标</Link>
              <Link href="/chat" className="font-semibold text-primary">
                AI 助手
              </Link>
            </div>
          </div>
        </nav>

        {/* Mode switcher + goal selector */}
        <div className="border-b px-6 py-2 flex items-center gap-3">
          <div className="flex items-center gap-1 bg-muted rounded-lg p-0.5">
            <button
              onClick={() => setMode('chat')}
              className={`px-3 py-1 text-sm rounded-md transition-colors ${
                mode === 'chat'
                  ? 'bg-background text-foreground shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              普通对话
            </button>
            <button
              onClick={() => setMode('task-breakdown')}
              className={`px-3 py-1 text-sm rounded-md transition-colors flex items-center gap-1 ${
                mode === 'task-breakdown'
                  ? 'bg-background text-foreground shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              <Sparkles className="size-3" />
              任务拆解
            </button>
          </div>

          {mode === 'task-breakdown' && (
            <input
              type="text"
              value={learningGoalId || ''}
              onChange={(e) => setLearningGoalId(e.target.value || null)}
              placeholder="学习目标 ID（可选）"
              className="text-sm border rounded px-2 py-1 bg-background w-60"
            />
          )}

          {taskTree && saveStatus === 'idle' && learningGoalId && (
            <Button size="sm" onClick={handleSaveAsTasks}>
              <Save className="size-3.5 mr-1" />
              保存为任务
            </Button>
          )}
          {saveStatus === 'saving' && (
            <span className="text-sm text-muted-foreground flex items-center gap-1">
              <Loader2 className="size-3.5 animate-spin" />
              保存中...
            </span>
          )}
          {saveStatus === 'saved' && (
            <span className="text-sm text-green-600">已保存</span>
          )}
          {saveStatus === 'error' && (
            <span className="text-sm text-destructive">保存失败</span>
          )}
        </div>

        {/* Messages */}
        <div className="flex-1 overflow-y-auto px-6 py-4">
          <div className="max-w-3xl mx-auto space-y-4">
            {messages.length === 0 && !isStreaming && (
              <div className="flex items-center justify-center h-full text-muted-foreground py-32">
                <div className="text-center">
                  <div className="size-14 rounded-full bg-primary/10 flex items-center justify-center mx-auto mb-4">
                    <Bot className="size-7 text-primary" />
                  </div>
                  <h2 className="text-xl font-semibold mb-2 text-foreground">
                    AI 学习助手
                  </h2>
                  <p>
                    {mode === 'chat'
                      ? '问我任何学习相关的问题'
                      : '描述你想学习的内容，AI 会帮你拆解为可执行的任务'}
                  </p>
                </div>
              </div>
            )}

            {messages.map((msg) => (
              <ChatMessage key={msg.id} message={msg} />
            ))}

            {/* Streaming response in progress */}
            {isStreaming && currentResponse && (
              <ChatMessage
                message={{
                  id: 'streaming',
                  conversation_id: '',
                  role: 'assistant',
                  content: currentResponse,
                  created_at: new Date().toISOString(),
                }}
              />
            )}

            {isStreaming && !currentResponse && (
              <div className="flex gap-2 mr-auto">
                <div className="size-8 rounded-full bg-primary/10 flex items-center justify-center">
                  <Bot className="size-4 text-primary" />
                </div>
                <div className="bg-muted rounded-2xl rounded-bl-md px-4 py-2.5">
                  <div className="flex items-center gap-1.5 text-sm text-muted-foreground">
                    <Loader2 className="size-3.5 animate-spin" />
                    思考中...
                  </div>
                </div>
              </div>
            )}

            <div ref={messagesEndRef} />
          </div>
        </div>

        {/* Input */}
        <ChatInput onSend={handleSend} disabled={isStreaming} />
      </div>
    </div>
  );
}
