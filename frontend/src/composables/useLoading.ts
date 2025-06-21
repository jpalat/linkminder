import { ref, computed, readonly, type Ref } from 'vue'

export interface UseLoadingReturn {
  loading: Readonly<Ref<boolean>>
  isLoading: Readonly<Ref<boolean>>
  setLoading: (loading: boolean) => void
  startLoading: () => void
  stopLoading: () => void
  withLoading: <T>(fn: () => Promise<T>) => Promise<T>
}

/**
 * Composable for managing loading states
 */
export function useLoading(initialState: boolean = false): UseLoadingReturn {
  const loading = ref(initialState)

  const isLoading = computed(() => loading.value)

  const setLoading = (state: boolean) => {
    loading.value = state
  }

  const startLoading = () => {
    loading.value = true
  }

  const stopLoading = () => {
    loading.value = false
  }

  /**
   * Execute an async function while managing loading state
   */
  const withLoading = async <T>(fn: () => Promise<T>): Promise<T> => {
    startLoading()
    try {
      return await fn()
    } finally {
      stopLoading()
    }
  }

  return {
    loading: readonly(loading),
    isLoading: readonly(isLoading),
    setLoading,
    startLoading,
    stopLoading,
    withLoading
  }
}

/**
 * Composable for managing multiple loading states
 */
export function useMultipleLoading(): {
  loadingStates: Ref<Record<string, boolean>>
  isLoading: (key: string) => boolean
  isAnyLoading: Readonly<Ref<boolean>>
  setLoading: (key: string, loading: boolean) => void
  startLoading: (key: string) => void
  stopLoading: (key: string) => void
  withLoading: <T>(key: string, fn: () => Promise<T>) => Promise<T>
} {
  const loadingStates = ref<Record<string, boolean>>({})

  const isAnyLoading = computed(() => 
    Object.values(loadingStates.value).some(loading => loading)
  )

  const isLoading = (key: string): boolean => {
    return loadingStates.value[key] || false
  }

  const setLoading = (key: string, loading: boolean) => {
    loadingStates.value[key] = loading
  }

  const startLoading = (key: string) => {
    setLoading(key, true)
  }

  const stopLoading = (key: string) => {
    setLoading(key, false)
  }

  const withLoading = async <T>(key: string, fn: () => Promise<T>): Promise<T> => {
    startLoading(key)
    try {
      return await fn()
    } finally {
      stopLoading(key)
    }
  }

  return {
    loadingStates,
    isLoading,
    isAnyLoading: readonly(isAnyLoading),
    setLoading,
    startLoading,
    stopLoading,
    withLoading
  }
}

/**
 * Debounced loading state - prevents flashing for quick operations
 */
export function useDebouncedLoading(delay: number = 200): UseLoadingReturn {
  const loading = ref(false)
  const actualLoading = ref(false)
  let timeoutId: number | null = null

  const isLoading = computed(() => loading.value)

  const setLoading = (state: boolean) => {
    actualLoading.value = state

    if (timeoutId) {
      clearTimeout(timeoutId)
      timeoutId = null
    }

    if (state) {
      // Show loading immediately when starting
      loading.value = true
    } else {
      // Delay hiding loading to prevent flashing
      timeoutId = setTimeout(() => {
        if (!actualLoading.value) {
          loading.value = false
        }
      }, delay) as unknown as number
    }
  }

  const startLoading = () => setLoading(true)
  const stopLoading = () => setLoading(false)

  const withLoading = async <T>(fn: () => Promise<T>): Promise<T> => {
    startLoading()
    try {
      return await fn()
    } finally {
      stopLoading()
    }
  }

  return {
    loading: readonly(loading),
    isLoading: readonly(isLoading),
    setLoading,
    startLoading,
    stopLoading,
    withLoading
  }
}

export default useLoading