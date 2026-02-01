<template>
  <div class="min-h-screen" style="background: linear-gradient(to bottom, #f5f5f7 0%, #ffffff 100%);">
    <!-- Header -->
    <header style="padding: 3rem 0 2rem;">
      <div class="container">
        <h1 style="text-align: center; margin-bottom: 0.5rem;">
          Arama Motoru
        </h1>
        <p style="text-align: center; font-size: 1.125rem; color: var(--color-text-secondary);">
          Hızlı ve akıllı içerik keşfi
        </p>
      </div>
    </header>

    <!-- Search Section -->
    <section style="padding: 2rem 0;">
      <div class="container">
        <!-- Search Bar -->
        <div style="max-width: 720px; margin: 0 auto 2rem;">
          <div style="position: relative;">
            <input
              v-model="searchQuery"
              type="search"
              placeholder="Arama yapın..."
              @keyup.enter="handleSearch"
              style="padding-left: 3rem; font-size: 1.125rem;"
            />
            <svg
              style="position: absolute; left: 1rem; top: 50%; transform: translateY(-50%); width: 20px; height: 20px; color: var(--color-text-secondary);"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
          </div>
        </div>

        <!-- Filters -->
        <div style="max-width: 720px; margin: 0 auto 3rem; display: flex; gap: 1rem; flex-wrap: wrap; align-items: center;">
          <select v-model="filters.type" @change="handleSearch" style="flex: 1; min-width: 150px;">
            <option value="">Tüm İçerikler</option>
            <option value="video">Video</option>
            <option value="article">Makale</option>
          </select>
          
          <select v-model="filters.sort" @change="handleSearch" style="flex: 1; min-width: 150px;">
            <option value="popularity">Popülerlik</option>
            <option value="relevance">Alakalılık</option>
          </select>
        </div>

        <!-- Loading -->
        <div v-if="loading" style="display: flex; justify-content: center; padding: 4rem 0;">
          <div class="spinner"></div>
        </div>

        <!-- Error -->
        <div v-else-if="error" style="max-width: 720px; margin: 0 auto; padding: 1.5rem; background: #fff5f5; border-radius: var(--radius-lg); color: #c53030;">
          <p style="margin: 0; color: #c53030;">{{ error }}</p>
        </div>

        <!-- Results -->
        <div v-else-if="results">
          <!-- Results Count -->
          <div style="max-width: 720px; margin: 0 auto 1.5rem; color: var(--color-text-secondary);">
            <p style="margin: 0;">{{ results.pagination.total_items }} sonuç bulundu</p>
          </div>

          <!-- Content Cards -->
          <div style="display: grid; gap: 1.5rem; max-width: 720px; margin: 0 auto;">
            <ContentCard
              v-for="item in results.items"
              :key="item.id"
              :content="item"
              class="fade-in"
            />
          </div>

          <!-- Pagination -->
          <Pagination
            v-if="results.pagination.total_pages > 1"
            :current-page="results.pagination.page"
            :total-pages="results.pagination.total_pages"
            @page-change="handlePageChange"
            style="margin-top: 3rem;"
          />
        </div>

        <!-- Empty State -->
        <div v-else style="text-align: center; padding: 4rem 0;">
          <svg style="width: 80px; height: 80px; margin: 0 auto 1.5rem; color: var(--color-text-secondary); opacity: 0.3;" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <h3 style="color: var(--color-text-secondary); font-weight: 500;">Aramaya başlayın</h3>
          <p style="color: var(--color-text-secondary);">İlgilendiğiniz konuları keşfedin</p>
        </div>
      </div>
    </section>

    <!-- Footer -->
    <footer style="padding: 3rem 0; text-align: center; color: var(--color-text-secondary); font-size: 0.875rem;">
      <div class="container">
        <p style="margin: 0;">© 2026 Onur ERDOĞAN.</p>
      </div>
    </footer>

    <!-- Test Tools Trigger -->
    <button class="test-tools-trigger" @click="showTestTools = true">
      🛠️ Test Araçları
    </button>

    <!-- Test Tools Modal -->
    <TestToolsModal 
      :show="showTestTools" 
      @close="showTestTools = false" 
      @synced="handleSearch"
    />
  </div>
</template>

<script setup lang="ts">
const searchQuery = ref('')
const filters = reactive({
  type: '',
  sort: 'popularity',
  page: 1
})

const { search, loading, error } = useSearch()
const results = ref<any>(null)
const showTestTools = ref(false)

// İlk yüklemede boş arama ile verileri getir
onMounted(async () => {
    try {
        results.value = await search({
            query: '',
            page: 1,
            pageSize: 10
        })
    } catch (e) {
        console.error('Initial search error:', e)
    }
})

const handleSearch = async () => {
    // Arama kutusu boşsa ve sonuç varsa aramayı engelleme (hepsini getirsin)
    // Ama gereksiz istekleri önlemek için yine de trim kontrolü yapalım
    const query = searchQuery.value.trim()
    
    filters.page = 1
    try {
        results.value = await search({
            query: query,
            type: filters.type,
            sort: filters.sort,
            page: filters.page,
            pageSize: 10
        })
    } catch (e) {
        // Error handled by composable
    }
}

const handlePageChange = async (page: number) => {
  filters.page = page
  try {
    results.value = await search({
      query: searchQuery.value,
      type: filters.type,
      sort: filters.sort,
      page: filters.page,
      pageSize: 10
    })
    window.scrollTo({ top: 0, behavior: 'smooth' })
  } catch (e) {
    // Error handled by composable
  }
}
</script>

<style scoped>
.test-tools-trigger {
  position: fixed;
  bottom: 2rem;
  right: 2rem;
  background: #333;
  color: white;
  padding: 0.75rem 1.25rem;
  border-radius: 999px;
  border: none;
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  cursor: pointer;
  z-index: 100;
  font-weight: 500;
  transition: transform 0.2s;
}

.test-tools-trigger:hover {
  transform: scale(1.05);
  background: #444;
}
</style>
