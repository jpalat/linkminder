<template>
  <div class="tag-view">
    <!-- Header -->
    <header class="tag-header">
      <div class="tag-header-content">
        <div class="tag-nav">
          <router-link to="/" class="back-link">← Dashboard</router-link>
          <span class="tag-nav-sep">/</span>
          <router-link to="/tags" class="back-link">All Tags</router-link>
        </div>

        <div class="tag-title-row">
          <h1 class="tag-title">
            <span class="tag-hash">#</span>{{ tag }}
          </h1>
          <span v-if="!loading" class="tag-count-badge">{{ total }} bookmark{{ total !== 1 ? 's' : '' }}</span>
        </div>

        <!-- Action breakdown -->
        <div v-if="!loading && bookmarks.length" class="tag-breakdown">
          <span v-for="(count, action) in actionCounts" :key="action" class="breakdown-pill" :class="`breakdown-${action}`">
            {{ actionLabel(action) }} {{ count }}
          </span>
        </div>
      </div>
    </header>

    <!-- Loading -->
    <div v-if="loading" class="loading-state">Loading bookmarks for #{{ tag }}…</div>

    <!-- Error -->
    <div v-else-if="error" class="error-state">{{ error }}</div>

    <!-- Empty -->
    <div v-else-if="bookmarks.length === 0" class="empty-state">
      <p>No bookmarks tagged <strong>#{{ tag }}</strong>.</p>
      <router-link to="/" class="back-link">Back to Dashboard</router-link>
    </div>

    <!-- Bookmark list -->
    <main v-else class="tag-main">
      <div class="tag-container">
        <div v-for="bookmark in bookmarks" :key="bookmark.id" class="tag-card">
          <!-- Title row -->
          <div class="card-title-row">
            <a :href="bookmark.url" target="_blank" rel="noopener noreferrer" class="card-title" @click.stop>
              {{ bookmark.title }}
            </a>
            <span v-if="bookmark.action" class="action-badge" :class="`action-${bookmark.action}`">
              {{ actionLabel(bookmark.action) }}
            </span>
          </div>

          <!-- URL -->
          <a
            :href="bookmark.url"
            class="card-url"
            :title="bookmark.url"
            target="_blank"
            rel="noopener noreferrer"
          >{{ bookmark.url }}</a>

          <!-- Description / AI Summary -->
          <p v-if="bookmark.ai_summary || bookmark.description" class="card-description">
            {{ truncate(bookmark.ai_summary || bookmark.description || '', 220) }}
          </p>

          <!-- Tags -->
          <div v-if="bookmark.tags && bookmark.tags.length" class="card-tags">
            <router-link
              v-for="t in bookmark.tags"
              :key="t"
              :to="`/tags/${encodeURIComponent(t)}`"
              class="card-tag"
              :class="{ 'card-tag-active': t === tag }"
            >#{{ t }}</router-link>
          </div>

          <!-- Footer -->
          <div class="card-footer">
            <span class="card-domain">{{ bookmark.domain }}</span>
            <span class="card-age">{{ bookmark.age }}</span>
            <span v-if="bookmark.topic" class="card-topic">📁 {{ bookmark.topic }}</span>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import { bookmarkService } from '@/services/bookmarkService'
import type { TriageBookmark } from '@/services/bookmarkService'

const route = useRoute()

const tag = computed(() => {
  const t = route.params.tag
  return decodeURIComponent(Array.isArray(t) ? t[0] ?? '' : t ?? '')
})
const bookmarks = ref<TriageBookmark[]>([])
const total = ref(0)
const loading = ref(false)
const error = ref<string | null>(null)

const actionCounts = computed(() => {
  const counts: Record<string, number> = {}
  bookmarks.value.forEach(b => {
    const a = b.action || 'triage'
    counts[a] = (counts[a] || 0) + 1
  })
  return counts
})

const actionLabel = (action: string) => {
  const labels: Record<string, string> = {
    'read-later': 'Triage',
    'working': 'Working',
    'share': 'Share',
    'archived': 'Archived',
    'irrelevant': 'Irrelevant',
    'triage': 'Triage',
  }
  return labels[action] || action
}

const truncate = (text: string, max: number) =>
  text.length <= max ? text : text.slice(0, max) + '…'

const load = async (tagName: string) => {
  loading.value = true
  error.value = null
  try {
    const data = await bookmarkService.getBookmarksByTag(tagName)
    bookmarks.value = data.bookmarks ?? []
    total.value = data.total
  } catch (e) {
    error.value = 'Failed to load bookmarks for this tag.'
    console.error(e)
  } finally {
    loading.value = false
  }
}

watch(tag, (t) => load(t), { immediate: true })
</script>

<style scoped>
.tag-view {
  min-height: 100vh;
  background: var(--bg-primary, #f8fafc);
}

.tag-header {
  background: white;
  border-bottom: 1px solid var(--border-light);
  padding: var(--spacing-lg) var(--spacing-xl);
  position: sticky;
  top: 0;
  z-index: var(--z-sticky);
  box-shadow: 0 1px 3px rgba(0,0,0,.05);
}

.tag-header-content {
  max-width: 900px;
  margin: 0 auto;
}

.tag-nav {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-sm);
  font-size: var(--font-size-sm);
}

.tag-nav-sep {
  color: var(--color-gray-400);
}

.back-link {
  color: var(--color-primary);
  text-decoration: none;
  font-size: var(--font-size-sm);
}

.back-link:hover {
  text-decoration: underline;
}

.tag-title-row {
  display: flex;
  align-items: baseline;
  gap: var(--spacing-md);
}

.tag-title {
  font-size: 1.5rem;
  font-weight: var(--font-weight-bold);
  color: var(--color-gray-900);
  margin: 0;
}

.tag-hash {
  color: var(--color-primary);
}

.tag-count-badge {
  font-size: var(--font-size-sm);
  color: var(--color-gray-500);
}

.tag-breakdown {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-xs);
  margin-top: var(--spacing-sm);
}

.breakdown-pill {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: var(--radius-xl);
  font-weight: var(--font-weight-medium);
  background: var(--color-gray-100);
  color: var(--color-gray-600);
}

.breakdown-working { background: #dbeafe; color: #1e40af; }
.breakdown-share { background: #dcfce7; color: #166534; }
.breakdown-archived { background: var(--color-gray-100); color: var(--color-gray-600); }
.breakdown-irrelevant { background: #fee2e2; color: #991b1b; }

/* Main content */
.tag-main {
  padding: var(--spacing-xl);
}

.tag-container {
  max-width: 900px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

/* Cards */
.tag-card {
  background: white;
  border: 1px solid var(--border-light);
  border-left: 3px solid var(--color-primary);
  border-radius: var(--radius-md);
  padding: var(--spacing-md) var(--spacing-lg);
}

.card-title-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--spacing-md);
  margin-bottom: 4px;
}

.card-title {
  font-weight: var(--font-weight-semibold);
  font-size: 0.9rem;
  color: var(--color-gray-800);
  text-decoration: none;
  flex: 1;
  line-height: var(--line-height-tight);
}

.card-title:hover {
  color: var(--color-primary);
}

.action-badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: var(--radius-xl);
  font-weight: var(--font-weight-medium);
  white-space: nowrap;
  flex-shrink: 0;
  background: var(--color-gray-100);
  color: var(--color-gray-600);
}

.action-working { background: #dbeafe; color: #1e40af; }
.action-share { background: #dcfce7; color: #166534; }
.action-archived { background: var(--color-gray-100); color: var(--color-gray-500); }
.action-irrelevant { background: #fee2e2; color: #991b1b; }

.card-url {
  display: block;
  font-size: 0.78rem;
  color: var(--color-primary);
  text-decoration: underline;
  text-decoration-color: transparent;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-bottom: 6px;
  transition: text-decoration-color var(--transition-fast);
}

.card-url:hover {
  text-decoration-color: currentColor;
}

.card-description {
  font-size: var(--font-size-sm);
  color: var(--color-gray-600);
  line-height: var(--line-height-normal);
  margin: 0 0 var(--spacing-sm) 0;
}

.card-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-bottom: var(--spacing-sm);
}

.card-tag {
  font-size: 11px;
  color: var(--color-primary);
  background: var(--color-primary-light, #ede9fe);
  padding: 2px 7px;
  border-radius: var(--radius-xl);
  font-weight: var(--font-weight-medium);
  text-decoration: none;
  transition: background var(--transition-fast), color var(--transition-fast);
}

.card-tag:hover,
.card-tag-active {
  background: var(--color-primary);
  color: white;
}

.card-footer {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  font-size: var(--font-size-xs);
  color: var(--color-gray-500);
}

.card-domain {
  background: var(--color-gray-100);
  padding: 2px 7px;
  border-radius: var(--radius-xl);
}

.card-topic {
  color: var(--color-gray-500);
}

/* States */
.loading-state,
.error-state,
.empty-state {
  max-width: 900px;
  margin: var(--spacing-2xl) auto;
  padding: var(--spacing-2xl);
  text-align: center;
  color: var(--color-gray-500);
}

.error-state { color: var(--color-danger, #dc2626); }

@media (max-width: 640px) {
  .tag-header { padding: var(--spacing-md); }
  .tag-main { padding: var(--spacing-md); }
  .card-title-row { flex-direction: column; gap: var(--spacing-xs); }
}
</style>
