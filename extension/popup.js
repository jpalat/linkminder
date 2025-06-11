document.addEventListener('DOMContentLoaded', async () => {
  const urlInput = document.getElementById('url');
  const titleInput = document.getElementById('title');
  const descriptionInput = document.getElementById('description');
  const actionSelect = document.getElementById('action');
  const shareField = document.getElementById('shareField');
  const shareToInput = document.getElementById('shareTo');
  const topicField = document.getElementById('topicField');
  const topicInput = document.getElementById('topic');
  const topicSuggestions = document.getElementById('topicSuggestions');
  const saveBtn = document.getElementById('saveBtn');
  const saveCloseBtn = document.getElementById('saveCloseBtn');
  const statusDiv = document.getElementById('status');
  
  function showStatus(message, isError = false) {
    statusDiv.textContent = message;
    statusDiv.className = `status ${isError ? 'error' : 'success'}`;
    statusDiv.style.display = 'block';
    
    setTimeout(() => {
      statusDiv.style.display = 'none';
    }, 3000);
  }
  
  
  // Fetch topics from server
  async function loadTopics() {
    try {
      const response = await chrome.runtime.sendMessage({
        action: 'getTopics'
      });
      
      if (response.success && response.topics) {
        topicSuggestions.innerHTML = '';
        response.topics.forEach(topic => {
          const option = document.createElement('option');
          option.value = topic;
          topicSuggestions.appendChild(option);
        });
      }
    } catch (error) {
      console.error('Error loading topics:', error);
    }
  }
  
  // Handle action selection changes
  function updateConditionalFields() {
    const selectedAction = actionSelect.value;
    
    shareField.classList.toggle('show', selectedAction === 'share');
    topicField.classList.toggle('show', selectedAction === 'working');
  }
  
  // Helper function to get page data with fallback
  async function getPageDataSafely(tab) {
    // Check if this is a restricted page
    if (tab.url.startsWith('chrome://') || tab.url.startsWith('chrome-extension://') || 
        tab.url.startsWith('edge://') || tab.url.startsWith('about:')) {
      return {
        url: tab.url,
        title: tab.title || 'Unknown Page',
        description: '',
        content: ''
      };
    }
    
    try {
      // First try to send message to existing content script
      const response = await chrome.tabs.sendMessage(tab.id, { action: 'getPageData' });
      return response;
    } catch (error) {
      // Content script not available, try to inject it
      try {
        await chrome.scripting.executeScript({
          target: { tabId: tab.id },
          files: ['content.js']
        });
        
        // Wait a bit for the script to load
        await new Promise(resolve => setTimeout(resolve, 100));
        
        // Try again
        const response = await chrome.tabs.sendMessage(tab.id, { action: 'getPageData' });
        return response;
      } catch (injectError) {
        // Fallback to basic tab info
        console.warn('Could not inject content script:', injectError);
        return {
          url: tab.url,
          title: tab.title || 'Unknown Page',
          description: '',
          content: ''
        };
      }
    }
  }
  
  // Initialize
  await loadTopics();
  
  // Set up event listeners
  actionSelect.addEventListener('change', updateConditionalFields);
  
  // Initial update of conditional fields
  updateConditionalFields();
  
  try {
    const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
    const pageData = await getPageDataSafely(tab);
    
    urlInput.value = pageData.url;
    titleInput.value = pageData.title;
    descriptionInput.value = pageData.description;
    
    if (!pageData.content) {
      showStatus('Page content not available - saved with basic info', false);
    }
  } catch (error) {
    console.error('Error getting page data:', error);
    showStatus('Error loading page data', true);
  }
  
  // Shared save function
  async function saveBookmark(shouldCloseTab = false) {
    const url = urlInput.value.trim();
    const title = titleInput.value.trim();
    const description = descriptionInput.value.trim();
    const action = actionSelect.value;
    const shareTo = shareToInput.value.trim();
    const topic = topicInput.value.trim();
    
    if (!url || !title) {
      showStatus('URL and title are required', true);
      return;
    }
    
    // Validate action-specific fields
    if (action === 'share' && !shareTo) {
      showStatus('Please specify who to share with', true);
      return;
    }
    
    if (action === 'working' && !topic) {
      showStatus('Please specify a topic', true);
      return;
    }
    
    // Disable both buttons during save
    saveBtn.disabled = true;
    saveCloseBtn.disabled = true;
    
    const originalSaveText = saveBtn.textContent;
    const originalSaveCloseText = saveCloseBtn.textContent;
    
    saveBtn.textContent = 'Saving...';
    saveCloseBtn.textContent = 'Saving...';
    
    try {
      // Get the full page data including content
      const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
      const pageData = await getPageDataSafely(tab);
      
      const bookmarkData = { 
        url, 
        title, 
        description, 
        content: pageData.content || '',
        action,
        shareTo: action === 'share' ? shareTo : '',
        topic: action === 'working' ? topic : ''
      };
      
      const response = await chrome.runtime.sendMessage({
        action: 'saveBookmark',
        data: bookmarkData
      });
      
      if (response.success) {
        showStatus('Bookmark saved successfully!');
        
        // Close tab if requested
        if (shouldCloseTab) {
          setTimeout(async () => {
            // Close the current browser tab (where the bookmark was saved from)
            const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
            if (tab) {
              chrome.tabs.remove(tab.id);
            }
          }, 1500);
        }
      } else {
        showStatus(`Error: ${response.error}`, true);
      }
    } catch (error) {
      showStatus('Failed to save bookmark', true);
    } finally {
      saveBtn.disabled = false;
      saveCloseBtn.disabled = false;
      saveBtn.textContent = originalSaveText;
      saveCloseBtn.textContent = originalSaveCloseText;
    }
  }

  // Event listeners for save buttons
  saveBtn.addEventListener('click', () => saveBookmark(false));
  saveCloseBtn.addEventListener('click', () => saveBookmark(true));
});