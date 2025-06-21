<template>
  <div class="project-list">
    <div v-if="projects.length === 0" class="empty-state">
      <h3>No active projects</h3>
      <p>Start working on bookmarks to create projects automatically.</p>
    </div>
    
    <div v-else class="projects-grid">
      <div
        v-for="project in projects"
        :key="project.topic"
        class="project-card"
        @click="handleProjectClick(project)"
      >
        <div class="project-header">
          <h3 class="project-title">{{ project.topic }}</h3>
          <AppBadge :variant="getStatusVariant(project.status)" size="sm">
            {{ project.status }}
          </AppBadge>
        </div>
        
        <div class="project-meta">
          <span>{{ project.count }} links</span>
          <span>Updated {{ formatDate(project.lastUpdated) }}</span>
        </div>
        
        <div class="project-actions">
          <AppButton size="xs" variant="primary" @click.stop="handleProjectClick(project)">
            View Details →
          </AppButton>
          <AppButton size="xs" variant="secondary" @click.stop>
            Export
          </AppButton>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import type { ProjectStat } from '@/types'
import AppButton from '@/components/ui/AppButton.vue'
import AppBadge from '@/components/ui/AppBadge.vue'

interface Props {
  projects: ProjectStat[]
}

defineProps<Props>()

const router = useRouter()

const emit = defineEmits<{
  'project-click': [project: ProjectStat]
}>()

const handleProjectClick = (project: ProjectStat) => {
  emit('project-click', project)
  // Navigate to project detail page
  router.push(`/project/${encodeURIComponent(project.topic)}`)
}

const getStatusVariant = (status: string): 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info' => {
  const variants: Record<string, 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info'> = {
    'active': 'success',
    'stale': 'warning',
    'inactive': 'default'
  }
  return variants[status] || 'default'
}

const formatDate = (dateString: string) => {
  const date = new Date(dateString)
  const now = new Date()
  const diffInMs = now.getTime() - date.getTime()
  const diffInHours = Math.floor(diffInMs / (1000 * 60 * 60))
  
  if (diffInHours < 1) return 'just now'
  if (diffInHours < 24) return `${diffInHours}h ago`
  
  const diffInDays = Math.floor(diffInHours / 24)
  if (diffInDays < 7) return `${diffInDays}d ago`
  
  const diffInWeeks = Math.floor(diffInDays / 7)
  return `${diffInWeeks}w ago`
}
</script>

<style scoped>
.project-list {
  width: 100%;
}

.projects-grid {
  display: grid;
  gap: var(--spacing-lg);
  max-height: 400px;
  overflow-y: auto;
}

.project-card {
  padding: var(--spacing-lg);
  background: var(--bg-card-hover);
  border-radius: var(--radius-lg);
  border-left: 4px solid var(--color-primary);
  cursor: pointer;
  transition: var(--transition-fast);
}

.project-card:hover {
  background: var(--color-gray-200);
  transform: translateY(-1px);
  box-shadow: var(--shadow-lg);
}

.project-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: var(--spacing-md);
  gap: var(--spacing-sm);
}

.project-title {
  font-weight: var(--font-weight-semibold);
  margin-bottom: var(--spacing-xs);
  color: var(--color-gray-800);
  font-size: var(--font-size-base);
  flex: 1;
}

.project-meta {
  font-size: var(--font-size-sm);
  color: var(--color-gray-600);
  display: flex;
  justify-content: space-between;
  margin-bottom: var(--spacing-md);
}

.project-actions {
  display: flex;
  gap: var(--spacing-sm);
  opacity: 0;
  transition: opacity var(--transition-fast);
}

.project-card:hover .project-actions {
  opacity: 1;
}

/* Responsive */
@media (max-width: 768px) {
  .project-actions {
    opacity: 1;
  }
  
  .project-meta {
    flex-direction: column;
    gap: var(--spacing-xs);
  }
}
</style>
