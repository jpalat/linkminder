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
  const autoCloseCheckbox = document.getElementById('autoClose');
  const saveBtn = document.getElementById('saveBtn');
  const statusDiv = document.getElementById('status');
  
  function showStatus(message, isError = false) {
    statusDiv.textContent = message;
    statusDiv.className = `status ${isError ? 'error' : 'success'}`;
    statusDiv.style.display = 'block';
    
    setTimeout(() => {
      statusDiv.style.display = 'none';
    }, 3000);
  }
  
  // Load persisted settings
  async function loadSettings() {
    const result = await chrome.storage.local.get('autoClose');
    autoCloseCheckbox.checked = result.autoClose ?? true;
  }
  
  // Save settings
  async function saveSettings() {
    await chrome.storage.local.set({ 
      autoClose: autoCloseCheckbox.checked 
    });
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
  
  // Initialize
  await loadSettings();
  await loadTopics();
  
  // Set up event listeners
  actionSelect.addEventListener('change', updateConditionalFields);
  autoCloseCheckbox.addEventListener('change', saveSettings);
  
  // Initial update of conditional fields
  updateConditionalFields();
  
  try {
    const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
    
    const response = await chrome.tabs.sendMessage(tab.id, { action: 'getPageData' });
    
    urlInput.value = response.url;
    titleInput.value = response.title;
    descriptionInput.value = response.description;
  } catch (error) {
    console.error('Error getting page data:', error);
    showStatus('Error loading page data', true);
  }
  
  saveBtn.addEventListener('click', async () => {
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
    
    saveBtn.disabled = true;
    saveBtn.textContent = 'Saving...';
    
    try {
      // Get the full page data including content
      const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
      const pageData = await chrome.tabs.sendMessage(tab.id, { action: 'getPageData' });
      
      const bookmarkData = { 
        url, 
        title, 
        description, 
        content: pageData.content,
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
        
        if (autoCloseCheckbox.checked) {
          setTimeout(() => {
            window.close();
          }, 1500);
        }
      } else {
        showStatus(`Error: ${response.error}`, true);
      }
    } catch (error) {
      showStatus('Failed to save bookmark', true);
    } finally {
      saveBtn.disabled = false;
      saveBtn.textContent = 'Save Bookmark';
    }
  });
});