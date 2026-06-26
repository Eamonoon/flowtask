'use client';

import { cn } from '@/lib/utils';
import { Bot } from 'lucide-react';
import type { AIMessage } from '@/types/api';

interface ChatMessageProps {
  message: AIMessage;
}

export function ChatMessage({ message }: ChatMessageProps) {
  const isUser = message.role === 'user';

  return (
    <div
      className={cn(
        'flex gap-2 max-w-[80%]',
        isUser ? 'ml-auto flex-row-reverse' : 'mr-auto'
      )}
    >
      {/* AI avatar */}
      {!isUser && (
        <div className="size-8 rounded-full bg-primary/10 flex items-center justify-center shrink-0">
          <Bot className="size-4 text-primary" />
        </div>
      )}

      {/* Bubble */}
      <div
        className={cn(
          'px-4 py-2.5 rounded-2xl text-sm leading-relaxed break-words',
          isUser
            ? 'bg-primary text-primary-foreground rounded-br-md'
            : 'bg-muted rounded-bl-md'
        )}
      >
        {isUser ? (
          <span className="whitespace-pre-wrap">{message.content}</span>
        ) : (
          <div className="prose-chat">{renderMarkdown(message.content)}</div>
        )}
      </div>
    </div>
  );
}

/**
 * Simple markdown renderer for AI messages.
 * Handles: bold, inline code, code blocks, unordered lists, headings.
 */
function renderMarkdown(text: string): React.ReactNode[] {
  const lines = text.split('\n');
  const elements: React.ReactNode[] = [];
  let i = 0;

  while (i < lines.length) {
    const line = lines[i];

    // Code block
    if (line.startsWith('```')) {
      const lang = line.slice(3).trim();
      const codeLines: string[] = [];
      i++;
      while (i < lines.length && !lines[i].startsWith('```')) {
        codeLines.push(lines[i]);
        i++;
      }
      i++; // skip closing ```
      elements.push(
        <pre
          key={`code-${elements.length}`}
          className="bg-foreground/5 rounded-md p-3 my-2 overflow-x-auto text-xs"
        >
          <code>
            {lang && (
              <span className="text-muted-foreground text-[10px] block mb-1">
                {lang}
              </span>
            )}
            {codeLines.join('\n')}
          </code>
        </pre>
      );
      continue;
    }

    // Heading (## or ###)
    if (line.startsWith('### ')) {
      elements.push(
        <h4
          key={`h-${elements.length}`}
          className="font-semibold text-sm mt-3 mb-1"
        >
          {line.slice(4)}
        </h4>
      );
      i++;
      continue;
    }
    if (line.startsWith('## ')) {
      elements.push(
        <h3
          key={`h-${elements.length}`}
          className="font-semibold text-base mt-3 mb-1"
        >
          {line.slice(3)}
        </h3>
      );
      i++;
      continue;
    }

    // Unordered list item
    if (line.match(/^[-*]\s/)) {
      const listItems: string[] = [];
      while (i < lines.length && lines[i].match(/^[-*]\s/)) {
        listItems.push(lines[i].replace(/^[-*]\s/, ''));
        i++;
      }
      elements.push(
        <ul
          key={`ul-${elements.length}`}
          className="list-disc pl-5 my-1.5 space-y-0.5"
        >
          {listItems.map((item, idx) => (
            <li key={idx}>{renderInline(item)}</li>
          ))}
        </ul>
      );
      continue;
    }

    // Ordered list item
    if (line.match(/^\d+\.\s/)) {
      const listItems: string[] = [];
      while (i < lines.length && lines[i].match(/^\d+\.\s/)) {
        listItems.push(lines[i].replace(/^\d+\.\s/, ''));
        i++;
      }
      elements.push(
        <ol
          key={`ol-${elements.length}`}
          className="list-decimal pl-5 my-1.5 space-y-0.5"
        >
          {listItems.map((item, idx) => (
            <li key={idx}>{renderInline(item)}</li>
          ))}
        </ol>
      );
      continue;
    }

    // Empty line
    if (line.trim() === '') {
      elements.push(<br key={`br-${elements.length}`} />);
      i++;
      continue;
    }

    // Regular paragraph
    elements.push(
      <p key={`p-${elements.length}`} className="my-0.5">
        {renderInline(line)}
      </p>
    );
    i++;
  }

  return elements;
}

/**
 * Render inline markdown: **bold**, `code`, and regular text.
 */
function renderInline(text: string): React.ReactNode {
  const parts: React.ReactNode[] = [];
  // Match **bold** and `code` patterns
  const regex = /(\*\*(.+?)\*\*)|(`(.+?)`)/g;
  let lastIndex = 0;
  let match: RegExpExecArray | null;

  while ((match = regex.exec(text)) !== null) {
    // Text before match
    if (match.index > lastIndex) {
      parts.push(text.slice(lastIndex, match.index));
    }

    if (match[2]) {
      // Bold
      parts.push(
        <strong key={`b-${lastIndex}`} className="font-semibold">
          {match[2]}
        </strong>
      );
    } else if (match[4]) {
      // Inline code
      parts.push(
        <code
          key={`c-${lastIndex}`}
          className="bg-foreground/5 rounded px-1 py-0.5 text-xs"
        >
          {match[4]}
        </code>
      );
    }

    lastIndex = match.index + match[0].length;
  }

  if (lastIndex < text.length) {
    parts.push(text.slice(lastIndex));
  }

  return parts.length === 0 ? text : parts;
}
