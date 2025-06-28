import { apiClient } from '@/services/api'

/**
 * Utility to check and debug server connection
 */
export class ServerStatus {
  static getCurrentServerURL(): string {
    return apiClient.getBaseURL() || 'Unknown'
  }

  static async checkServerHealth(): Promise<{
    isConnected: boolean
    serverURL: string
    responseTime?: number
    error?: string
  }> {
    const serverURL = this.getCurrentServerURL()
    const startTime = Date.now()
    
    try {
      // Try to fetch topics as a health check
      const response = await fetch(`${serverURL}/topics`, {
        method: 'GET',
        headers: { 'Accept': 'application/json' }
      })
      
      const responseTime = Date.now() - startTime
      
      if (response.ok) {
        return {
          isConnected: true,
          serverURL,
          responseTime
        }
      } else {
        return {
          isConnected: false,
          serverURL,
          responseTime,
          error: `Server responded with ${response.status}: ${response.statusText}`
        }
      }
    } catch (error) {
      return {
        isConnected: false,
        serverURL,
        responseTime: Date.now() - startTime,
        error: error instanceof Error ? error.message : 'Unknown error'
      }
    }
  }

  static async listAvailableEndpoints(): Promise<{
    endpoint: string
    method: string
    status: 'working' | 'error' | 'unknown'
    responseTime?: number
  }[]> {
    const serverURL = this.getCurrentServerURL()
    const endpoints = [
      { endpoint: '/topics', method: 'GET' },
      { endpoint: '/api/stats/summary', method: 'GET' },
      { endpoint: '/api/bookmarks/triage', method: 'GET' },
      { endpoint: '/api/projects', method: 'GET' },
      { endpoint: '/bookmark', method: 'POST' }
    ]

    const results = await Promise.all(
      endpoints.map(async ({ endpoint, method }) => {
        const startTime = Date.now()
        try {
          const response = await fetch(`${serverURL}${endpoint}`, {
            method: method === 'GET' ? 'GET' : 'HEAD', // Use HEAD for non-GET to avoid side effects
            headers: { 'Accept': 'application/json' }
          })
          
          return {
            endpoint,
            method,
            status: response.ok ? 'working' as const : 'error' as const,
            responseTime: Date.now() - startTime
          }
        } catch {
          return {
            endpoint,
            method,
            status: 'error' as const,
            responseTime: Date.now() - startTime
          }
        }
      })
    )

    return results
  }

  static setCustomServerURL(url: string): void {
    console.log(`🔧 Switching API server from ${this.getCurrentServerURL()} to ${url}`)
    apiClient.setBaseURL(url)
  }

  static resetToDefault(): void {
    const defaultURL = 'http://localhost:9090'
    console.log(`🔄 Resetting API server to default: ${defaultURL}`)
    apiClient.setBaseURL(defaultURL)
  }
}

// Global functions for browser console debugging
declare global {
  interface Window {
    serverStatus?: typeof ServerStatus
  }
}

if (typeof window !== 'undefined') {
  window.serverStatus = ServerStatus
  console.log('🛠️  Server debugging available via: window.serverStatus')
}