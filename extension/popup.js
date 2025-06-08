document.addEventListener('DOMContentLoaded', async () => {
  const urlInput = document.getElementById('url');
  const titleInput = document.getElementById('title');
  const descriptionInput = document.getElementById('description');
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
    
    if (!url || !title) {
      showStatus('URL and title are required', true);
      return;
    }
    
    saveBtn.disabled = true;
    saveBtn.textContent = 'Saving...';
    
    try {
      // Get the full page data including content
      const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
      const pageData = await chrome.tabs.sendMessage(tab.id, { action: 'getPageData' });
      
      const response = await chrome.runtime.sendMessage({
        action: 'saveBookmark',
        data: { 
          url, 
          title, 
          description, 
          content: pageData.content 
        }
      });
      
      if (response.success) {
        showStatus('Bookmark saved successfully!');
        setTimeout(() => {
          window.close();
        }, 1500);
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