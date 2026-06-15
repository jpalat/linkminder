<template>
  <div class="tags-index">
    <header class="tags-header">
      <div class="tags-header-content">
        <router-link to="/" class="back-link">← Dashboard</router-link>
        <h1 class="tags-title">All Tags</h1>
        <p v-if="!loading" class="tags-subtitle">{{ tags.length }} tag{{ tags.length !== 1 ? 's' : '' }} across {{ totalBookmarks }} bookmark{{ totalBookmarks !== 1 ? 's' : '' }}</p>
      </div>
    </header>

    <main class="tags-main">
      <div class="tags-container">
        <div v-if="loading" class="loading-state">Loading tags…</div>
        <div v-else-if="error" class="error-state">{{ error }}</div>
        <div v-else-if="tags.length === 0" class="empty-state">No tags yet. Add tags to your bookmarks to see them here.</div>
        <div v-else class="tags-grid">
          <router-link
            v-for="t in tags"
            :key="t.tag"
            :to="`/tags/${encodeURIComponent(t.tag)}`"
            class="tag-card"
          >
            <span class="tag-name">#{{ t.tag }}</span>
            <span class="tag-count">{{ t.count }}</span>
          </router-link>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { bookmarkService } from '@/services/bookmarkService'
import type { TagCount } from '@/services/bookmarkService'

const tags = ref<TagCount[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

const totalBookmarks = computed(() =>
  tags.value.reduce((sum, t) => sum + t.count, 0)
)

onMounted(async () => {
  loading.value = true
  try {
    const data = await bookmarkService.getTags()
    tags.value = data.tags ?? []
  } catch (e) {
    error.value = 'Failed to load tags.'
    console.error(e)
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.tags-index {
  min-height: 100vh;
  background: var(--bg-primary, #f8fafc);
}

.tags-header {
  background: white;
  border-bottom: 1px solid var(--border-light);
  padding: var(--spacing-lg) var(--spacing-xl);
  box-shadow: 0 1px 3px rgba(0,0,0,.05);
}

.tags-header-content {
  max-width: 900px;
  margin: 0 auto;
}

.back-link {
  color: var(--color-primary);
  text-decoration: none;
  font-size: var(--font-size-sm);
  display: block;
  margin-bottom: var(--spacing-sm);
}

.back-link:hover { text-decoration: underline; }

.tags-title {
  font-size: 1.5rem;
  font-weight: var(--font-weight-bold);
  color: var(--color-gray-900);
  margin: 0 0 4px 0;
}

.tags-subtitle {
  font-size: var(--font-size-sm);
  color: var(--color-gray-500);
  margin: 0;
}

.tags-main {
  padding: var(--spacing-xl);
}

.tags-container {
  max-width: 900px;
  margin: 0 auto;
}

.tags-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  gap: var(--spacing-sm);
}

.tag-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-md) var(--spacing-lg);
  background: white;
  border: 1px solid var(--border-light);
  border-radius: var(--radius-md);
  text-decoration: none;
  transition: all var(--transition-fast);
  gap: var(--spacing-sm);
}

.tag-card:hover {
  border-color: var(--color-primary);
  background: var(--color-primary-light, #ede9fe);
  transform: translateY(-1px);
  box-shadow: var(--shadow-sm);
}

.tag-name {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tag-count {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  color: white;
  background: var(--color-primary);
  padding: 1px 7px;
  border-radius: var(--radius-xl);
  flex-shrink: 0;
}

.loading-state,
.error-state,
.empty-state {
  padding: var(--spacing-2xl);
  text-align: center;
  color: var(--color-gray-500);
}

.error-state { color: var(--color-danger, #dc2626); }

@media (max-width: 640px) {
  .tags-header { padding: var(--spacing-md); }
  .tags-main { padding: var(--spacing-md); }
  .tags-grid { grid-template-columns: repeat(auto-fill, minmax(130px, 1fr)); }
}
</style>
