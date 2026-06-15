<template>
  <div class="mcp-docs">
    <header class="header">
      <div class="header-content">
        <div class="breadcrumb">
          <router-link to="/" class="breadcrumb-link">Dashboard</router-link>
          <span class="breadcrumb-separator">›</span>
          <span class="breadcrumb-current">MCP Setup</span>
        </div>
        <div class="header-actions">
          <router-link to="/" class="back-btn">← Back</router-link>
        </div>
      </div>
    </header>

    <main class="main-content">
      <div class="container">
        <div class="page-title">
          <h1>🔌 MCP Server Setup</h1>
          <p class="subtitle">
            BookMinder exposes an MCP (Model Context Protocol) server on port
            <code>9091</code> using SSE transport. Connect any MCP-compatible AI
            agent to read, triage, tag, and update your bookmarks.
          </p>
        </div>

        <!-- Status Card -->
        <div class="status-card">
          <div class="status-dot"></div>
          <div>
            <strong>Server endpoint</strong>
            <code class="endpoint">http://localhost:9091/sse</code>
          </div>
          <div class="env-note">
            Override port with <code>MCP_PORT</code> environment variable
          </div>
        </div>

        <!-- Available Tools -->
        <section class="section">
          <h2>Available Tools</h2>
          <div class="tools-grid">
            <div class="tool-card" v-for="tool in tools" :key="tool.name">
              <div class="tool-name">{{ tool.name }}</div>
              <div class="tool-desc">{{ tool.desc }}</div>
            </div>
          </div>
        </section>

        <!-- Claude Code -->
        <section class="section">
          <h2>Claude Code</h2>
          <p>Add to your project's <code>.claude/settings.json</code> or your global Claude Code settings:</p>
          <pre class="code-block">{{ claudeCodeConfig }}</pre>
          <p class="note">
            The server must be running (<code>go run main.go</code> or the compiled binary) before connecting.
          </p>
        </section>

        <!-- Claude Desktop -->
        <section class="section">
          <h2>Claude Desktop</h2>
          <p>Add to <code>~/Library/Application Support/Claude/claude_desktop_config.json</code> (macOS) or the equivalent on your platform:</p>
          <pre class="code-block">{{ claudeDesktopConfig }}</pre>
          <p class="note">Restart Claude Desktop after editing the config.</p>
        </section>

        <!-- Generic MCP Client -->
        <section class="section">
          <h2>Any MCP-Compatible Client</h2>
          <p>The server speaks the standard MCP protocol over SSE. Point your client at:</p>
          <div class="endpoint-table">
            <div class="endpoint-row">
              <span class="endpoint-label">SSE stream</span>
              <code>http://localhost:9091/sse</code>
            </div>
            <div class="endpoint-row">
              <span class="endpoint-label">Message endpoint</span>
              <code>http://localhost:9091/message</code>
            </div>
          </div>
        </section>

        <!-- Example Prompts -->
        <section class="section">
          <h2>Example Prompts</h2>
          <div class="prompts-list">
            <div class="prompt-item" v-for="prompt in examplePrompts" :key="prompt">
              <span class="prompt-quote">"</span>{{ prompt }}<span class="prompt-quote">"</span>
            </div>
          </div>
        </section>

        <!-- Environment -->
        <section class="section">
          <h2>Environment Variables</h2>
          <div class="endpoint-table">
            <div class="endpoint-row">
              <span class="endpoint-label"><code>MCP_PORT</code></span>
              <span>Override the default port <code>9091</code></span>
            </div>
          </div>
        </section>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
const tools = [
  { name: 'get_stats', desc: 'Dashboard summary: triage count, active projects, ready-to-share, totals.' },
  { name: 'list_bookmarks', desc: 'List bookmarks with optional action filter, text search, and pagination.' },
  { name: 'get_bookmark', desc: 'Full details for a single bookmark by ID, including content, tags, and ai_summary.' },
  { name: 'update_bookmark', desc: 'Update action, tags, project assignment, and/or ai_summary for a bookmark.' },
  { name: 'bulk_update_bookmarks', desc: 'Set action (and optionally project) on multiple bookmarks at once.' },
  { name: 'list_projects', desc: 'All active projects with bookmark counts and last-updated timestamps.' },
  { name: 'create_project', desc: 'Create a new named project with an optional description.' },
]

const claudeCodeConfig = `{
  "mcpServers": {
    "bookminder": {
      "url": "http://localhost:9091/sse"
    }
  }
}`

const claudeDesktopConfig = `{
  "mcpServers": {
    "bookminder": {
      "url": "http://localhost:9091/sse"
    }
  }
}`

const examplePrompts = [
  'Show me my triage queue and summarize what needs attention.',
  'What bookmarks are ready to share, and who are they for?',
  'Find all bookmarks tagged with "mcp" and move them to the Agents project.',
  'Summarize the last 10 bookmarks I saved and suggest tags for each.',
  'Which projects have gone stale? Archive anything inactive for over 30 days.',
]
</script>

<style scoped>
.mcp-docs {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--color-gray-50, #f9fafb);
}

/* Header */
.header {
  background: white;
  border-bottom: 1px solid var(--color-gray-200, #e5e7eb);
  padding: var(--spacing-md, 12px) var(--spacing-xl, 24px);
  position: sticky;
  top: 0;
  z-index: 10;
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  max-width: 900px;
  margin: 0 auto;
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: var(--font-size-sm, 14px);
}

.breadcrumb-link {
  color: var(--color-primary, #6366f1);
  text-decoration: none;
}

.breadcrumb-link:hover {
  text-decoration: underline;
}

.breadcrumb-separator {
  color: var(--color-gray-400, #9ca3af);
}

.breadcrumb-current {
  color: var(--color-gray-600, #4b5563);
  font-weight: var(--font-weight-medium, 500);
}

.back-btn {
  font-size: var(--font-size-sm, 14px);
  color: var(--color-gray-600, #4b5563);
  text-decoration: none;
  padding: 6px 12px;
  border: 1px solid var(--color-gray-200, #e5e7eb);
  border-radius: var(--border-radius, 6px);
  background: white;
  transition: background 0.15s;
}

.back-btn:hover {
  background: var(--color-gray-50, #f9fafb);
}

/* Main */
.main-content {
  flex: 1;
  padding: var(--spacing-xl, 24px);
}

.container {
  max-width: 900px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 32px;
}

/* Page title */
.page-title h1 {
  font-size: 24px;
  font-weight: var(--font-weight-bold, 700);
  color: var(--color-gray-900, #111827);
  margin: 0 0 8px 0;
}

.subtitle {
  color: var(--color-gray-600, #4b5563);
  font-size: var(--font-size-sm, 14px);
  line-height: 1.6;
  margin: 0;
}

/* Status card */
.status-card {
  background: white;
  border: 1px solid var(--color-gray-200, #e5e7eb);
  border-radius: var(--border-radius, 6px);
  padding: 16px 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.status-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #22c55e;
  flex-shrink: 0;
}

.status-card strong {
  display: block;
  font-size: var(--font-size-sm, 14px);
  color: var(--color-gray-700, #374151);
  margin-bottom: 4px;
}

.endpoint {
  font-family: monospace;
  font-size: 13px;
  background: var(--color-gray-100, #f3f4f6);
  padding: 2px 8px;
  border-radius: 4px;
  color: var(--color-primary, #6366f1);
}

.env-note {
  margin-left: auto;
  font-size: var(--font-size-xs, 12px);
  color: var(--color-gray-500, #6b7280);
}

/* Sections */
.section {
  background: white;
  border: 1px solid var(--color-gray-200, #e5e7eb);
  border-radius: var(--border-radius, 6px);
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.section h2 {
  font-size: 16px;
  font-weight: var(--font-weight-semibold, 600);
  color: var(--color-gray-900, #111827);
  margin: 0;
}

.section p {
  font-size: var(--font-size-sm, 14px);
  color: var(--color-gray-600, #4b5563);
  margin: 0;
  line-height: 1.6;
}

.note {
  font-size: var(--font-size-xs, 12px) !important;
  color: var(--color-gray-500, #6b7280) !important;
}

code {
  font-family: monospace;
  font-size: 13px;
  background: var(--color-gray-100, #f3f4f6);
  padding: 1px 6px;
  border-radius: 4px;
  color: var(--color-gray-800, #1f2937);
}

/* Tools grid */
.tools-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 10px;
}

.tool-card {
  background: var(--color-gray-50, #f9fafb);
  border: 1px solid var(--color-gray-200, #e5e7eb);
  border-radius: var(--border-radius, 6px);
  padding: 12px 14px;
}

.tool-name {
  font-family: monospace;
  font-size: 13px;
  font-weight: var(--font-weight-semibold, 600);
  color: var(--color-primary, #6366f1);
  margin-bottom: 4px;
}

.tool-desc {
  font-size: var(--font-size-xs, 12px);
  color: var(--color-gray-600, #4b5563);
  line-height: 1.5;
}

/* Code block */
.code-block {
  background: var(--color-gray-900, #111827);
  color: #e5e7eb;
  border-radius: var(--border-radius, 6px);
  padding: 16px;
  font-family: monospace;
  font-size: 13px;
  line-height: 1.6;
  overflow-x: auto;
  margin: 0;
  white-space: pre;
}

/* Endpoint table */
.endpoint-table {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.endpoint-row {
  display: flex;
  align-items: center;
  gap: 16px;
  font-size: var(--font-size-sm, 14px);
}

.endpoint-label {
  min-width: 160px;
  color: var(--color-gray-500, #6b7280);
  font-size: var(--font-size-sm, 14px);
}

/* Example prompts */
.prompts-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.prompt-item {
  background: var(--color-gray-50, #f9fafb);
  border: 1px solid var(--color-gray-200, #e5e7eb);
  border-radius: var(--border-radius, 6px);
  padding: 10px 14px;
  font-size: var(--font-size-sm, 14px);
  color: var(--color-gray-700, #374151);
  font-style: italic;
  line-height: 1.5;
}

.prompt-quote {
  color: var(--color-gray-400, #9ca3af);
  font-style: normal;
}
</style>
