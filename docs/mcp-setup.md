# BookMinder MCP Server

BookMinder exposes an MCP server on port 9091 (HTTP/SSE transport) alongside the main API on port 9090. Connect any MCP-compatible client (Claude Desktop, Claude Code) to it for AI-assisted bookmark triage.

## Available Tools

| Tool | Description |
|------|-------------|
| `get_stats` | Dashboard overview: needsTriage, activeProjects, readyToShare, totalBookmarks |
| `list_bookmarks` | List bookmarks with optional action filter, text search, limit/offset |
| `get_bookmark` | Full bookmark detail by ID (includes content and ai_summary) |
| `update_bookmark` | Set action, tags, project, ai_summary on a single bookmark |
| `bulk_update_bookmarks` | Apply action + project to multiple bookmark IDs at once |
| `list_projects` | All projects with link counts and last-updated age |
| `create_project` | Create a new project by name |

## Configuration

### Claude Desktop

Add to `~/.config/Claude/claude_desktop_config.json` (Linux) or `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS):

```json
{
  "mcpServers": {
    "bookminder": {
      "url": "http://localhost:9091/sse"
    }
  }
}
```

Restart Claude Desktop after saving. BookMinder must be running before you open Claude Desktop.

### Claude Code

```bash
claude mcp add bookminder --transport sse http://localhost:9091/sse
```

Or add to `.claude/settings.json` in this repo:

```json
{
  "mcpServers": {
    "bookminder": {
      "type": "sse",
      "url": "http://localhost:9091/sse"
    }
  }
}
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MCP_PORT` | `9091` | Port the MCP server listens on |

## Example Triage Session

Once connected, you can ask Claude things like:

- *"How many bookmarks need triage?"* → calls `get_stats`
- *"Show me the 20 oldest untriaged bookmarks"* → calls `list_bookmarks`
- *"What's bookmark 42 about?"* → calls `get_bookmark`
- *"Mark bookmarks 12, 34, and 56 as irrelevant"* → calls `bulk_update_bookmarks`
- *"Assign bookmark 99 to the Energy project and tag it renewable, solar"* → calls `update_bookmark`
- *"Create a project called 'AI Tools'"* → calls `create_project`

## Action Values

- `read-later` — in the triage queue, not yet decided
- `working` — actively being used for a project
- `share` — worth sharing with others
- `archived` — done / no longer active
- `irrelevant` — not useful
