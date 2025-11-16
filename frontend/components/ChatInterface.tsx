'use client'

import { useState, useEffect, useRef } from 'react'
import { chatAPI } from '@/lib/api'
import { useTheme } from 'next-themes'
import { 
  MessageCircle, 
  PiggyBank, 
  Landmark, 
  Users2, 
  Megaphone,
  TrendingUp,
  FileText,
  BarChart3,
  DollarSign,
  Scale,
  Briefcase,
  Target,
  Send
} from 'lucide-react'

interface Message {
  id: string
  message: string
  response: string
  category?: string
  created_at: string
  chat_id?: string
}

interface ChatInterfaceProps {
  chatId: string | null
  onChatCreated?: (chatId: string) => void
}

const CATEGORY_BLOCKS = [
  { 
    id: 'financial', 
    title: 'Финансовый анализ', 
    description: 'Анализ прибыли, выручки и расходов',
    icon: <DollarSign size={24} />,
    color: 'from-green-500 to-emerald-600',
    prompt: 'Проанализируй финансовые показатели моего бизнеса. Какая прибыль, выручка и расходы?'
  },
  { 
    id: 'legal', 
    title: 'Юридические вопросы', 
    description: 'Помощь с правовыми аспектами',
    icon: <Scale size={24} />,
    color: 'from-blue-500 to-indigo-600',
    prompt: 'Помоги с юридическими вопросами для моего бизнеса. Что нужно знать?'
  },
  { 
    id: 'hr', 
    title: 'Управление персоналом', 
    description: 'Вопросы по сотрудникам и кадрам',
    icon: <Users2 size={24} />,
    color: 'from-purple-500 to-pink-600',
    prompt: 'Проанализируй информацию о персонале. Какие рекомендации по управлению сотрудниками?'
  },
  { 
    id: 'marketing', 
    title: 'Маркетинг и продвижение', 
    description: 'Стратегии роста и привлечения клиентов',
    icon: <Target size={24} />,
    color: 'from-orange-500 to-red-600',
    prompt: 'Дай рекомендации по маркетингу и продвижению моего бизнеса. Как увеличить продажи?'
  },
  { 
    id: 'growth', 
    title: 'Рост бизнеса', 
    description: 'Стратегии развития и масштабирования',
    icon: <TrendingUp size={24} />,
    color: 'from-cyan-500 to-blue-600',
    prompt: 'Как мой бизнес может расти? Какие стратегии развития ты можешь предложить?'
  },
  { 
    id: 'reports', 
    title: 'Анализ отчетов', 
    description: 'Детальный анализ загруженных данных',
    icon: <BarChart3 size={24} />,
    color: 'from-violet-500 to-purple-600',
    prompt: 'Проанализируй все загруженные отчеты и файлы. Какие выводы можно сделать?'
  },
]

export default function ChatInterface({ chatId, onChatCreated }: ChatInterfaceProps) {
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [selectedCategory, setSelectedCategory] = useState<string>('')
  const [loading, setLoading] = useState(false)
  const [currentChatId, setCurrentChatId] = useState<string | null>(chatId)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const [errorMessage, setErrorMessage] = useState('')

  useEffect(() => {
    setCurrentChatId(chatId)
    if (chatId) {
      loadHistory(chatId)
    } else {
      setMessages([])
    }
  }, [chatId])

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  const loadHistory = async (chatIdToLoad: string) => {
    try {
      const response = await chatAPI.getHistory(chatIdToLoad)
      if (response.messages && Array.isArray(response.messages)) {
        setMessages(response.messages)
      } else {
        setMessages([])
      }
    } catch (error) {
      console.error('Failed to load chat history:', error)
      setMessages([])
    }
  }

  const handleCategoryClick = (categoryId: string, prompt: string) => {
    setSelectedCategory(categoryId)
    setInput(prompt)
  }

  const handleSend = async (e?: React.FormEvent, customMessage?: string) => {
    if (e) e.preventDefault()
    const messageToSend = customMessage || input.trim()
    if (!messageToSend || loading) return
    setErrorMessage('')

    setInput('')
    setLoading(true)

    // Добавляем сообщение пользователя сразу
    const tempMessage: Message = {
      id: 'temp',
      message: messageToSend,
      response: '',
      category: selectedCategory,
      created_at: new Date().toISOString(),
      chat_id: currentChatId || undefined,
    }
    setMessages((prev) => [...prev, tempMessage])

    try {
      const response = await chatAPI.sendMessage({
        message: messageToSend,
        category: selectedCategory || undefined,
        chat_id: currentChatId || undefined,
      })

      // Если создан новый чат, обновляем текущий chatId
      if (response.chat_id && !currentChatId) {
        setCurrentChatId(response.chat_id)
        if (onChatCreated) {
          onChatCreated(response.chat_id)
        }
        window.dispatchEvent(new Event('chat-created'))
      }

      // Заменяем временное сообщение на реальное
      setMessages((prev) =>
        prev.map((msg) =>
          msg.id === 'temp'
            ? {
                id: response.id,
                message: response.message,
                response: response.response,
                category: response.category,
                created_at: response.created_at,
                chat_id: response.chat_id,
              }
            : msg
        )
      )
      setSelectedCategory('')
    } catch (error: any) {
      console.error('Failed to send message:', error)
      setMessages((prev) =>
        prev.filter((msg) => msg.id !== 'temp')
      )
      setErrorMessage(error?.message || error?.response?.data?.error || 'Ошибка при отправке сообщения')
    } finally {
      setLoading(false)
    }
  }

  const { theme } = useTheme()

  return (
    <div className="bg-white dark:bg-[#1a1a1a] rounded-xl shadow-sm h-full flex flex-col border border-gray-200 dark:border-zinc-800">
      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-4 md:p-6 space-y-4">
        {messages.length === 0 ? (
          <div className="h-full flex flex-col items-center justify-center py-12 px-4">
            <div className="text-center mb-8">
              <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-alfa-red/10 dark:bg-alfa-red/20 mb-4">
                <MessageCircle className="text-alfa-red" size={32} />
              </div>
              <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-2">
                Добро пожаловать!
              </h2>
              <p className="text-gray-600 dark:text-gray-400 mb-8">
                Выберите тему для начала разговора или задайте свой вопрос
              </p>
            </div>
            
            {/* Category Blocks Grid */}
            <div className="w-full max-w-4xl grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 sm:gap-4">
              {CATEGORY_BLOCKS.map((block) => (
                <button
                  key={block.id}
                  onClick={() => handleCategoryClick(block.id, block.prompt)}
                  className={`group relative overflow-hidden rounded-xl p-4 sm:p-6 text-left transition-all duration-300 hover:scale-[1.02] hover:shadow-lg bg-gradient-to-br ${block.color} border border-transparent hover:border-white/20`}
                >
                  <div className="relative z-10">
                    <div className="flex items-center justify-between mb-3">
                      <div className="p-2 bg-white/20 dark:bg-black/20 rounded-lg backdrop-blur-sm">
                        <div className="text-white">
                          {block.icon}
                        </div>
                      </div>
                    </div>
                    <h3 className="text-white font-semibold text-base sm:text-lg mb-1">
                      {block.title}
                    </h3>
                    <p className="text-white/80 text-xs sm:text-sm leading-relaxed">
                      {block.description}
                    </p>
                  </div>
                  <div className="absolute inset-0 bg-gradient-to-t from-black/20 to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />
                </button>
              ))}
            </div>
          </div>
        ) : (
          <>
            {messages.map((msg) => (
              <div key={msg.id} className="space-y-3">
                {/* User Message */}
                <div className="flex justify-end">
                  <div className="bg-alfa-red text-white rounded-2xl rounded-tr-sm px-3 py-2.5 sm:px-4 sm:py-3 max-w-[90%] sm:max-w-[85%] md:max-w-[70%] shadow-sm break-words overflow-wrap-anywhere">
                    <p className="text-xs sm:text-sm leading-relaxed break-words overflow-wrap-anywhere word-break-break-word">{msg.message}</p>
                  </div>
                </div>
                {/* Bot Response */}
                {msg.response && (
                  <div className="flex justify-start">
                    <div className="bg-gray-100 dark:bg-zinc-800 text-gray-900 dark:text-gray-100 rounded-2xl rounded-tl-sm px-3 py-2.5 sm:px-4 sm:py-3 max-w-[90%] sm:max-w-[85%] md:max-w-[70%] shadow-sm break-words overflow-wrap-anywhere">
                      <p className="text-xs sm:text-sm leading-relaxed whitespace-pre-wrap break-words overflow-wrap-anywhere word-break-break-word">{msg.response}</p>
                    </div>
                  </div>
                )}
                {msg.id === 'temp' && loading && (
                  <div className="flex justify-start">
                    <div className="bg-gray-100 dark:bg-zinc-800 rounded-2xl rounded-tl-sm px-4 py-3">
                      <div className="flex space-x-1.5">
                        <div className="w-2 h-2 bg-gray-400 dark:bg-gray-500 rounded-full animate-bounce"></div>
                        <div className="w-2 h-2 bg-gray-400 dark:bg-gray-500 rounded-full animate-bounce" style={{ animationDelay: '0.1s' }}></div>
                        <div className="w-2 h-2 bg-gray-400 dark:bg-gray-500 rounded-full animate-bounce" style={{ animationDelay: '0.2s' }}></div>
                      </div>
                    </div>
                  </div>
                )}
              </div>
            ))}
            {errorMessage && (
              <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-800 dark:text-red-200 text-sm px-4 py-3 rounded-lg">
                {errorMessage}
              </div>
            )}
          </>
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <form onSubmit={handleSend} className="p-3 sm:p-4 border-t border-gray-200 dark:border-zinc-800 bg-white dark:bg-[#1a1a1a]">
        <div className="flex gap-2">
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder="Задайте вопрос..."
            disabled={loading}
            className="flex-1 px-3 py-2.5 sm:px-4 sm:py-3 text-sm sm:text-base bg-gray-50 dark:bg-zinc-900 border border-gray-200 dark:border-zinc-700 rounded-xl focus:ring-2 focus:ring-alfa-red focus:border-transparent outline-none disabled:opacity-50 disabled:cursor-not-allowed text-gray-900 dark:text-gray-100 placeholder-gray-500 dark:placeholder-gray-400"
          />
          <button
            type="submit"
            disabled={loading || !input.trim()}
            className="bg-alfa-red text-white px-4 py-2.5 sm:px-6 sm:py-3 rounded-xl text-sm sm:text-base font-medium hover:bg-red-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center min-w-[80px] sm:min-w-[100px] shadow-sm"
          >
            {loading ? (
              <div className="w-4 h-4 sm:w-5 sm:h-5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
            ) : (
              <>
                <Send size={16} className="sm:mr-2 sm:block hidden" />
                <span className="hidden sm:inline">Отправить</span>
                <Send size={18} className="sm:hidden" />
              </>
            )}
          </button>
        </div>
        {selectedCategory && (
          <div className="mt-2 text-xs text-gray-500 dark:text-gray-400">
            Категория: {CATEGORY_BLOCKS.find((c) => c.id === selectedCategory)?.title}
          </div>
        )}
      </form>
    </div>
  )
}

