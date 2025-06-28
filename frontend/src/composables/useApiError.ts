import { ref, computed, readonly, type Ref } from 'vue'
import { getErrorMessage, isApiError, type ApiError } from '@/services/api'

export interface UseApiErrorReturn {
  error: Readonly<Ref<string | null>>
  isError: Readonly<Ref<boolean>>
  setError: (error: string | Error | ApiError | null) => void
  clearError: () => void
  handleError: (error: unknown) => void
}

/**
 * Composable for handling API errors consistently across components
 */
export function useApiError(): UseApiErrorReturn {
  const error = ref<string | null>(null)

  const isError = computed(() => !!error.value)

  const setError = (err: string | Error | ApiError | null) => {
    if (err === null) {
      error.value = null
      return
    }

    if (typeof err === 'string') {
      error.value = err
      return
    }

    error.value = getErrorMessage(err)
  }

  const clearError = () => {
    error.value = null
  }

  const handleError = (err: unknown) => {
    console.error('API Error:', err)
    if (typeof err === 'string') {
      setError(err)
    } else if (err instanceof Error) {
      setError(err)
    } else {
      setError('An unknown error occurred')
    }
  }

  return {
    error: readonly(error),
    isError: readonly(isError),
    setError,
    clearError,
    handleError
  }
}

/**
 * Convert API errors to user-friendly messages
 */
export function getErrorDisplayMessage(error: unknown): string {
  if (isApiError(error)) {
    switch (error.status) {
      case 400:
        return 'Invalid request. Please check your input and try again.'
      case 401:
        return 'You are not authorized to perform this action.'
      case 403:
        return 'Access denied. You do not have permission for this action.'
      case 404:
        return 'The requested resource was not found.'
      case 409:
        return 'A conflict occurred. The resource may already exist.'
      case 422:
        return 'The data provided is invalid or incomplete.'
      case 429:
        return 'Too many requests. Please wait a moment and try again.'
      case 500:
        return 'A server error occurred. Please try again later.'
      case 502:
      case 503:
      case 504:
        return 'The service is temporarily unavailable. Please try again later.'
      default:
        return error.message || 'An unexpected error occurred.'
    }
  }

  if (error instanceof Error) {
    if (error.message.includes('fetch')) {
      return 'Unable to connect to the server. Please check your internet connection.'
    }
    return error.message
  }

  return 'An unknown error occurred. Please try again.'
}

/**
 * Check if an error is a network error
 */
export function isNetworkError(error: unknown): boolean {
  if (isApiError(error)) {
    return error.status === 0 || error.statusText === 'NetworkError'
  }
  
  if (error instanceof Error) {
    return error.message.includes('fetch') || 
           error.message.includes('network') ||
           error.message.includes('Network')
  }
  
  return false
}

/**
 * Check if an error is a server error (5xx)
 */
export function isServerError(error: unknown): boolean {
  if (isApiError(error)) {
    return error.status !== undefined && error.status >= 500 && error.status < 600
  }
  return false
}

/**
 * Check if an error is a client error (4xx)
 */
export function isClientError(error: unknown): boolean {
  if (isApiError(error)) {
    return error.status !== undefined && error.status >= 400 && error.status < 500
  }
  return false
}

export default useApiError