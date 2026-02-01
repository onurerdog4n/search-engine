// API client composable - Nuxt auto-imports ref, readonly, useRuntimeConfig, $fetch
export const useSearch = () => {
    const config = useRuntimeConfig()
    const loading = ref(false)
    const error = ref<string | null>(null)

    const search = async (params: {
        query: string
        type?: string
        sort?: string
        page?: number
        pageSize?: number
    }) => {
        loading.value = true
        error.value = null

        try {
            const response = await $fetch(`${config.public.apiBase}/api/v1/search`, {
                params: {
                    query: params.query,
                    type: params.type || undefined,
                    sort: params.sort || 'popularity',
                    page: params.page || 1,
                    page_size: params.pageSize || 20
                }
            })
            return response
        } catch (e: any) {
            error.value = e.data?.error || e.message || 'Bir hata oluştu'
            throw e
        } finally {
            loading.value = false
        }
    }

    return {
        search,
        loading: readonly(loading),
        error: readonly(error)
    }
}
