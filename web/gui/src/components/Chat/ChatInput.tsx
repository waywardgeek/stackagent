import React, { useState } from 'react';
import { Send } from 'lucide-react';
import { useAppStore } from '@/store';
import type { WebSocketEventType } from '@/types';

interface ChatInputProps {
  sendMessage: (type: WebSocketEventType, data: any) => void;
}

export const ChatInput: React.FC<ChatInputProps> = ({ sendMessage }) => {
  const [message, setMessage] = useState('');
  const { addMessage } = useAppStore();
  
  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (message.trim()) {
      // Generate unique ID for the message
      const messageId = Date.now().toString();
      
      // Add user message to store immediately
      addMessage({
        id: messageId,
        sessionId: 'current',
        type: 'user',
        content: message,
        timestamp: new Date(),
      });
      
      // Send message to backend via WebSocket
      sendMessage('chat_message', {
        id: messageId,
        message: message,
      });
      
      setMessage('');
    }
  };
  
  return (
    <form onSubmit={handleSubmit} className="p-4">
      <div className="flex space-x-2">
        <input
          type="text"
          value={message}
          onChange={(e) => setMessage(e.target.value)}
          placeholder="Ask StackAgent anything..."
          className="flex-1 input"
          data-chat-input
        />
        <button
          type="submit"
          disabled={!message.trim()}
          className="btn-primary px-3 py-2 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <Send className="w-4 h-4" />
        </button>
      </div>
    </form>
  );
}; 