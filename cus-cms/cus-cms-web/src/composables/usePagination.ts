import { ref, computed } from 'vue'

export function usePagination(defaultPageSize = 20) {
  const page = ref(1)
  const pageSize = ref(defaultPageSize)
  const total = ref(0)

  const totalPages = computed(() => {
    if (total.value === 0) return 1
    return Math.ceil(total.value / pageSize.value)
  })

  function handlePageChange(newPage: number, newPageSize?: number) {
    page.value = newPage
    if (newPageSize && newPageSize !== pageSize.value) {
      pageSize.value = newPageSize
      page.value = 1
    }
  }

  function reset() {
    page.value = 1
  }

  return {
    page, pageSize, total, totalPages,
    handlePageChange, reset,
  }
}