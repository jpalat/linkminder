import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import ShareGroups from '@/components/share/ShareGroups.vue'
import type { ShareGroup, Bookmark } from '@/types'

// Mock document.execCommand for clipboard testing
const mockExecCommand = vi.fn().mockReturnValue(true)
Object.defineProperty(document, 'execCommand', {
  value: mockExecCommand,
  writable: true
})

// Mock notifications
const mockSuccess = vi.fn()
const mockError = vi.fn()
vi.mock('@/composables/useNotifications', () => ({
  useNotifications: () => ({
    success: mockSuccess,
    error: mockError
  })
}))

describe('Share Copy Functionality', () => {
  const mockBookmarks: Bookmark[] = [
    {
      id: '1',
      url: 'https://example.com/article1',
      title: 'Test Article 1',
      description: 'This is a test article about React',
      action: 'share',
      shareTo: 'Custom Recipient',
      timestamp: '2024-01-15T10:30:00Z',
      domain: 'example.com',
      age: '2h'
    },
    {
      id: '2',
      url: 'https://test.com/guide',
      title: 'Vue.js Guide',
      description: 'Complete guide to Vue.js development',
      action: 'share',
      shareTo: 'Custom Recipient',
      timestamp: '2024-01-15T08:15:00Z',
      domain: 'test.com',
      age: '4h'
    }
  ]

  const mockShareGroups: ShareGroup[] = [
    {
      destination: 'Custom Recipient',
      items: mockBookmarks,
      icon: '📤',
      color: '#6b7280'
    }
  ]

  beforeEach(() => {
    vi.clearAllMocks()
    mockExecCommand.mockReturnValue(true)
  })

  it('should render copy buttons for each format', () => {
    const wrapper = mount(ShareGroups, {
      props: {
        groups: mockShareGroups
      }
    })

    // Check that format buttons exist for individual items
    const formatButtons = wrapper.findAll('.format-btn')
    expect(formatButtons.length).toBeGreaterThan(0)

    // Check for the three required formats
    const buttonTexts = formatButtons.map(btn => btn.text())
    expect(buttonTexts.some(text => text.includes('Rich Text'))).toBe(true)
    expect(buttonTexts.some(text => text.includes('Markdown'))).toBe(true)
    expect(buttonTexts.some(text => text.includes('Plain Text'))).toBe(true)
  })

  it('should copy item in rich-text format', async () => {
    const wrapper = mount(ShareGroups, {
      props: {
        groups: mockShareGroups
      }
    })

    const component = wrapper.vm as any
    await component.copyItemFormat(mockBookmarks[0], 'rich-text')

    expect(mockExecCommand).toHaveBeenCalledWith('copy')
    expect(mockSuccess).toHaveBeenCalledWith(
      'Copied "Test Article 1" as rich-text format',
      { title: 'Bookmark Copied' }
    )
  })

  it('should copy item in markdown format', async () => {
    const wrapper = mount(ShareGroups, {
      props: {
        groups: mockShareGroups
      }
    })

    const component = wrapper.vm as any
    await component.copyItemFormat(mockBookmarks[0], 'markdown')

    expect(mockExecCommand).toHaveBeenCalledWith('copy')
    expect(mockSuccess).toHaveBeenCalledWith(
      'Copied "Test Article 1" as markdown format',
      { title: 'Bookmark Copied' }
    )
  })

  it('should copy item in plain text format', async () => {
    const wrapper = mount(ShareGroups, {
      props: {
        groups: mockShareGroups
      }
    })

    const component = wrapper.vm as any
    await component.copyItemFormat(mockBookmarks[0], 'plain')

    expect(mockExecCommand).toHaveBeenCalledWith('copy')
    expect(mockSuccess).toHaveBeenCalledWith(
      'Copied "Test Article 1" as plain format',
      { title: 'Bookmark Copied' }
    )
  })

  it('should copy group items in different formats', async () => {
    const wrapper = mount(ShareGroups, {
      props: {
        groups: mockShareGroups
      }
    })

    const component = wrapper.vm as any

    // Test rich-text group format
    await component.copyGroupItems(mockShareGroups[0], 'rich-text')
    expect(mockExecCommand).toHaveBeenCalledWith('copy')

    // Test markdown group format
    await component.copyGroupItems(mockShareGroups[0], 'markdown')
    expect(mockExecCommand).toHaveBeenCalledWith('copy')

    // Test plain text group format
    await component.copyGroupItems(mockShareGroups[0], 'plain')
    expect(mockExecCommand).toHaveBeenCalledWith('copy')
  })

  it('should handle items without descriptions', async () => {
    const bookmarkWithoutDescription: Bookmark = {
      id: '3',
      url: 'https://minimal.com/article',
      title: 'Minimal Article',
      action: 'share',
      shareTo: 'Another Recipient',
      timestamp: '2024-01-15T12:00:00Z',
      domain: 'minimal.com',
      age: '1h'
    }

    const wrapper = mount(ShareGroups, {
      props: {
        groups: [
          {
            destination: 'Another Recipient',
            items: [bookmarkWithoutDescription],
            icon: '📤',
            color: '#6b7280'
          }
        ]
      }
    })

    const component = wrapper.vm as any

    // Test rich-text format without description
    await component.copyItemFormat(bookmarkWithoutDescription, 'rich-text')
    expect(mockExecCommand).toHaveBeenCalledWith('copy')

    // Test markdown format without description
    await component.copyItemFormat(bookmarkWithoutDescription, 'markdown')
    expect(mockExecCommand).toHaveBeenCalledWith('copy')
  })

  it('should show correct format icons and labels', () => {
    const wrapper = mount(ShareGroups, {
      props: {
        groups: mockShareGroups
      }
    })

    // Check individual item format buttons
    const formatButtons = wrapper.findAll('.format-btn')
    const buttonTexts = formatButtons.map(btn => btn.text())

    expect(buttonTexts.some(text => text.includes('📄') && text.includes('Rich Text'))).toBe(true)
    expect(buttonTexts.some(text => text.includes('📝') && text.includes('Markdown'))).toBe(true)
    expect(buttonTexts.some(text => text.includes('📃') && text.includes('Plain Text'))).toBe(true)
  })

  it('should handle copy errors gracefully', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    mockExecCommand.mockReturnValue(false) // Simulate copy failure

    const wrapper = mount(ShareGroups, {
      props: {
        groups: mockShareGroups
      }
    })

    const component = wrapper.vm as any
    await component.copyItemFormat(mockBookmarks[0], 'markdown')

    expect(mockError).toHaveBeenCalledWith(
      'Failed to copy to clipboard. Please try again.',
      { title: 'Copy Failed' }
    )
    expect(mockSuccess).not.toHaveBeenCalled()

    consoleSpy.mockRestore()
  })
})