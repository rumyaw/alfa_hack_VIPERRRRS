'use client'

import { useState } from 'react'
import ChatInterface from './ChatInterface'
import ChatList from './ChatList'
import FileUpload from './FileUpload'
import FileList from './FileList'
import Account from './Account'
import { useTheme } from 'next-themes'
import { Sun, Moon, MessageSquare, FolderOpen, User, Menu, X } from 'lucide-react'

interface DashboardProps {
  token: string
  onLogout: () => void
}

export default function Dashboard({ onLogout }: DashboardProps) {
  const [activeTab, setActiveTab] = useState<'chat' | 'files' | 'account'>('chat')
  const [selectedChatId, setSelectedChatId] = useState<string | null>(null)
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)
  const { theme, setTheme } = useTheme()

  const tabs = [
    { id: 'chat' as const, label: 'Чат-бот', icon: <MessageSquare size={20} /> },
    { id: 'files' as const, label: 'Файлы', icon: <FolderOpen size={20} /> },
    { id: 'account' as const, label: 'Аккаунт', icon: <User size={20} /> },
  ]

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-[#0f0f0f] transition-colors">
      {/* Header */}
      <header className="bg-white dark:bg-[#1a1a1a] border-b border-gray-200 dark:border-zinc-800 sticky top-0 z-50 backdrop-blur-sm bg-white/95 dark:bg-[#1a1a1a]/95">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center space-x-3">
              <div className="bg-gradient-to-br from-alfa-red to-red-600 text-white px-3 py-1.5 rounded-lg text-lg font-bold shadow-sm">
                АЛЬФА
              </div>
              <h1 className="hidden sm:block text-lg font-semibold text-gray-900 dark:text-gray-100">
                Помощник для бизнеса
              </h1>
            </div>
            <div className="flex items-center gap-3">
              <button
                onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
                className="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-zinc-800 transition-colors text-gray-700 dark:text-gray-300"
                aria-label="Переключить тему"
              >
                {theme === 'dark' ? <Sun size={20} /> : <Moon size={20} />}
              </button>
              <button
                onClick={onLogout}
                className="px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:text-alfa-red transition-colors"
              >
                Выйти
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-3 sm:px-4 md:px-6 lg:px-8 py-4 sm:py-6">
        {/* Tabs - Desktop */}
        <div className="hidden md:flex bg-white dark:bg-[#1a1a1a] rounded-xl shadow-sm mb-4 sm:mb-6 border border-gray-200 dark:border-zinc-800 overflow-hidden">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`flex-1 px-6 py-4 text-center font-medium transition-all flex items-center justify-center gap-2 ${
                activeTab === tab.id
                  ? 'text-alfa-red bg-alfa-red/5 dark:bg-alfa-red/10 border-b-2 border-alfa-red'
                  : 'text-gray-600 dark:text-gray-400 hover:text-alfa-red hover:bg-gray-50 dark:hover:bg-zinc-900'
              }`}
            >
              {tab.icon}
              <span>{tab.label}</span>
            </button>
          ))}
        </div>

        {/* Tabs - Mobile */}
        <div className="md:hidden mb-4">
          <button
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            className="w-full bg-white dark:bg-[#1a1a1a] rounded-xl px-4 py-3 flex items-center justify-between border border-gray-200 dark:border-zinc-800"
          >
            <div className="flex items-center gap-2">
              {tabs.find(t => t.id === activeTab)?.icon}
              <span className="font-medium text-gray-900 dark:text-gray-100">
                {tabs.find(t => t.id === activeTab)?.label}
              </span>
            </div>
            {mobileMenuOpen ? <X size={20} /> : <Menu size={20} />}
          </button>
          {mobileMenuOpen && (
            <div className="mt-2 bg-white dark:bg-[#1a1a1a] rounded-xl shadow-lg border border-gray-200 dark:border-zinc-800 overflow-hidden">
              {tabs.map((tab) => (
                <button
                  key={tab.id}
                  onClick={() => {
                    setActiveTab(tab.id)
                    setMobileMenuOpen(false)
                  }}
                  className={`w-full px-4 py-3 text-left flex items-center gap-3 transition-colors ${
                    activeTab === tab.id
                      ? 'bg-alfa-red/5 dark:bg-alfa-red/10 text-alfa-red'
                      : 'text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-zinc-900'
                  }`}
                >
                  {tab.icon}
                  <span>{tab.label}</span>
                </button>
              ))}
            </div>
          )}
        </div>

        {/* Tab Content */}
        {activeTab === 'chat' && (
          <div className="flex flex-col lg:flex-row gap-4 h-[calc(100vh-220px)] lg:h-[calc(100vh-200px)]">
            <div className="hidden lg:block lg:w-64 flex-shrink-0">
              <ChatList
                selectedChatId={selectedChatId}
                onSelectChat={setSelectedChatId}
                onNewChat={() => setSelectedChatId(null)}
              />
            </div>
            <div className="flex-1 min-h-0">
              <ChatInterface
                chatId={selectedChatId}
                onChatCreated={setSelectedChatId}
              />
            </div>
          </div>
        )}
        {activeTab === 'files' && (
          <div className="space-y-6">
            <FileUpload />
            <FileList />
          </div>
        )}
        {activeTab === 'account' && <Account />}
      </main>
    </div>
  )
}

