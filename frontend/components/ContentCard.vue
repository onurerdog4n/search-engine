<template>
  <div class="card">
    <!-- Header -->
    <div style="display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 1rem;">
      <div style="flex: 1;">
        <h3 style="margin-bottom: 0.5rem; font-size: 1.25rem;">{{ content.title }}</h3>
        <p style="margin-bottom: 0.75rem; font-size: 0.9375rem; line-height: 1.5;">
          {{ content.description }}
        </p>
      </div>
      
      <!-- Score Badge -->
      <div style="margin-left: 1rem; text-align: center; min-width: 70px;">
        <div style="font-size: 1.75rem; font-weight: 700; color: var(--color-accent); line-height: 1;">
          {{ content.score?.final_score?.toFixed(1) || '0.0' }}
        </div>
        <div style="font-size: 0.75rem; color: var(--color-text-secondary); margin-top: 0.25rem;">
          Skor
        </div>
      </div>
    </div>

    <!-- Meta Info -->
    <div style="display: flex; gap: 0.75rem; flex-wrap: wrap; margin-bottom: 1rem;">
      <!-- Content Type Badge -->
      <span class="badge badge-primary">
        {{ content.content_type === 'video' ? '📹 Video' : '📄 Makale' }}
      </span>
      
      <!-- Date -->
      <span class="badge">
        {{ formatDate(content.published_at) }}
      </span>

      <!-- Relevance Score Badge -->
      <span v-if="content.relevance_score > 0" class="badge" style="background: var(--color-accent); color: white; border: none;">
        🎯 Alakalılık: {{ Math.min(Math.round(content.relevance_score * 1000) / 10, 100) }}%
      </span>
    </div>

    <!-- Stats -->
    <div v-if="content.stats" style="display: flex; gap: 1.5rem; flex-wrap: wrap; margin-bottom: 1rem; font-size: 0.875rem; color: var(--color-text-secondary);">
      <div v-if="content.stats.views > 0" style="display: flex; align-items: center; gap: 0.375rem;">
        <span>👁️</span>
        <span>{{ formatNumber(content.stats.views) }}</span>
      </div>
      <div v-if="content.stats.likes > 0" style="display: flex; align-items: center; gap: 0.375rem;">
        <span>❤️</span>
        <span>{{ formatNumber(content.stats.likes) }}</span>
      </div>
      <div v-if="content.stats.reading_time > 0" style="display: flex; align-items: center; gap: 0.375rem;">
        <span>📖</span>
        <span>{{ content.stats.reading_time }} dk</span>
      </div>
      <div v-if="content.stats.reactions > 0" style="display: flex; align-items: center; gap: 0.375rem;">
        <span>💬</span>
        <span>{{ formatNumber(content.stats.reactions) }}</span>
      </div>
    </div>

    <!-- Tags -->
    <div v-if="content.tags?.length" style="display: flex; gap: 0.5rem; flex-wrap: wrap; margin-bottom: 1rem;">
      <span
        v-for="tag in content.tags"
        :key="tag.id"
        style="padding: 0.25rem 0.625rem; font-size: 0.8125rem; background: var(--color-bg-secondary); color: var(--color-text-secondary); border-radius: var(--radius-full);"
      >
        #{{ tag.name }}
      </span>
    </div>

    <!-- Score Details (Collapsible) -->
    <details style="margin-top: 1rem; cursor: pointer;">
      <summary style="font-size: 0.875rem; color: var(--color-accent); font-weight: 500; user-select: none;">
        Skor Detayları
      </summary>
      <div style="margin-top: 0.75rem; padding: 1rem; background: var(--color-bg-secondary); border-radius: var(--radius-sm); font-size: 0.875rem;">
        <div style="display: grid; gap: 0.5rem;">
          <div style="display: flex; justify-content: space-between;">
            <span style="color: var(--color-text-secondary);">Temel Skor:</span>
            <span style="font-weight: 500;">{{ content.score?.base_score?.toFixed(2) || '0.00' }}</span>
          </div>
          <div style="display: flex; justify-content: space-between;">
            <span style="color: var(--color-text-secondary);">Tür Katsayısı:</span>
            <span style="font-weight: 500;">×{{ content.score?.type_weight || '1.0' }}</span>
          </div>
          <div style="display: flex; justify-content: space-between;">
            <span style="color: var(--color-text-secondary);">Güncellik:</span>
            <span style="font-weight: 500; color: #10b981;">+{{ content.score?.recency_score || '0' }}</span>
          </div>
          <div style="display: flex; justify-content: space-between;">
            <span style="color: var(--color-text-secondary);">Etkileşim:</span>
            <span style="font-weight: 500; color: #10b981;">+{{ content.score?.engagement_score?.toFixed(2) || '0.00' }}</span>
          </div>
          <div v-if="content.relevance_score > 0" style="display: flex; justify-content: space-between;">
            <span style="color: var(--color-text-secondary);">Alakalılık Skoru (FTS):</span>
            <span style="font-weight: 500; color: var(--color-accent);">{{ content.relevance_score.toFixed(4) }}</span>
          </div>
          <div style="border-top: 1px solid var(--color-border); margin-top: 0.5rem; padding-top: 0.5rem; display: flex; justify-content: space-between;">
            <span style="font-weight: 600;">Toplam Skor:</span>
            <span style="font-weight: 700; color: var(--color-accent);">{{ content.score?.final_score?.toFixed(2) || '0.00' }}</span>
          </div>
        </div>
      </div>
    </details>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  content: any
}>()

const formatDate = (date: string) => {
  return new Date(date).toLocaleDateString('tr-TR', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
}

const formatNumber = (num: number) => {
  return new Intl.NumberFormat('tr-TR').format(num)
}
</script>
