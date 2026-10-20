<template>
  <div class="category-sidebar">
    <div class="sidebar-header">
      <span class="sidebar-title">{{ $t('common.field.category') }}</span>
      <a-button type="text" size="small" @click="openAddModal">
        <PlusOutlined />
      </a-button>
    </div>
    <ul class="category-list">
      <li
        class="category-item"
        :class="{ active: modelValue === '' }"
        @click="emit('update:modelValue', '')"
      >
        <span class="category-name">{{ $t('common.status.all') }}</span>
      </li>
      <li
        v-for="cat in categories"
        :key="cat.id"
        class="category-item"
        :class="{ active: modelValue === cat.id }"
        @click="emit('update:modelValue', cat.id)"
      >
        <span class="category-name">{{ cat.name }}</span>
        <a-dropdown :trigger="['click']">
          <MoreOutlined class="category-more" @click.stop />
          <template #overlay>
            <a-menu>
              <a-menu-item @click="handleEdit(cat)">{{ $t('common.actions.edit') }}</a-menu-item>
              <a-menu-item @click="handleDelete(cat)">{{ $t('common.actions.delete') }}</a-menu-item>
            </a-menu>
          </template>
        </a-dropdown>
      </li>
    </ul>

    <!-- 添加/编辑分类弹窗（与其他模块共用同一弹窗） -->
    <CategoryEditModal
      v-model:visible="modalVisible"
      type="music"
      :category="editingCategory"
      @saved="emit('refresh')"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { PlusOutlined, MoreOutlined } from '@ant-design/icons-vue'
import { message, Modal } from 'ant-design-vue'
import { categoryApi } from '@/api/category'
import type { Category } from '@/types/category'
import CategoryEditModal from '@/components/common/CategoryEditModal.vue'

defineProps<{
  categories: Category[]
  modelValue: string
}>()

const { t } = useI18n()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'refresh'): void
}>()

const modalVisible = ref(false)
const editingCategory = ref<Category | null>(null)

function openAddModal() {
  editingCategory.value = null
  modalVisible.value = true
}

function handleEdit(cat: Category) {
  editingCategory.value = cat
  modalVisible.value = true
}

function handleDelete(cat: Category) {
  Modal.confirm({
    title: t('common.message.deleteConfirm'),
    content: t('common.message.deleteConfirmNamed', { name: cat.name }),
    okText: t('common.actions.delete'),
    okType: 'danger',
    cancelText: t('common.actions.cancel'),
    onOk: async () => {
      try {
        await categoryApi.remove(cat.id)
        message.success(t('common.message.deleteSuccess'))
        emit('refresh')
      } catch {
        // 错误由拦截器处理
      }
    },
  })
}
</script>

<style scoped>
.category-sidebar {
  width: 224px;
  flex-shrink: 0;
  background: var(--bg-card);
  border-radius: var(--border-radius-lg);
  box-shadow: var(--shadow-card);
  overflow: hidden;
}

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  border-bottom: 1px solid var(--border-color);
}

.sidebar-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.category-list {
  list-style: none;
  margin: 0;
  padding: 8px;
  max-height: calc(100vh - 260px);
  overflow-y: auto;
}

.category-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  border-radius: var(--border-radius);
  cursor: pointer;
  transition: all var(--transition-fast);
  color: var(--text-primary);
}

.category-item:hover {
  background: var(--color-primary-light);
}

.category-item.active {
  background: var(--color-primary);
  color: #fff;
}

.category-item.active .category-more {
  color: #fff;
}

.category-name {
  font-size: 14px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.category-more {
  font-size: 16px;
  color: var(--text-tertiary);
  padding: 2px;
  border-radius: 4px;
  cursor: pointer;
  transition: color var(--transition-fast);
}

.category-more:hover {
  color: var(--text-primary);
}
</style>
