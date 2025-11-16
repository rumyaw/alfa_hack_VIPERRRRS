'use client'

import { useState, useEffect } from 'react'
import { filesAPI } from '@/lib/api'

interface File {
  id: string
  filename: string
  file_type: string
  file_size: number
  uploaded_at: string
}

export default function FileList() {
  const [files, setFiles] = useState<File[]>([])
  const [loading, setLoading] = useState(true)
  const [deleting, setDeleting] = useState<string | null>(null)

  const loadFiles = async () => {
    try {
      setLoading(true)
      const response = await filesAPI.getAll()
      setFiles(response.files || [])
    } catch (error) {
      console.error('Failed to load files:', error)
      setFiles([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadFiles()

    // Слушаем событие обновления файлов
    const handleFilesUpdated = () => {
      loadFiles()
    }
    window.addEventListener('files-updated', handleFilesUpdated)
    return () => window.removeEventListener('files-updated', handleFilesUpdated)
  }, [])

  const handleDelete = async (id: string) => {
    if (!confirm('Вы уверены, что хотите удалить этот файл?')) return

    setDeleting(id)
    try {
      await filesAPI.delete(id)
      setFiles(files.filter((f) => f.id !== id))
    } catch (error: any) {
      alert(error.response?.data?.error || 'Ошибка при удалении файла')
    } finally {
      setDeleting(null)
    }
  }

  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    return date.toLocaleString('ru-RU')
  }

  const formatSize = (bytes: number) => {
    if (bytes < 1024) return bytes + ' B'
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(2) + ' KB'
    return (bytes / (1024 * 1024)).toFixed(2) + ' MB'
  }

  return (
    <div className="bg-white dark:bg-[#1a1a1a] rounded-xl shadow-sm border border-gray-200 dark:border-zinc-800 p-6">
      <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100 mb-4">
        📁 Загруженные файлы
      </h2>

      {loading ? (
        <div className="text-center py-8 text-gray-500 dark:text-gray-400">Загрузка...</div>
      ) : files.length === 0 ? (
        <div className="text-center py-8 text-gray-500 dark:text-gray-400">
          <p>Файлы не загружены</p>
          <p className="text-sm mt-2">
            Загрузите файлы с данными о вашем бизнесе для более точных ответов
          </p>
        </div>
      ) : (
        <div className="space-y-2">
          {files.map((file) => (
            <div
              key={file.id}
              className="flex flex-col sm:flex-row items-start sm:items-center justify-between p-4 border border-gray-200 dark:border-zinc-700 rounded-xl hover:bg-gray-50 dark:hover:bg-zinc-900 transition-colors gap-3"
            >
              <div className="flex-1 min-w-0 w-full sm:w-auto">
                <div className="flex items-center space-x-3">
                  <span className="text-2xl flex-shrink-0">
                    {file.file_type === 'pdf'
                      ? '📄'
                      : file.file_type === 'doc' || file.file_type === 'docx'
                      ? '📝'
                      : file.file_type === 'xls' || file.file_type === 'xlsx'
                      ? '📊'
                      : file.file_type === 'txt'
                      ? '📃'
                      : '📎'}
                  </span>
                  <div className="min-w-0 flex-1">
                    <p className="font-medium text-gray-900 dark:text-gray-100 truncate">{file.filename}</p>
                    <p className="text-sm text-gray-500 dark:text-gray-400">
                      {formatSize(file.file_size)} • {formatDate(file.uploaded_at)}
                    </p>
                  </div>
                </div>
              </div>
              <button
                onClick={() => handleDelete(file.id)}
                disabled={deleting === file.id}
                className="w-full sm:w-auto px-4 py-2 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors disabled:opacity-50 text-sm font-medium"
              >
                {deleting === file.id ? 'Удаление...' : 'Удалить'}
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

