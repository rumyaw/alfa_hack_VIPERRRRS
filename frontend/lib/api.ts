import axios from 'axios'

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api'

const api = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Хранилище токена только для авторизации на backend
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

export interface RegisterData {
  username: string
  password: string
  business_name: string
  specialization: string
}

export interface LoginData {
  username: string
  password: string
}

export interface ChatMessage {
  message: string
  category?: string
  chat_id?: string
}

export const authAPI = {
  register: async (data: RegisterData) => {
    const response = await api.post('/register', data)
    return response.data
  },
  login: async (data: LoginData) => {
    const response = await api.post('/login', data)
    return response.data
  },
}

export const filesAPI = {
  upload: async (file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    const response = await api.post('/files/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    return response.data
  },
  getAll: async () => {
    const response = await api.get('/files')
    return response.data
  },
  delete: async (id: string) => {
    const response = await api.delete(`/files/${id}`)
    return response.data
  },
}

export interface Chat {
  id: string
  title: string
  created_at: string
  updated_at: string
}

export const chatAPI = {
  sendMessage: async (data: ChatMessage) => {
    const response = await api.post('/chat', data)
    return response.data
  },
  getHistory: async (chatId: string) => {
    const response = await api.get(`/chat/${chatId}/history`)
    return response.data
  },
}

export const chatsAPI = {
  create: async (title?: string) => {
    const response = await api.post('/chats', { title: title || 'Новый чат' })
    return response.data
  },
  getAll: async () => {
    const response = await api.get('/chats')
    return response.data
  },
  delete: async (id: string) => {
    const response = await api.delete(`/chats/${id}`)
    return response.data
  },
}

export const apiUser = {
  getCurrent: async () => {
    const response = await api.get('/user')
    return response.data
  },
}

export { api }

