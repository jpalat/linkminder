document.addEventListener('DOMContentLoaded', async () => {
  const apiUrlInput = document.getElementById('apiUrl');
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
  
  // Load saved settings
  try {
    const result = await chrome.storage.sync.get(['apiUrl']);
    if (result.apiUrl) {
      apiUrlInput.value = result.apiUrl;
    } else {
      // Default URL
      apiUrlInput.value = 'http://localhost:9090';
    }
  } catch (error) {
    console.error('Error loading settings:', error);
    apiUrlInput.value = 'http://localhost:9090';
  }
  
  // Save settings
  saveBtn.addEventListener('click', async () => {
    const apiUrl = apiUrlInput.value.trim();
    
    if (!apiUrl) {
      showStatus('API URL is required', true);
      return;
    }
    
    try {
      // Validate URL format
      new URL(apiUrl);
      
      // Remove trailing slash
      const cleanUrl = apiUrl.replace(/\/$/, '');
      
      await chrome.storage.sync.set({ apiUrl: cleanUrl });
      showStatus('Settings saved successfully!');
    } catch (error) {
      if (error instanceof TypeError) {
        showStatus('Please enter a valid URL', true);
      } else {
        showStatus('Error saving settings', true);
      }
    }
  });
});