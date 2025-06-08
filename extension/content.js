function getPageData() {
  const url = window.location.href;
  const title = document.title;
  
  let description = '';
  let content = '';
  
  // Get meta description
  const metaDescription = document.querySelector('meta[name="description"]');
  if (metaDescription) {
    description = metaDescription.getAttribute('content');
  } else {
    const firstParagraph = document.querySelector('p');
    if (firstParagraph) {
      description = firstParagraph.textContent.trim().substring(0, 200);
    }
  }
  
  // Extract main content
  const contentSelectors = [
    'main',
    'article', 
    '[role="main"]',
    '.content',
    '.main-content',
    '#content',
    '#main'
  ];
  
  let mainElement = null;
  for (const selector of contentSelectors) {
    mainElement = document.querySelector(selector);
    if (mainElement) break;
  }
  
  if (!mainElement) {
    mainElement = document.body;
  }
  
  // Extract text content, cleaning up whitespace
  if (mainElement) {
    content = mainElement.innerText
      .replace(/\s+/g, ' ')
      .trim()
      .substring(0, 2000); // Limit to 2000 characters
  }
  
  return {
    url,
    title,
    description,
    content
  };
}

chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.action === 'getPageData') {
    sendResponse(getPageData());
  }
});