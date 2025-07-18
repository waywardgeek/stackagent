import React from 'react';
import { useAppStore } from '@/store';
import { User, Bot } from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';

export const MessageList: React.FC = () => {
  const { messages } = useAppStore();
  
  return (
    <div className="scrollable-container p-4 space-y-4">
      {messages.map((message) => (
        <div
          key={message.id}
          className={`flex ${message.type === 'user' ? 'justify-end' : 'justify-start'}`}
        >
          <div
            className={`max-w-[80%] rounded-lg px-4 py-2 ${
              message.type === 'user'
                ? 'chat-message-user'
                : 'chat-message-assistant'
            }`}
          >
            <div className="flex items-center space-x-2 mb-1">
              {message.type === 'user' ? (
                <User className="w-4 h-4" />
              ) : (
                <Bot className="w-4 h-4" />
              )}
              <span className="text-sm font-medium">
                {message.type === 'user' ? 'You' : 'StackAgent'}
              </span>
              <span className="text-xs opacity-75">
                {formatDistanceToNow(new Date(message.timestamp), { addSuffix: true })}
              </span>
            </div>
            <div className="text-sm whitespace-pre-wrap">
              {message.content}
            </div>
            {message.cost && (
              <div className="text-xs opacity-75 mt-1">
                Cost: ${message.cost.toFixed(4)} • {message.tokens} tokens
              </div>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}; 