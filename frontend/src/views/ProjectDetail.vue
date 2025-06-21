<template>
  <div class="project-detail">
    <header class="header">
      <div class="header-content">
        <div class="breadcrumb">
          <router-link to="/" class="breadcrumb-link">Dashboard</router-link>
          <span class="breadcrumb-separator">></span>
          <span class="breadcrumb-current">{{ projectName }}</span>
        </div>
        
        <div class="header-actions">
          <AppButton variant="secondary" @click="$router.back()">
            ← Back
          </AppButton>
          <AppButton variant="primary">
            Export
          </AppButton>
        </div>
      </div>
    </header>
    
    <main class="main-content">
      <div class="container">
        <div class="section">
          <div class="section-header">
            <div class="section-title">
              📁 {{ projectName }} Project
            </div>
          </div>
          <div class="section-content">
            <div class="empty-state">
              <h3>Project Detail View</h3>
              <p>This will be implemented to match your project detail mockup.</p>
              <p>Project ID: {{ $route.params.id }}</p>
            </div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import AppButton from '@/components/ui/AppButton.vue'

const route = useRoute()

const projectName = computed(() => {
  // This would normally come from the store or API based on route.params.id
  return route.params.id || 'Unknown Project'
})
</script>

<style scoped>
.project-detail {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.header {
  background: white;
  border-bottom: 1px solid var(--border-light);
  padding: var(--spacing-lg) var(--spacing-xl);
  box-shadow: var(--shadow-sm);
}

.header-content {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  color: var(--color-gray-600);
  font-size: var(--font-size-base);
}

.breadcrumb-link {
  color: var(--color-primary);
  text-decoration: none;
}

.breadcrumb-link:hover {
  text-decoration: underline;
}

.breadcrumb-separator {
  color: var(--color-gray-400);
}

.breadcrumb-current {
  font-weight: var(--font-weight-semibold);
  color: var(--color-gray-800);
}

.header-actions {
  display: flex;
  gap: var(--spacing-sm);
}

.main-content {
  flex: 1;
  padding: var(--spacing-xl);
}

.container {
  max-width: 1200px;
  margin: 0 auto;
}
</style>
