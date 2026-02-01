<template>
  <main>
    <ContentRenderer v-if="page" :value="page" />
    <div v-else>Page not found</div>
  </main>
</template>

<script setup>
const route = useRoute()
const { locale } = useI18n()

// Fetch page data with better error handling and path normalization
const { data: page } = await useAsyncData('page-' + route.path, () => {
  let path = route.path

  // Nuxt Content v3 handling for i18n:
  // Content is organized in 'content/tr/...' and 'content/en/...'
  // We need to prepend the locale to the path
  
  // Add locale prefix if not already present
  if (!path.startsWith(`/${locale.value}`)) {
    path = `/${locale.value}${path === '/' ? '' : path}`
  }

  // Sanitize trailing slash
  if (path.length > 1 && path.endsWith('/')) {
    path = path.slice(0, -1)
  }

  console.log('Fetching content for path:', path)
  return queryCollection('docs').path(path).first()
})
</script>
