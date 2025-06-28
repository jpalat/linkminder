#!/usr/bin/env node

// Quick script to test if the Go server is running
// eslint-disable-next-line @typescript-eslint/no-require-imports
const http = require('http');

const SERVER_URL = 'http://localhost:9090';

function checkServer() {
  console.log(`🔍 Checking server at ${SERVER_URL}...`);
  
  const req = http.get(`${SERVER_URL}/topics`, (res) => {
    console.log(`✅ Server is running! Status: ${res.statusCode}`);
    console.log(`📊 Content-Type: ${res.headers['content-type']}`);
    
    let data = '';
    res.on('data', chunk => data += chunk);
    res.on('end', () => {
      try {
        const parsed = JSON.parse(data);
        console.log(`📋 Topics available: ${parsed.length || 'Unknown'}`);
      } catch {
        console.log(`📋 Response: ${data.substring(0, 100)}...`);
      }
    });
  });
  
  req.on('error', (err) => {
    console.log(`❌ Server not reachable: ${err.message}`);
    console.log(`\n💡 To start the Go server:`);
    console.log(`   cd ../`);
    console.log(`   go run main.go`);
  });
  
  req.setTimeout(5000, () => {
    console.log(`⏰ Server timeout - may be starting up`);
    req.destroy();
  });
}

checkServer();