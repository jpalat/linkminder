import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import { useNotifications } from '@/composables/useNotifications'
import AppToast from '@/components/ui/AppToast.vue'
import ToastContainer from '@/components/ui/ToastContainer.vue'

// Mock timers for testing
vi.useFakeTimers()

describe('Notification System', () => {
  beforeEach(() => {
    // Clear any existing notifications
    const { clearAll } = useNotifications()
    clearAll()
  })

  afterEach(() => {
    vi.runAllTimers()
    vi.clearAllTimers()
  })

  describe('useNotifications composable', () => {
    it('should add notifications correctly', () => {
      const { addNotification, notifications } = useNotifications()
      
      const id = addNotification('success', 'Test message', { title: 'Test Title' })
      
      expect(notifications.value).toHaveLength(1)
      expect(notifications.value[0]).toMatchObject({
        id,
        type: 'success',
        message: 'Test message',
        title: 'Test Title',
        dismissible: true
      })
    })

    it('should remove notifications correctly', () => {
      const { addNotification, removeNotification, notifications } = useNotifications()
      
      const id = addNotification('info', 'Test message')
      expect(notifications.value).toHaveLength(1)
      
      removeNotification(id)
      expect(notifications.value).toHaveLength(0)
    })

    it('should auto-remove non-persistent notifications', () => {
      const { addNotification, notifications } = useNotifications()
      
      addNotification('success', 'Test message', { duration: 1000 })
      expect(notifications.value).toHaveLength(1)
      
      // Fast-forward time
      vi.advanceTimersByTime(10000) // Use max duration fallback
      
      expect(notifications.value).toHaveLength(0)
    })

    it('should not auto-remove persistent notifications', () => {
      const { addNotification, notifications } = useNotifications()
      
      addNotification('error', 'Persistent message', { persistent: true, duration: 1000 })
      expect(notifications.value).toHaveLength(1)
      
      vi.advanceTimersByTime(10000)
      
      expect(notifications.value).toHaveLength(1)
    })

    it('should clear all notifications', () => {
      const { addNotification, clearAll, notifications } = useNotifications()
      
      addNotification('success', 'Message 1')
      addNotification('error', 'Message 2')
      addNotification('warning', 'Message 3')
      
      expect(notifications.value).toHaveLength(3)
      
      clearAll()
      expect(notifications.value).toHaveLength(0)
    })
  })

  describe('Convenience notification methods', () => {
    it('should create success notifications', () => {
      const { success, notifications } = useNotifications()
      
      success('Operation successful')
      
      expect(notifications.value[0]).toMatchObject({
        type: 'success',
        message: 'Operation successful',
        duration: 5000
      })
    })

    it('should create error notifications with longer duration', () => {
      const { error, notifications } = useNotifications()
      
      error('Something went wrong')
      
      expect(notifications.value[0]).toMatchObject({
        type: 'error',
        message: 'Something went wrong',
        duration: 7000
      })
    })

    it('should create warning notifications', () => {
      const { warning, notifications } = useNotifications()
      
      warning('This is a warning')
      
      expect(notifications.value[0]).toMatchObject({
        type: 'warning',
        message: 'This is a warning'
      })
    })

    it('should create info notifications', () => {
      const { info, notifications } = useNotifications()
      
      info('Information message')
      
      expect(notifications.value[0]).toMatchObject({
        type: 'info',
        message: 'Information message'
      })
    })
  })

  describe('Bookmark operation notifications', () => {
    it('should create bookmark created notification', () => {
      const { bookmarkCreated, notifications } = useNotifications()
      
      bookmarkCreated('Test Article')
      
      expect(notifications.value[0]).toMatchObject({
        type: 'success',
        message: '"Test Article" has been saved',
        title: 'Bookmark Created'
      })
    })

    it('should create bookmark updated notification', () => {
      const { bookmarkUpdated, notifications } = useNotifications()
      
      bookmarkUpdated('Updated Article')
      
      expect(notifications.value[0]).toMatchObject({
        type: 'success',
        message: '"Updated Article" has been updated',
        title: 'Bookmark Updated'
      })
    })

    it('should create bookmark deleted notification', () => {
      const { bookmarkDeleted, notifications } = useNotifications()
      
      bookmarkDeleted('Deleted Article')
      
      expect(notifications.value[0]).toMatchObject({
        type: 'success',
        message: '"Deleted Article" has been deleted',
        title: 'Bookmark Deleted'
      })
    })

    it('should create bookmark moved notification', () => {
      const { bookmarkMoved, notifications } = useNotifications()
      
      bookmarkMoved('working', 'Project Article')
      
      expect(notifications.value[0]).toMatchObject({
        type: 'success',
        message: '"Project Article" has been moved to working',
        title: 'Bookmark Updated'
      })
    })

    it('should create bulk operation notification', () => {
      const { bulkOperation, notifications } = useNotifications()
      
      bulkOperation(5, 'archived')
      
      expect(notifications.value[0]).toMatchObject({
        type: 'success',
        message: '5 bookmarks archived',
        title: 'Bulk Operation Complete'
      })
    })
  })

  describe('Error handling notifications', () => {
    it('should create API error notification', () => {
      const { apiError, notifications } = useNotifications()
      
      const error = new Error('Network timeout')
      apiError('create bookmark', error)
      
      expect(notifications.value[0]).toMatchObject({
        type: 'error',
        message: 'Failed to create bookmark: Network timeout',
        title: 'Operation Failed',
        duration: 8000
      })
    })

    it('should create network error notification', () => {
      const { networkError, notifications } = useNotifications()
      
      networkError()
      
      expect(notifications.value[0]).toMatchObject({
        type: 'error',
        title: 'Network Error',
        duration: 10000
      })
    })

    it('should create validation error notification', () => {
      const { validationError, notifications } = useNotifications()
      
      validationError('Please fill in all required fields')
      
      expect(notifications.value[0]).toMatchObject({
        type: 'warning',
        message: 'Please fill in all required fields',
        title: 'Validation Error',
        duration: 6000
      })
    })
  })

  describe('Project operation notifications', () => {
    it('should create project created notification', () => {
      const { projectCreated, notifications } = useNotifications()
      
      projectCreated('New Project')
      
      expect(notifications.value[0]).toMatchObject({
        type: 'success',
        message: 'Project "New Project" has been created',
        title: 'Project Created'
      })
    })

    it('should create project updated notification', () => {
      const { projectUpdated, notifications } = useNotifications()
      
      projectUpdated('Updated Project')
      
      expect(notifications.value[0]).toMatchObject({
        type: 'success',
        message: 'Project "Updated Project" has been updated',
        title: 'Project Updated'
      })
    })

    it('should create project deleted notification', () => {
      const { projectDeleted, notifications } = useNotifications()
      
      projectDeleted('Deleted Project')
      
      expect(notifications.value[0]).toMatchObject({
        type: 'success',
        message: 'Project "Deleted Project" has been deleted',
        title: 'Project Deleted'
      })
    })
  })
})

describe('AppToast Component', () => {
  it('should render correctly with basic props', () => {
    const wrapper = mount(AppToast, {
      props: {
        id: 'test-1',
        type: 'success',
        message: 'Test message'
      }
    })

    expect(wrapper.find('.toast').exists()).toBe(true)
    expect(wrapper.find('.toast--success').exists()).toBe(true)
    expect(wrapper.find('.toast__message').text()).toBe('Test message')
    expect(wrapper.find('span').text()).toBe('✅')
  })

  it('should render title when provided', () => {
    const wrapper = mount(AppToast, {
      props: {
        type: 'info',
        title: 'Test Title',
        message: 'Test message'
      }
    })

    expect(wrapper.find('.toast__title').text()).toBe('Test Title')
    expect(wrapper.find('.toast__message').text()).toBe('Test message')
  })

  it('should render close button when dismissible', () => {
    const wrapper = mount(AppToast, {
      props: {
        type: 'warning',
        message: 'Test message',
        dismissible: true
      }
    })

    expect(wrapper.find('.toast__close').exists()).toBe(true)
  })

  it('should not render close button when not dismissible', () => {
    const wrapper = mount(AppToast, {
      props: {
        type: 'warning',
        message: 'Test message',
        dismissible: false
      }
    })

    expect(wrapper.find('.toast__close').exists()).toBe(false)
  })

  it('should emit close event when close button is clicked', async () => {
    const wrapper = mount(AppToast, {
      props: {
        id: 'test-close',
        type: 'error',
        message: 'Test message',
        dismissible: true
      }
    })

    await wrapper.find('.toast__close').trigger('click')
    
    // Wait for transition with vi timers
    vi.advanceTimersByTime(300)
    await nextTick()
    
    expect(wrapper.emitted('close')).toBeTruthy()
    expect(wrapper.emitted('close')![0]).toEqual(['test-close'])
  })

  it('should auto-close after duration', async () => {
    const wrapper = mount(AppToast, {
      props: {
        id: 'test-auto-close',
        type: 'success',
        message: 'Test message',
        duration: 1000
      }
    })

    // Fast-forward time
    vi.advanceTimersByTime(1000)
    
    // Wait for transition with vi timers
    vi.advanceTimersByTime(300)
    await nextTick()
    
    expect(wrapper.emitted('close')).toBeTruthy()
  })

  it('should render correct icons for different types', () => {
    const types = [
      { type: 'success', icon: '✅' },
      { type: 'error', icon: '❌' },
      { type: 'warning', icon: '⚠️' },
      { type: 'info', icon: 'ℹ️' }
    ]

    types.forEach(({ type, icon }) => {
      const wrapper = mount(AppToast, {
        props: {
          type: type as any,
          message: 'Test message'
        }
      })

      expect(wrapper.find('.toast__icon span').text()).toBe(icon)
    })
  })
})

describe('ToastContainer Component', () => {
  it('should render notifications from the global state', async () => {
    const { addNotification } = useNotifications()
    
    // Add some test notifications
    addNotification('success', 'Success message')
    addNotification('error', 'Error message')
    
    const wrapper = mount(ToastContainer)
    await nextTick()

    expect(wrapper.findAllComponents(AppToast)).toHaveLength(2)
  })

  it('should not render container when no notifications', () => {
    const { clearAll } = useNotifications()
    clearAll()
    
    const wrapper = mount(ToastContainer)

    expect(wrapper.find('.toast-container').exists()).toBe(false)
  })

  it('should remove notification when close event is emitted', async () => {
    const { addNotification, notifications } = useNotifications()
    
    const id = addNotification('info', 'Test message')
    
    const wrapper = mount(ToastContainer)
    await nextTick()

    expect(notifications.value).toHaveLength(1)
    
    // Emit close event from the toast
    const toast = wrapper.findComponent(AppToast)
    await toast.vm.$emit('close', id)
    
    expect(notifications.value).toHaveLength(0)
  })
})

describe('Share Destination Integration', () => {
  beforeEach(() => {
    // Clear notifications before each test
    const { clearAll } = useNotifications()
    clearAll()
  })

  it('should show notification when bookmark is shared with recipient', () => {
    const { bookmarkMoved, notifications } = useNotifications()
    
    bookmarkMoved('share', 'React Article')
    
    expect(notifications.value[0]).toMatchObject({
      type: 'success',
      message: '"React Article" has been marked for sharing',
      title: 'Bookmark Updated'
    })
  })

  it('should show bulk operation notification for multiple shares', () => {
    const { bulkOperation, notifications } = useNotifications()
    
    bulkOperation(3, 'marked for sharing')
    
    expect(notifications.value[0]).toMatchObject({
      type: 'success',
      message: '3 bookmarks marked for sharing',
      title: 'Bulk Operation Complete'
    })
  })
})