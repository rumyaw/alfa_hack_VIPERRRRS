'use client'

import { useState, useEffect } from 'react'
import { chatsAPI, Chat } from '@/lib/api'

interface ChatListProps {
  selectedChatId: string | null
  onSelectChat: (chatId: string) => void
  onNewChat: () => void
}

export default function ChatList({ selectedChatId, onSelectChat, onNewChat }: ChatListProps) {
  const [chats, setChats] = useState<Chat[]>([])
  const [loading, setLoading] = useState(true)

  const loadChats = async () => {
    try {
      setLoading(true)
      const response = await chatsAPI.getAll()
      setChats(response.chats || [])
    } catch (error) {
      console.error('Failed to load chats:', error)
      setChats([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadChats()
    
    // Обновляем список при создании нового чата
    const handleChatCreated = () => {
      loadChats()
    }
    window.addEventListener('chat-created', handleChatCreated)
    return () => window.removeEventListener('chat-created', handleChatCreated)
  }, [])

  const handleDelete = async (e: React.MouseEvent, chatId: string) => {
    e.stopPropagation()
    if (!confirm('Вы уверены, что хотите удалить этот чат?')) return

    try {
      await chatsAPI.delete(chatId)
      setChats(chats.filter((c) => c.id !== chatId))
      if (selectedChatId === chatId) {
        onNewChat()
      }
    } catch (error: any) {
      alert(error.response?.data?.error || 'Ошибка при удалении чата')
    }
  }

  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    const now = new Date()
    const diff = now.getTime() - date.getTime()
    const days = Math.floor(diff / (1000 * 60 * 60 * 24))

    if (days === 0) {
      return date.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
    } else if (days === 1) {
      return 'Вчера'
    } else if (days < 7) {
      return `${days} дн. назад`
    } else {
      return date.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' })
    }
  }

  return (
    <div className="bg-white dark:bg-[#1a1a1a] rounded-xl shadow-sm h-full flex flex-col border border-gray-200 dark:border-zinc-800">
      <div className="p-4 border-b border-gray-200 dark:border-zinc-800">
        <button
          onClick={onNewChat}
          className="w-full bg-gradient-to-r from-alfa-red to-red-600 text-white py-2.5 px-4 rounded-xl font-semibold hover:from-red-600 hover:to-red-700 transition-all flex items-center justify-center gap-2 shadow-sm"
        >
          <span className="text-lg">+</span>
          <span>Новый чат</span>
        </button>
      </div>

      <div className="flex-1 overflow-y-auto p-2">
        {loading ? (
          <div className="text-center py-4 text-gray-500 dark:text-gray-400 text-sm">Загрузка...</div>
        ) : chats.length === 0 ? (
          <div className="text-center py-8 text-gray-500 dark:text-gray-400 text-sm">
            <p>Нет чатов</p>
            <p className="mt-2">Создайте новый чат</p>
          </div>
        ) : (
          <div className="space-y-1">
            {chats.map((chat) => (
              <div
                key={chat.id}
                onClick={() => onSelectChat(chat.id)}
                className={`p-3 rounded-xl cursor-pointer transition-all group ${
                  selectedChatId === chat.id
                    ? 'bg-alfa-red text-white shadow-sm'
                    : 'hover:bg-gray-100 dark:hover:bg-zinc-800 text-gray-900 dark:text-gray-100'
                }`}
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1 min-w-0">
                    <p
                      className={`font-medium truncate text-sm ${
                        selectedChatId === chat.id ? 'text-white' : 'text-gray-900 dark:text-gray-100'
                      }`}
                    >
                      {chat.title || 'Без названия'}
                    </p>
                    <p
                      className={`text-xs mt-1 ${
                        selectedChatId === chat.id ? 'text-red-100' : 'text-gray-500 dark:text-gray-400'
                      }`}
                    >
                      {formatDate(chat.updated_at)}
                    </p>
                  </div>
                  <button
                    onClick={(e) => handleDelete(e, chat.id)}
                    className={`ml-2 opacity-0 group-hover:opacity-100 transition-opacity text-lg leading-none ${
                      selectedChatId === chat.id ? 'text-white hover:text-red-200' : 'text-gray-400 dark:text-gray-500 hover:text-red-600 dark:hover:text-red-500'
                    }`}
                  >
                    ×
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

