<template>
  <div class="category-bar">
    <div class="category-chips">
      <span
        class="category-chip"
        :class="{ active: !selectedId }"
        @click="handleSelect(undefined)"
      >
        {{ $t('common.status.all') }}
      </span>
      <a-dropdown
        v-for="cat in categories"
        :key="cat.id"
        :trigger="['contextmenu']"
      >
        <span
          class="category-chip"
          :class="{ active: selectedId === cat.id }"
          @click="handleSelect(cat.id)"
        >
          {{ cat.name }}
        </span>
        <template #overlay>
          <a-menu @click="handleMenuClick($event, cat)">
            <a-menu-item key="edit">{{ $t('common.actions.edit') }}</a-menu-item>
            <a-menu-item key="delete">{{ $t('common.actions.delete') }}</a-menu-item>
          </a-menu>
        </template>
      </a-dropdown>
      <a-button type="dashed" size="small" @click="showAddModal" class="add-btn">
        <PlusOutlined /> {{ $t('category.add') }}
      </a-button>
    </div>

    <!-- Add/Edit Category Modal -->
    <CategoryEditModal
      v-model:visible="modalVisible"
      :type="type"
      :category="editingCategory"
      @saved="loadCategories"
    />

    <!-- Delete Confirmation -->
    <a-modal
      v-model:open="deleteVisible"
      :title="$t('category.message.deleteTitle')"
      :ok-text="$t('common.actions.delete')"
      ok-type="danger"
      :cancel-text="$t('common.actions.cancel')"
      @ok="handleDelete"
    >
      <p>{{ $t('category.message.deleteConfirm', { name: deletingCategory?.name }) }}</p>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import { useI18n } from 'vue-i18n'
import { categoryApi } from '@/api/category'
import type { Category } from '@/types/category'
import CategoryEditModal from '@/components/common/CategoryEditModal.vue'

const { t } = useI18n()

const props = defineProps<{
  type: string
}>()

const emit = defineEmits<{
  (e: 'select', value: string | undefined): void
}>()

const categories = ref<Category[]>([])
const selectedId = ref<string | undefined>(undefined)
const loading = ref(false)

// Modal state
const modalVisible = ref(false)
const editingCategory = ref<Category | null>(null)

// Delete state
const deleteVisible = ref(false)
const deletingCategory = ref<Category | null>(null)

onMounted(() => {
  loadCategories()
})

async function loadCategories() {
  loading.value = true
  try {
    categories.value = await categoryApi.getList(props.type)
  } catch {
    // error handled by interceptor
  } finally {
    loading.value = false
  }
}

function handleSelect(id: string | undefined) {
  selectedId.value = id
  emit('select', id)
}

function showAddModal() {
  editingCategory.value = null
  modalVisible.value = true
}

function handleMenuClick(e: { key: string }, cat: Category) {
  if (e.key === 'edit') {
    editingCategory.value = cat
    modalVisible.value = true
  } else if (e.key === 'delete') {
    deletingCategory.value = cat
    deleteVisible.value = true
  }
}

async function handleDelete() {
  if (!deletingCategory.value) return
  try {
    await categoryApi.remove(deletingCategory.value.id)
    message.success(t('category.message.deleteSuccess'))
    if (selectedId.value === deletingCategory.value.id) {
      handleSelect(undefined)
    }
    deleteVisible.value = false
    loadCategories()
  } catch {
    // error handled by interceptor
  }
}
</script>

<style scoped>
.category-bar {
  margin-bottom: 16px;
}

.category-chips {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.category-chip {
  display: inline-flex;
  align-items: center;
  padding: 4px 14px;
  border-radius: 16px;
  font-size: 13px;
  cursor: pointer;
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-color, #E5E9F2);
  color: var(--text-secondary, #667085);
  transition: all 0.2s;
  user-select: none;
  white-space: nowrap;
}

.category-chip:hover {
  border-color: var(--primary, #526FE8);
  color: var(--primary, #526FE8);
}

.category-chip.active {
  background: var(--primary, #526FE8);
  border-color: var(--primary, #526FE8);
  color: #fff;
}

.add-btn {
  flex-shrink: 0;
  cursor: pointer;
}
</style>
