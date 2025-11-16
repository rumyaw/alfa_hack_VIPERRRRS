'use client'

import { useState, useEffect } from 'react'
import { api } from '@/lib/api'

interface User {
  id: string
  username: string
  specialization: string
  created_at: string
}

interface Stats {
  files_count: number
  messages_count: number
}

export default function Account() {
  const [user, setUser] = useState<User | null>(null)
  const [stats, setStats] = useState<Stats | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadUserData()
  }, [])

  const loadUserData = async () => {
    try {
      setLoading(true)
      const response = await api.get('/user')
      setUser(response.data.user)
      setStats(response.data.stats)
    } catch (error) {
      console.error('Failed to load user data:', error)
    } finally {
      setLoading(false)
    }
  }

  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    return date.toLocaleDateString('ru-RU', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    })
  }

  if (loading) {
    return (
      <div className="bg-white dark:bg-[#1a1a1a] rounded-xl shadow-sm border border-gray-200 dark:border-zinc-800 p-6">
        <div className="text-center py-8 text-gray-500 dark:text-gray-400">Загрузка...</div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Информация о пользователе */}
      <div className="bg-white dark:bg-[#1a1a1a] rounded-xl shadow-sm border border-gray-200 dark:border-zinc-800 p-6">
        <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100 mb-4">
          👤 Информация об аккаунте
        </h2>
        {user && (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Имя пользователя
              </label>
              <p className="text-lg font-semibold text-gray-900 dark:text-gray-100">{user.username}</p>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Специализация бизнеса
              </label>
              <p className="text-lg text-gray-900 dark:text-gray-100">{user.specialization}</p>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Дата регистрации
              </label>
              <p className="text-gray-600 dark:text-gray-400">{formatDate(user.created_at)}</p>
            </div>
          </div>
        )}
      </div>

      {/* Статистика */}
      {stats && (
        <div className="bg-white dark:bg-[#1a1a1a] rounded-xl shadow-sm border border-gray-200 dark:border-zinc-800 p-6">
          <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100 mb-4">
            📊 Статистика
          </h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div className="bg-gray-50 dark:bg-zinc-900 rounded-xl p-4 border border-gray-200 dark:border-zinc-700">
              <div className="text-sm text-gray-600 dark:text-gray-400 mb-1">Загружено файлов</div>
              <div className="text-2xl font-bold text-alfa-red">{stats.files_count}</div>
            </div>
            <div className="bg-gray-50 dark:bg-zinc-900 rounded-xl p-4 border border-gray-200 dark:border-zinc-700">
              <div className="text-sm text-gray-600 dark:text-gray-400 mb-1">Сообщений в чате</div>
              <div className="text-2xl font-bold text-alfa-red">{stats.messages_count}</div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

