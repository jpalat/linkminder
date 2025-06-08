const API_BASE_URL = 'http://192.168.1.112:9090';

async function saveBookmark(bookmarkData) {
  try {
    const response = await fetch(`${API_BASE_URL}/bookmark`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(bookmarkData)
    });
    
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    
    const result = await response.json();
    return { success: true, data: result };
  } catch (error) {
    console.error('Error saving bookmark:', error);
    return { success: false, error: error.message };
  }
}

async function getTopics() {
  try {
    const response = await fetch(`${API_BASE_URL}/topics`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      }
    });
    
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    
    const result = await response.json();
    return { success: true, topics: result.topics || [] };
  } catch (error) {
    console.error('Error fetching topics:', error);
    return { success: false, error: error.message, topics: [] };
  }
}

chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.action === 'saveBookmark') {
    saveBookmark(request.data)
      .then(result => sendResponse(result))
      .catch(error => sendResponse({ success: false, error: error.message }));
    return true;
  } else if (request.action === 'getTopics') {
    getTopics()
      .then(result => sendResponse(result))
      .catch(error => sendResponse({ success: false, error: error.message, topics: [] }));
    return true;
  }
});