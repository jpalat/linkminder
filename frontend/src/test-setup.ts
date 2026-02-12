import { vi } from 'vitest'

// Mock console.log, console.warn, console.error to reduce noise in test output
global.console = {
  ...console,
  log: vi.fn(),
  warn: vi.fn(),
  error: vi.fn()
}

// Mock IntersectionObserver for components that might use it
global.IntersectionObserver = vi.fn(() => ({
  observe: vi.fn(),
  disconnect: vi.fn(),
  unobserve: vi.fn()
}))

// Mock ResizeObserver for components that might use it
global.ResizeObserver = vi.fn(() => ({
  observe: vi.fn(),
  disconnect: vi.fn(),
  unobserve: vi.fn()
}))

// Mock window.matchMedia for responsive components
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation((query) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(), // deprecated
    removeListener: vi.fn(), // deprecated
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn()
  }))
})

// Mock fetch if needed
global.fetch = vi.fn()

// Mock localStorage for Pinia devtools and other components
const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
  key: vi.fn(),
  length: 0
}

Object.defineProperty(window, 'localStorage', {
  value: localStorageMock,
  writable: true
})

// Also set it globally for Node.js context
global.localStorage = localStorageMock as any