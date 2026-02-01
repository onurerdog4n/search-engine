// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },
  modules: ['@nuxt/content', '@nuxtjs/i18n'],
  i18n: {
    locales: [
      { code: 'tr', file: 'tr.json', name: 'Türkçe' }
    ],
    langDir: 'locales',
    defaultLocale: 'tr',
    strategy: 'no_prefix'
  },
  app: {
    baseURL: '/docs_search/'
  }
})
