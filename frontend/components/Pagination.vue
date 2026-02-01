<template>
  <div style="display: flex; justify-content: center; align-items: center; gap: 0.5rem; flex-wrap: wrap;">
    <!-- Previous Button -->
    <button
      @click="$emit('page-change', currentPage - 1)"
      :disabled="currentPage === 1"
      class="btn btn-secondary"
      style="min-width: 44px; padding: 0.625rem 1rem;"
      :style="{ opacity: currentPage === 1 ? 0.5 : 1, cursor: currentPage === 1 ? 'not-allowed' : 'pointer' }"
    >
      ←
    </button>

    <!-- Page Numbers -->
    <div style="display: flex; gap: 0.375rem;">
      <button
        v-for="page in visiblePages"
        :key="page"
        @click="page !== '...' && $emit('page-change', page)"
        :class="page === currentPage ? 'btn-primary' : 'btn-secondary'"
        class="btn"
        style="min-width: 44px; padding: 0.625rem 1rem;"
        :disabled="page === '...'"
        :style="{ cursor: page === '...' ? 'default' : 'pointer' }"
      >
        {{ page }}
      </button>
    </div>

    <!-- Next Button -->
    <button
      @click="$emit('page-change', currentPage + 1)"
      :disabled="currentPage === totalPages"
      class="btn btn-secondary"
      style="min-width: 44px; padding: 0.625rem 1rem;"
      :style="{ opacity: currentPage === totalPages ? 0.5 : 1, cursor: currentPage === totalPages ? 'not-allowed' : 'pointer' }"
    >
      →
    </button>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  currentPage: number
  totalPages: number
}>()

defineEmits<{
  'page-change': [page: number]
}>()

const visiblePages = computed(() => {
  const pages: (number | string)[] = []
  const maxVisible = 5
  
  if (props.totalPages <= maxVisible) {
    // Tüm sayfaları göster
    for (let i = 1; i <= props.totalPages; i++) {
      pages.push(i)
    }
  } else {
    // İlk sayfa
    pages.push(1)
    
    // Orta kısım
    if (props.currentPage > 3) {
      pages.push('...')
    }
    
    const start = Math.max(2, props.currentPage - 1)
    const end = Math.min(props.totalPages - 1, props.currentPage + 1)
    
    for (let i = start; i <= end; i++) {
      if (i !== 1 && i !== props.totalPages) {
        pages.push(i)
      }
    }
    
    // Son kısım
    if (props.currentPage < props.totalPages - 2) {
      pages.push('...')
    }
    
    // Son sayfa
    pages.push(props.totalPages)
  }
  
  return pages
})
</script>
