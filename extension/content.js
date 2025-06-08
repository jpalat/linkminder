function getPageData() {
  const url = window.location.href;
  const title = document.title;
  
  let description = '';
  
  const metaDescription = document.querySelector('meta[name="description"]');
  if (metaDescription) {
    description = metaDescription.getAttribute('content');
  } else {
    const firstParagraph = document.querySelector('p');
    if (firstParagraph) {
      description = firstParagraph.textContent.trim().substring(0, 200);
    }
  }
  
  return {
    url,
    title,
    description
  };
}

chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.action === 'getPageData') {
    sendResponse(getPageData());
  }
});