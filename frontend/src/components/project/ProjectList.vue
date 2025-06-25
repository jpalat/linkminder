<template>
  <div class="project-list">
    <!-- Header Controls -->
    <div class="project-header-controls">
      <div class="project-count-indicator">
        <span class="project-count">{{ projects.length }} projects</span>
        <span class="view-indicator">Compact</span>
      </div>
      <div class="header-controls">
        <input 
          type="text" 
          class="search-box" 
          placeholder="Search projects, tags..."
          v-model="searchQuery"
          @input="handleSearch"
        >
        <button class="filter-btn" :class="{ active: showActiveOnly }" @click="toggleActiveFilter">
          {{ showActiveOnly ? 'All' : 'Active' }}
        </button>
        <button class="view-btn" @click="toggleCompactView">
          {{ isCompactView ? 'Grid' : 'Compact' }}
        </button>
      </div>
    </div>

    <!-- Empty State -->
    <div v-if="filteredProjects.length === 0 && projects.length === 0" class="empty-state">
      <div class="empty-icon">📁</div>
      <h3>No active projects</h3>
      <p>Start working on bookmarks to create projects automatically.</p>
    </div>

    <!-- No Results State -->
    <div v-else-if="filteredProjects.length === 0" class="empty-state">
      <div class="empty-icon">🔍</div>
      <h3>No projects found</h3>
      <p>Try adjusting your search or filters.</p>
    </div>
    
    <!-- Projects Grid -->
    <div v-else class="projects-grid" :class="{ 'compact-view': isCompactView }">
      <div
        v-for="project in filteredProjects"
        :key="project.topic"
        class="project-card"
        @click="handleProjectClick(project)"
      >
        <!-- Project Header -->
        <div class="project-header">
          <div class="project-title-row">
            <h3 class="project-title">{{ project.topic }}</h3>
            <span class="project-status" :class="getStatusClass(project.status)">
              {{ project.status }}
            </span>
          </div>
          <div class="project-meta">
            <span class="meta-item">📚 {{ project.count }}</span>
            <span class="meta-item">🕒 {{ formatDate(project.lastUpdated) }}</span>
          </div>
          <div class="tags-row">
            <span class="tag primary">{{ project.topic.split(' ')[0] }}</span>
            <span v-if="project.topic.split(' ').length > 1" class="tag">{{ project.topic.split(' ').slice(1).join(' ') }}</span>
            <span class="tag" :class="{ 'tag-active': project.status === 'active' }">
              {{ project.status === 'active' ? 'Current' : project.status }}
            </span>
          </div>
        </div>

        <!-- Project Summary -->
        <div class="project-summary">
          <div class="summary-stats">
            <span class="stat">📖 Links: {{ project.count }}</span>
            <span class="stat">⏱️ {{ getActivityLevel(project.lastUpdated) }}</span>
          </div>
          <div class="last-activity">
            Last: {{ formatDate(project.lastUpdated) }}
          </div>
        </div>

        <!-- Sample Bookmarks Preview (simulated) -->
        <div class="bookmarks-preview" v-if="!isCompactView">
          <div class="bookmark-item">
            <div class="bookmark-favicon">📖</div>
            <div class="bookmark-content">
              <div class="bookmark-title">{{ project.latestTitle || getSampleBookmarkTitle(project.topic) }}</div>
              <div class="bookmark-domain">{{ getActualDomain(project) }}</div>
              <div class="bookmark-tags">
                <span class="bookmark-tag">{{ project.status }}</span>
                <span class="bookmark-tag">{{ project.topic.toLowerCase() }}</span>
              </div>
              <div class="bookmark-meta">{{ formatDate(project.lastUpdated) }} • Working</div>
            </div>
          </div>
          
          <button class="show-more-btn" @click.stop="handleProjectClick(project)">
            ⌄ View all {{ project.count }} references
          </button>
        </div>

        <!-- Project Footer -->
        <div class="project-footer">
          <span class="bookmark-count">{{ project.count }} total</span>
          <div class="quick-actions">
            <button class="action-btn primary" @click.stop="handleProjectClick(project)">
              View
            </button>
            <button class="action-btn" @click.stop="handleExport(project)">
              Export
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import type { ProjectStat } from '@/types'

interface Props {
  projects: ProjectStat[]
}

const props = defineProps<Props>()

const router = useRouter()
const searchQuery = ref('')
const showActiveOnly = ref(false)
const isCompactView = ref(false)

const emit = defineEmits<{
  'project-click': [project: ProjectStat]
  'export-project': [project: ProjectStat]
}>()

// Computed filtered projects
const filteredProjects = computed(() => {
  let filtered = props.projects

  // Filter by search query
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(project => 
      project.topic.toLowerCase().includes(query)
    )
  }

  // Filter by active status
  if (showActiveOnly.value) {
    filtered = filtered.filter(project => project.status === 'active')
  }

  return filtered
})

const handleProjectClick = (project: ProjectStat) => {
  emit('project-click', project)
  router.push(`/project/${encodeURIComponent(project.topic)}`)
}

const handleExport = (project: ProjectStat) => {
  emit('export-project', project)
}

const handleSearch = () => {
  // Search is reactive through computed property
}

const toggleActiveFilter = () => {
  showActiveOnly.value = !showActiveOnly.value
}

const toggleCompactView = () => {
  isCompactView.value = !isCompactView.value
}

const getStatusClass = (status: string): string => {
  const classes: Record<string, string> = {
    'active': 'status-active',
    'stale': 'status-stale', 
    'inactive': 'status-inactive'
  }
  return classes[status] || 'status-inactive'
}

const getActivityLevel = (dateString: string): string => {
  const date = new Date(dateString)
  const now = new Date()
  const diffInHours = Math.floor((now.getTime() - date.getTime()) / (1000 * 60 * 60))
  
  if (diffInHours < 24) return 'Very Active'
  if (diffInHours < 168) return 'Active' // 1 week
  return 'Quiet'
}

const getSampleBookmarkTitle = (topic: string): string => {
  const samples: Record<string, string> = {
    'React': 'React 18 Concurrent Features Guide',
    'Vue': 'Vue 3 Composition API Deep Dive',
    'TypeScript': 'Advanced TypeScript Patterns',
    'Performance': 'Core Web Vitals Optimization',
    'Docker': 'Kubernetes Best Practices',
    'AI': 'GPT-4 API Documentation'
  }
  
  const key = Object.keys(samples).find(k => topic.includes(k))
  return key ? samples[key] : `${topic} - Latest Resource`
}

const getSampleDomain = (topic: string): string => {
  const domains: Record<string, string> = {
    'React': 'reactjs.org',
    'Vue': 'vuejs.org', 
    'TypeScript': 'typescriptlang.org',
    'Performance': 'web.dev',
    'Docker': 'kubernetes.io',
    'AI': 'openai.com'
  }
  
  const key = Object.keys(domains).find(k => topic.includes(k))
  return key ? domains[key] : 'docs.example.com'
}

const getActualDomain = (project: ProjectStat): string => {
  // Use real URL if available
  if (project.latestURL) {
    try {
      const url = new URL(project.latestURL)
      return url.hostname
    } catch (e) {
      // Fallback to sample domain if URL parsing fails
      return getSampleDomain(project.topic)
    }
  }
  // Fallback to sample domain if no URL available
  return getSampleDomain(project.topic)
}

const formatDate = (dateString: string) => {
  const date = new Date(dateString)
  const now = new Date()
  const diffInMs = now.getTime() - date.getTime()
  const diffInHours = Math.floor(diffInMs / (1000 * 60 * 60))
  
  if (diffInHours < 1) return 'now'
  if (diffInHours < 24) return `${diffInHours}h`
  
  const diffInDays = Math.floor(diffInHours / 24)
  if (diffInDays < 7) return `${diffInDays}d`
  
  const diffInWeeks = Math.floor(diffInDays / 7)
  if (diffInWeeks < 4) return `${diffInWeeks}w`
  
  const diffInMonths = Math.floor(diffInDays / 30)
  return `${diffInMonths}mo`
}
</script>

<style scoped>
.project-list {
  width: 100%;
  font-size: 12px;
  line-height: 1.2;
}

/* Header Controls */
.project-header-controls {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid #e2e8f0;
}

.project-count-indicator {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.project-count {
  font-size: 0.8rem;
  color: #718096;
  font-weight: 500;
}

.view-indicator {
  font-size: 0.65rem;
  color: #4c51bf;
  background: #e6fffa;
  padding: 0.2rem 0.5rem;
  border-radius: 8px;
}

.header-controls {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.search-box {
  padding: 0.3rem 0.5rem;
  border: 1px solid #cbd5e0;
  border-radius: 3px;
  font-size: 0.75rem;
  width: 200px;
}

.filter-btn, .view-btn {
  padding: 0.3rem 0.5rem;
  border: 1px solid #cbd5e0;
  background: white;
  border-radius: 3px;
  font-size: 0.7rem;
  cursor: pointer;
  transition: all 0.15s ease;
}

.filter-btn:hover, .view-btn:hover {
  background: #f7fafc;
}

.filter-btn.active {
  background: #4c51bf;
  color: white;
  border-color: #4c51bf;
}

/* Empty States */
.empty-state {
  text-align: center;
  padding: 3rem 2rem;
  color: #718096;
}

.empty-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
  opacity: 0.5;
}

.empty-state h3 {
  font-size: 1.1rem;
  color: #4a5568;
  margin-bottom: 0.5rem;
}

.empty-state p {
  font-size: 0.9rem;
}

/* Projects Grid */
.projects-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 0.75rem;
}

.projects-grid.compact-view {
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 0.5rem;
}

.project-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  transition: all 0.15s ease;
  cursor: pointer;
}

.project-card:hover {
  border-color: #4c51bf;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

/* Project Header */
.project-header {
  padding: 0.75rem;
  background: #f8fafc;
  border-bottom: 1px solid #e2e8f0;
}

.project-title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.project-title {
  font-size: 1rem;
  font-weight: 600;
  color: #2d3748;
  margin: 0;
}

.project-title:hover {
  color: #4c51bf;
}

.project-status {
  padding: 0.15rem 0.5rem;
  border-radius: 8px;
  font-size: 0.65rem;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.status-active {
  background: #d4edda;
  color: #155724;
}

.status-stale {
  background: #fff3cd;
  color: #856404;
}

.status-inactive {
  background: #f8d7da;
  color: #721c24;
}

.project-meta {
  display: flex;
  gap: 0.75rem;
  font-size: 0.7rem;
  color: #718096;
  margin-bottom: 0.5rem;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.tags-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem;
}

.tag {
  padding: 0.1rem 0.4rem;
  background: #edf2f7;
  color: #4a5568;
  font-size: 0.65rem;
  font-weight: 500;
  border-radius: 6px;
  cursor: pointer;
}

.tag:hover {
  background: #e2e8f0;
}

.tag.primary {
  background: #bee3f8;
  color: #2c5282;
}

.tag.tag-active {
  background: #c6f6d5;
  color: #22543d;
}

/* Project Summary */
.project-summary {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.4rem 0.75rem;
  background: #f0f9ff;
  border-bottom: 1px solid #e2e8f0;
  font-size: 0.7rem;
}

.summary-stats {
  display: flex;
  gap: 0.75rem;
  color: #1e40af;
}

.stat {
  white-space: nowrap;
}

.last-activity {
  color: #64748b;
  font-size: 0.65rem;
}

/* Bookmarks Preview */
.bookmarks-preview {
  padding: 0.5rem 0.75rem 0.75rem;
}

.bookmark-item {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  padding: 0.4rem 0;
  border-bottom: 1px solid #f7fafc;
}

.bookmark-item:hover {
  background: #f8fafc;
  margin: 0 -0.75rem;
  padding-left: 0.75rem;
  padding-right: 0.75rem;
}

.bookmark-favicon {
  width: 14px;
  height: 14px;
  border-radius: 2px;
  background: #e2e8f0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.6rem;
  flex-shrink: 0;
  margin-top: 1px;
}

.bookmark-content {
  flex: 1;
  min-width: 0;
}

.bookmark-title {
  font-weight: 500;
  color: #2d3748;
  font-size: 0.8rem;
  line-height: 1.2;
  margin-bottom: 0.2rem;
  cursor: pointer;
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.bookmark-title:hover {
  color: #4c51bf;
}

.bookmark-domain {
  color: #4c51bf;
  font-size: 0.7rem;
  margin-bottom: 0.2rem;
}

.bookmark-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.2rem;
  margin-bottom: 0.2rem;
}

.bookmark-tag {
  padding: 0.05rem 0.3rem;
  background: #e6fffa;
  color: #234e52;
  font-size: 0.6rem;
  font-weight: 500;
  border-radius: 4px;
}

.bookmark-meta {
  font-size: 0.65rem;
  color: #a0aec0;
}

.show-more-btn {
  width: 100%;
  padding: 0.3rem;
  border: 1px dashed #cbd5e0;
  background: #f8fafc;
  color: #718096;
  font-size: 0.7rem;
  cursor: pointer;
  border-radius: 4px;
  margin-top: 0.3rem;
  transition: all 0.15s ease;
}

.show-more-btn:hover {
  background: #f0f4f8;
  color: #4a5568;
}

/* Project Footer */
.project-footer {
  padding: 0.5rem 0.75rem;
  background: #f8fafc;
  border-top: 1px solid #e2e8f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.bookmark-count {
  font-size: 0.7rem;
  color: #718096;
  font-weight: 500;
}

.quick-actions {
  display: flex;
  gap: 0.4rem;
}

.action-btn {
  padding: 0.25rem 0.5rem;
  border: 1px solid #e2e8f0;
  background: white;
  border-radius: 3px;
  font-size: 0.65rem;
  cursor: pointer;
  transition: all 0.15s ease;
}

.action-btn:hover {
  background: #f7fafc;
  border-color: #cbd5e0;
}

.action-btn.primary {
  background: #4c51bf;
  color: white;
  border-color: #4c51bf;
}

.action-btn.primary:hover {
  background: #434190;
}

/* Compact view adjustments */
.compact-view .project-header {
  padding: 0.5rem;
}

.compact-view .project-summary {
  padding: 0.3rem 0.5rem;
}

.compact-view .project-footer {
  padding: 0.4rem 0.5rem;
}

.compact-view .bookmark-item {
  padding: 0.25rem 0;
}

/* Responsive */
@media (max-width: 768px) {
  .project-header-controls {
    flex-direction: column;
    gap: 0.75rem;
    align-items: stretch;
  }

  .header-controls {
    flex-wrap: wrap;
    justify-content: space-between;
  }

  .search-box {
    width: 100%;
    margin-bottom: 0.5rem;
  }

  .projects-grid {
    grid-template-columns: 1fr;
  }

  .project-title-row {
    flex-direction: column;
    gap: 0.3rem;
    align-items: flex-start;
  }

  .project-meta {
    flex-wrap: wrap;
    gap: 0.3rem;
  }

  .project-footer {
    flex-direction: column;
    gap: 0.4rem;
    align-items: stretch;
  }

  .quick-actions {
    justify-content: space-between;
  }
}

@media (max-width: 480px) {
  .project-list {
    font-size: 11px;
  }

  .project-header {
    padding: 0.5rem;
  }

  .bookmarks-preview {
    padding: 0.3rem 0.5rem;
  }

  .project-footer {
    padding: 0.4rem 0.5rem;
  }
}
</style>
