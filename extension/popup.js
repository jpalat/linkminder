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
  const tagsInput = document.getElementById('tagsInput');
  const tagsDisplay = document.getElementById('tagsDisplay');
  const propertyKeyInput = document.getElementById('propertyKey');
  const propertyValueInput = document.getElementById('propertyValue');
  const addPropertyBtn = document.getElementById('addProperty');
  const propertiesList = document.getElementById('propertiesList');
  const saveBtn = document.getElementById('saveBtn');
  const saveCloseBtn = document.getElementById('saveCloseBtn');
  const statusDiv = document.getElementById('status');
  
  let tags = [];
  let customProperties = {};
  
  function showStatus(message, isError = false) {
    statusDiv.textContent = message;
    statusDiv.className = `status ${isError ? 'error' : 'success'}`;
    statusDiv.style.display = 'block';
    
    setTimeout(() => {
      statusDiv.style.display = 'none';
    }, 3000);
  }
  
  // Function to show settings reminder
  function showSettingsReminder() {
    const settingsDiv = document.createElement('div');
    settingsDiv.className = 'status error';
    settingsDiv.innerHTML = 'API URL not configured. <a href="#" id="openSettings">Open Settings</a>';
    settingsDiv.style.display = 'block';
    statusDiv.parentNode.insertBefore(settingsDiv, statusDiv.nextSibling);
    
    document.getElementById('openSettings').addEventListener('click', (e) => {
      e.preventDefault();
      chrome.runtime.openOptionsPage();
    });
  }
  
  // Tags management functions
  function renderTags() {
    tagsDisplay.innerHTML = '';
    tags.forEach((tag, index) => {
      const tagElement = document.createElement('div');
      tagElement.className = 'tag';
      tagElement.innerHTML = `
        ${tag}
        <button class="tag-remove" data-index="${index}">×</button>
      `;
      tagsDisplay.appendChild(tagElement);
    });
  }
  
  function addTag(tagText) {
    const trimmedTag = tagText.trim();
    if (trimmedTag && !tags.includes(trimmedTag)) {
      tags.push(trimmedTag);
      renderTags();
    }
  }
  
  function removeTag(index) {
    tags.splice(index, 1);
    renderTags();
  }
  
  // Custom properties management functions
  function renderCustomProperties() {
    propertiesList.innerHTML = '';
    Object.entries(customProperties).forEach(([key, value]) => {
      const propertyElement = document.createElement('div');
      propertyElement.className = 'property-item';
      propertyElement.innerHTML = `
        <div class="property-display">
          <span class="property-key-display">${key}:</span>
          <span class="property-value-display">${value}</span>
        </div>
        <button class="remove-property-btn" data-key="${key}">×</button>
      `;
      propertiesList.appendChild(propertyElement);
    });
  }
  
  function addCustomProperty() {
    const key = propertyKeyInput.value.trim();
    const value = propertyValueInput.value.trim();
    
    if (key && value) {
      customProperties[key] = value;
      propertyKeyInput.value = '';
      propertyValueInput.value = '';
      renderCustomProperties();
    }
  }
  
  function removeCustomProperty(key) {
    delete customProperties[key];
    renderCustomProperties();
  }
  
  // Fetch topics from server
  async function loadTopics() {
    try {
      // Check if API URL is configured
      const result = await chrome.storage.sync.get(['apiUrl']);
      if (!result.apiUrl) {
        showSettingsReminder();
        return;
      }
      
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
        tab.url.startsWith('edge://') || tab.url.startsWith('about:') ||
        tab.url.startsWith('moz-extension://') || tab.url.startsWith('safari-extension://')) {
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
  
  // Check for existing bookmark
  async function checkExistingBookmark(url) {
    try {
      const response = await chrome.runtime.sendMessage({
        action: 'getBookmarkByUrl',
        url: url
      });
      
      if (response.success && response.found) {
        return response.bookmark;
      }
      return null;
    } catch (error) {
      console.error('Error checking existing bookmark:', error);
      return null;
    }
  }
  
  // Function to populate form with existing bookmark data
  function populateFormWithBookmark(bookmark) {
    // Set form fields
    titleInput.value = bookmark.title;
    descriptionInput.value = bookmark.description || '';
    actionSelect.value = bookmark.action || 'read-later';
    shareToInput.value = bookmark.shareTo || '';
    topicInput.value = bookmark.topic || '';
    
    // Set tags
    tags = bookmark.tags || [];
    renderTags();
    
    // Set custom properties
    customProperties = bookmark.customProperties || {};
    renderCustomProperties();
    
    // Update conditional fields
    updateConditionalFields();
    
    // Show status that bookmark exists
    showStatus('Found existing bookmark - fields populated', false);
    
    // Add visual indicator
    const existingIndicator = document.createElement('div');
    existingIndicator.className = 'status info';
    existingIndicator.innerHTML = `
      <strong>Previously saved:</strong> ${bookmark.age} ago
      <br>
      <small>Saving will update the existing bookmark</small>
    `;
    existingIndicator.style.display = 'block';
    statusDiv.parentNode.insertBefore(existingIndicator, statusDiv.nextSibling);
  }
  
  // Initialize
  await loadTopics();
  
  // Set up event listeners
  actionSelect.addEventListener('change', updateConditionalFields);
  
  // Tags input event listeners
  tagsInput.addEventListener('keydown', (e) => {
    if (e.key === 'Enter' || e.key === ',') {
      e.preventDefault();
      const tagText = tagsInput.value.trim();
      if (tagText) {
        // Handle comma-separated tags
        const newTags = tagText.split(',').map(t => t.trim()).filter(t => t);
        newTags.forEach(tag => addTag(tag));
        tagsInput.value = '';
      }
    }
  });
  
  tagsInput.addEventListener('blur', () => {
    const tagText = tagsInput.value.trim();
    if (tagText) {
      addTag(tagText);
      tagsInput.value = '';
    }
  });
  
  // Tag removal event listener (using event delegation)
  tagsDisplay.addEventListener('click', (e) => {
    if (e.target.classList.contains('tag-remove')) {
      const index = parseInt(e.target.dataset.index);
      removeTag(index);
    }
  });
  
  // Custom properties event listeners
  addPropertyBtn.addEventListener('click', addCustomProperty);
  
  propertyKeyInput.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      propertyValueInput.focus();
    }
  });
  
  propertyValueInput.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      addCustomProperty();
    }
  });
  
  // Property removal event listener (using event delegation)
  propertiesList.addEventListener('click', (e) => {
    if (e.target.classList.contains('remove-property-btn')) {
      const key = e.target.dataset.key;
      removeCustomProperty(key);
    }
  });
  
  // Initial update of conditional fields
  updateConditionalFields();
  
  try {
    const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
    const pageData = await getPageDataSafely(tab);
    
    urlInput.value = pageData.url;
    
    // Check for existing bookmark first
    const existingBookmark = await checkExistingBookmark(pageData.url);
    
    if (existingBookmark) {
      // Use existing bookmark data
      populateFormWithBookmark(existingBookmark);
    } else {
      // Use fresh page data
      titleInput.value = pageData.title;
      descriptionInput.value = pageData.description;
      
      if (!pageData.content) {
        showStatus('Page content not available - saved with basic info', false);
      }
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
        topic: action === 'working' ? topic : '',
        tags: tags,
        customProperties: customProperties
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