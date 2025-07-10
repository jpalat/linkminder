// Simple in-memory cache for bookmark lookups
const bookmarkCache = new Map();
const CACHE_DURATION = 5 * 60 * 1000; // 5 minutes

async function getApiBaseUrl() {
  try {
    const result = await chrome.storage.sync.get(['apiUrl']);
    return result.apiUrl || 'http://localhost:9090';
  } catch (error) {
    console.error('Error getting API URL from storage:', error);
    return 'http://localhost:9090';
  }
}

async function saveBookmark(bookmarkData) {
  try {
    const apiBaseUrl = await getApiBaseUrl();
    const response = await fetch(`${apiBaseUrl}/bookmark`, {
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
    
    // Clear cache for this URL since it was updated
    bookmarkCache.delete(bookmarkData.url);
    
    return { success: true, data: result };
  } catch (error) {
    console.error('Error saving bookmark:', error);
    return { success: false, error: error.message };
  }
}

async function getTopics() {
  try {
    const apiBaseUrl = await getApiBaseUrl();
    const response = await fetch(`${apiBaseUrl}/topics`, {
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

async function getBookmarkByUrl(url) {
  try {
    // Check cache first
    const cached = bookmarkCache.get(url);
    if (cached && Date.now() - cached.timestamp < CACHE_DURATION) {
      console.log('Returning cached bookmark for URL:', url);
      return cached.data;
    }
    
    const apiBaseUrl = await getApiBaseUrl();
    const encodedUrl = encodeURIComponent(url);
    const response = await fetch(`${apiBaseUrl}/api/bookmark/by-url?url=${encodedUrl}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      }
    });
    
    if (!response.ok) {
      if (response.status === 404) {
        const result = { success: true, found: false };
        // Cache negative result for a shorter duration
        bookmarkCache.set(url, {
          data: result,
          timestamp: Date.now()
        });
        return result;
      }
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    
    const result = await response.json();
    const responseData = { success: true, found: result.found, bookmark: result.bookmark };
    
    // Cache the result
    bookmarkCache.set(url, {
      data: responseData,
      timestamp: Date.now()
    });
    
    return responseData;
  } catch (error) {
    console.error('Error fetching bookmark by URL:', error);
    return { success: false, error: error.message, found: false };
  }
}

// Clean up old cache entries periodically
function cleanupCache() {
  const now = Date.now();
  for (const [url, cached] of bookmarkCache.entries()) {
    if (now - cached.timestamp > CACHE_DURATION) {
      bookmarkCache.delete(url);
    }
  }
}

// Run cleanup every 10 minutes
setInterval(cleanupCache, 10 * 60 * 1000);

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
  } else if (request.action === 'getBookmarkByUrl') {
    getBookmarkByUrl(request.url)
      .then(result => sendResponse(result))
      .catch(error => sendResponse({ success: false, error: error.message, found: false }));
    return true;
  }
});