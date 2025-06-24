function getPageData() {
  const url = window.location.href;
  const title = document.title;
  
  let description = '';
  let content = '';
  let metadata = {};
  
  // Enhanced meta description extraction
  const metaDescription = document.querySelector('meta[name="description"]') || 
                         document.querySelector('meta[property="og:description"]') ||
                         document.querySelector('meta[name="twitter:description"]');
  
  if (metaDescription) {
    description = metaDescription.getAttribute('content');
  } else {
    // Fallback to first meaningful paragraph
    const paragraphs = document.querySelectorAll('p');
    for (const p of paragraphs) {
      const text = p.textContent.trim();
      if (text.length > 50) {
        description = text.substring(0, 200);
        break;
      }
    }
  }
  
  // Extract additional metadata
  const ogTitle = document.querySelector('meta[property="og:title"]');
  const author = document.querySelector('meta[name="author"]') || 
                document.querySelector('[rel="author"]') ||
                document.querySelector('.author, .byline');
  const publishDate = document.querySelector('meta[property="article:published_time"]') ||
                     document.querySelector('time[datetime]') ||
                     document.querySelector('.date, .published');
  
  if (ogTitle && ogTitle.getAttribute('content') !== title) {
    metadata.ogTitle = ogTitle.getAttribute('content');
  }
  if (author) {
    metadata.author = author.textContent || author.getAttribute('content');
  }
  if (publishDate) {
    metadata.publishDate = publishDate.getAttribute('datetime') || publishDate.textContent;
  }
  
  // Enhanced content extraction with better selectors
  const contentSelectors = [
    'main article',
    'article',
    'main',
    '[role="main"]',
    '.post-content',
    '.article-content',
    '.entry-content',
    '.content-body',
    '.post-body',
    '.article-body',
    '.content',
    '.main-content',
    '#content',
    '#main'
  ];
  
  let mainElement = null;
  for (const selector of contentSelectors) {
    mainElement = document.querySelector(selector);
    if (mainElement && mainElement.textContent.trim().length > 100) {
      break;
    }
  }
  
  // Fallback to body but exclude common non-content elements
  if (!mainElement || mainElement.textContent.trim().length < 100) {
    mainElement = document.body;
  }
  
  // Extract and clean text content
  if (mainElement) {
    // Clone element to avoid modifying original
    const clone = mainElement.cloneNode(true);
    
    // Remove unwanted elements
    const unwantedSelectors = [
      'nav', 'header', 'footer', 'aside',
      '.navigation', '.navbar', '.menu',
      '.sidebar', '.ads', '.advertisement',
      '.social-share', '.comments', '.related',
      'script', 'style', 'noscript'
    ];
    
    unwantedSelectors.forEach(selector => {
      const elements = clone.querySelectorAll(selector);
      elements.forEach(el => el.remove());
    });
    
    content = clone.innerText
      .replace(/\s+/g, ' ')
      .replace(/\n\s*\n/g, '\n')
      .trim()
      .substring(0, 3000); // Increased limit for better content
  }
  
  // Extract domain for categorization
  const domain = new URL(url).hostname.replace('www.', '');
  
  return {
    url,
    title,
    description,
    content,
    metadata,
    domain
  };
}

chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.action === 'getPageData') {
    sendResponse(getPageData());
  }
});