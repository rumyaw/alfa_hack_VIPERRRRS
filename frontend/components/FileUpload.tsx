'use client'

import { useState } from 'react'
import { filesAPI } from '@/lib/api'

export default function FileUpload() {
  const [file, setFile] = useState<File | null>(null)
  const [loading, setLoading] = useState(false)
  const [success, setSuccess] = useState(false)
  const [error, setError] = useState('')

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      setFile(e.target.files[0])
      setSuccess(false)
      setError('')
    }
  }

  const handleUpload = async () => {
    if (!file) return

    setLoading(true)
    setError('')
    setSuccess(false)

    try {
      await filesAPI.upload(file)
      setSuccess(true)
      setFile(null)
      // Сброс input
      const fileInput = document.getElementById('file-input') as HTMLInputElement
      if (fileInput) fileInput.value = ''
      
      // Обновление списка файлов через событие
      window.dispatchEvent(new Event('files-updated'))
    } catch (err: any) {
      setError(err.response?.data?.error || 'Ошибка при загрузке файла')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="bg-white dark:bg-[#1a1a1a] rounded-xl shadow-sm border border-gray-200 dark:border-zinc-800 p-6">
      <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100 mb-4">
        📤 Загрузить файл
      </h2>
      <div className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Выберите файл
          </label>
          <input
            id="file-input"
            type="file"
            onChange={handleFileChange}
            className="block w-full text-sm text-gray-500 dark:text-gray-400 file:mr-4 file:py-2 file:px-4 file:rounded-xl file:border-0 file:text-sm file:font-semibold file:bg-gradient-to-r file:from-alfa-red file:to-red-600 file:text-white hover:file:from-red-600 hover:file:to-red-700 file:cursor-pointer cursor-pointer"
          />
        </div>

        {file && (
          <div className="bg-gray-50 dark:bg-zinc-900 rounded-xl p-4 border border-gray-200 dark:border-zinc-700">
            <p className="text-sm text-gray-700 dark:text-gray-300">
              <span className="font-medium">Файл:</span> {file.name}
            </p>
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
              Размер: {(file.size / 1024).toFixed(2)} KB
            </p>
          </div>
        )}

        {error && (
          <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-800 dark:text-red-200 px-4 py-3 rounded-xl text-sm">
            {error}
          </div>
        )}

        {success && (
          <div className="bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 text-green-800 dark:text-green-200 px-4 py-3 rounded-xl text-sm">
            Файл успешно загружен!
          </div>
        )}

        <button
          onClick={handleUpload}
          disabled={!file || loading}
          className="w-full bg-gradient-to-r from-alfa-red to-red-600 text-white py-3 rounded-xl font-semibold hover:from-red-600 hover:to-red-700 transition-all disabled:opacity-50 disabled:cursor-not-allowed shadow-sm"
        >
          {loading ? (
            <span className="flex items-center justify-center">
              <div className="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin mr-2" />
              Загрузка...
            </span>
          ) : (
            'Загрузить файл'
          )}
        </button>
      </div>
    </div>
  )
}

