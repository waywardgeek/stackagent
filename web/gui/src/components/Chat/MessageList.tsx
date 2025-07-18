import React, { useEffect, useRef } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { useAppStore, selectMessages } from '@/store';
import MessageBubble from './MessageBubble';
import LiveOperationsDisplay from './LiveOperationsDisplay';

interface MessageListProps {
  className?: string;
}

const MessageList: React.FC<MessageListProps> = ({ className = '' }) => {
  const messages = useAppStore(selectMessages);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to bottom when new messages arrive
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  return (
    <div className={`flex-1 overflow-y-auto p-4 space-y-4 ${className}`}>
      {/* Real-time operations display at the top */}
      <LiveOperationsDisplay />
      
      <AnimatePresence mode="popLayout">
        {messages.map((message) => (
          <motion.div
            key={message.id}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            transition={{ duration: 0.3 }}
          >
            <MessageBubble message={message} />
          </motion.div>
        ))}
      </AnimatePresence>
      
      {/* Invisible div to scroll to */}
      <div ref={messagesEndRef} />
    </div>
  );
};

export default MessageList; 