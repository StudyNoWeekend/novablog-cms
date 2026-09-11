import { ref, onScopeDispose, type Ref } from 'vue'

export function useDebounce<T>(value: Ref<T>, delay = 300) {
  const debouncedValue = ref(value.value) as Ref<T>
  let timer: ReturnType<typeof setTimeout>

  onScopeDispose(() => {
    if (timer) clearTimeout(timer)
  })

  return {
    debouncedValue,
    setDebounce(val: T) {
      clearTimeout(timer)
      timer = setTimeout(() => {
        debouncedValue.value = val
      }, delay)
    },
  }
}