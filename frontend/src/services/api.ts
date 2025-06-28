// Base API client configuration and utilities
export interface ApiResponse<T = unknown> {
  data: T
  status: number
  statusText: string
  headers: Record<string, string>
}

export interface ApiErrorData {
  error?: string
  message?: string
  details?: string | Record<string, unknown>
}

export class ApiError extends Error {
  status?: number
  statusText?: string
  data?: ApiErrorData

  constructor(options: {
    message: string
    status?: number
    statusText?: string
    data?: ApiErrorData
  }) {
    super(options.message)
    this.name = 'ApiError'
    this.status = options.status
    this.statusText = options.statusText
    this.data = options.data
  }
}

export class ApiClient {
  private baseURL: string
  private defaultHeaders: Record<string, string>

  constructor(baseURL: string = 'http://localhost:9090') {
    this.baseURL = baseURL
    this.defaultHeaders = {
      'Content-Type': 'application/json',
      'Accept': 'application/json'
    }
  }

  private async request<T>(
    endpoint: string, 
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    const url = `${this.baseURL}${endpoint}`
    
    const config: RequestInit = {
      headers: {
        ...this.defaultHeaders,
        ...options.headers
      },
      ...options
    }

    try {
      const response = await fetch(url, config)
      
      // Extract headers
      const headers: Record<string, string> = {}
      response.headers.forEach((value, key) => {
        headers[key] = value
      })

      // Parse response body
      let data: T
      const contentType = response.headers.get('content-type')
      
      if (contentType && contentType.includes('application/json')) {
        data = await response.json()
      } else {
        data = await response.text() as T
      }

      // Check if response is successful
      if (!response.ok) {
        throw new ApiError({
          message: `Request failed: ${response.status} ${response.statusText}`,
          status: response.status,
          statusText: response.statusText,
          data
        })
      }

      return {
        data,
        status: response.status,
        statusText: response.statusText,
        headers
      }
    } catch (error) {
      if (error instanceof ApiError) {
        throw error
      }

      // Network or other errors
      if (error instanceof Error) {
        throw new ApiError({
          message: error.message || 'Network error occurred',
          status: 0,
          statusText: 'NetworkError'
        })
      }

      throw new ApiError({
        message: 'Unknown error occurred',
        status: 0,
        statusText: 'UnknownError'
      })
    }
  }

  // HTTP methods
  async get<T>(endpoint: string, params?: Record<string, string | number | boolean>): Promise<ApiResponse<T>> {
    let url = endpoint
    
    if (params) {
      const searchParams = new URLSearchParams()
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          searchParams.append(key, String(value))
        }
      })
      const queryString = searchParams.toString()
      if (queryString) {
        url += `?${queryString}`
      }
    }

    return this.request<T>(url, { method: 'GET' })
  }

  async post<T>(endpoint: string, data?: Record<string, unknown> | unknown[]): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined
    })
  }

  async put<T>(endpoint: string, data?: Record<string, unknown> | unknown[]): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, {
      method: 'PUT',
      body: data ? JSON.stringify(data) : undefined
    })
  }

  async patch<T>(endpoint: string, data?: Record<string, unknown> | unknown[]): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, {
      method: 'PATCH',
      body: data ? JSON.stringify(data) : undefined
    })
  }

  async delete<T>(endpoint: string): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, { method: 'DELETE' })
  }

  // Utility methods
  getBaseURL(): string {
    return this.baseURL
  }

  setBaseURL(url: string): void {
    this.baseURL = url
  }

  setDefaultHeader(key: string, value: string): void {
    this.defaultHeaders[key] = value
  }

  removeDefaultHeader(key: string): void {
    delete this.defaultHeaders[key]
  }
}

// Create default API client instance
export const apiClient = new ApiClient()

// Error handling utilities
export function isApiError(error: unknown): error is ApiError {
  return error && typeof error.message === 'string'
}

export function getErrorMessage(error: unknown): string {
  if (isApiError(error)) {
    return error.message
  }
  
  if (error instanceof Error) {
    return error.message
  }
  
  return 'An unknown error occurred'
}

// Response type utilities
export function isSuccessResponse<T>(response: ApiResponse<T>): boolean {
  return response.status >= 200 && response.status < 300
}

// Environment configuration
export function getApiBaseURL(): string {
  // Check for Vite environment variable first
  const envApiUrl = import.meta.env?.VITE_API_BASE_URL
  if (envApiUrl) {
    return envApiUrl
  }
  
  if (typeof window !== 'undefined') {
    // Browser environment
    const hostname = window.location.hostname
    const isDevelopment = hostname === 'localhost' || hostname === '127.0.0.1'
    
    if (isDevelopment) {
      return 'http://localhost:9090'
    }
    
    // Production - same origin
    return `${window.location.protocol}//${window.location.host}`
  }
  
  // Server-side rendering or Node.js environment
  // @ts-expect-error - process may not be available in browser
  return (typeof process !== 'undefined' && process.env?.API_BASE_URL) || 'http://localhost:9090'
}

// Initialize API client with environment-specific base URL
apiClient.setBaseURL(getApiBaseURL())

export default apiClient