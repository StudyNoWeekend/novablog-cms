<template>
  <div class="category-sidebar">
    <div class="sidebar-header">
      <span class="sidebar-title">分类</span>
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
        <span class="category-name">全部</span>
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
              <a-menu-item @click="handleEdit(cat)">编辑</a-menu-item>
              <a-menu-item @click="handleDelete(cat)">删除</a-menu-item>
            </a-menu>
          </template>
        </a-dropdown>
      </li>
    </ul>

    <!-- 添加/编辑分类弹窗 -->
    <a-modal
      v-model:open="showAddModal"
      :title="editingCategory ? '编辑分类' : '新增分类'"
      :confirm-loading="saving"
      @ok="handleSave"
    >
      <a-form layout="vertical">
        <a-form-item label="分类名称" required>
          <a-input v-model:value="formState.name" placeholder="请输入分类名称" />
        </a-form-item>
        <a-form-item label="排序">
          <a-input-number v-model:value="formState.sort_order" :min="0" style="width: 100%" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { PlusOutlined, MoreOutlined } from '@ant-design/icons-vue'
import { message, Modal } from 'ant-design-vue'
import { categoryApi } from '@/api/category'
import type { Category, CategoryCreateReq } from '@/types/category'

defineProps<{
  categories: Category[]
  modelValue: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'refresh'): void
}>()

const showAddModal = ref(false)
const saving = ref(false)
const editingCategory = ref<Category | null>(null)

const formState = reactive<CategoryCreateReq>({
  name: '',
  sort_order: 0,
})

function openAddModal() {
  editingCategory.value = null
  formState.name = ''
  formState.slug = undefined
  formState.sort_order = 0
  showAddModal.value = true
}

function handleEdit(cat: Category) {
  editingCategory.value = cat
  formState.name = cat.name
  formState.slug = cat.slug
  formState.sort_order = cat.sort_order
  showAddModal.value = true
}

function handleDelete(cat: Category) {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除分类「${cat.name}」吗？`,
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    onOk: async () => {
      try {
        await categoryApi.remove(cat.id)
        message.success('删除成功')
        emit('refresh')
      } catch {
        // 错误由拦截器处理
      }
    },
  })
}

async function handleSave() {
  if (!formState.name.trim()) {
    message.warning('请输入分类名称')
    return
  }
  saving.value = true
  try {
    if (editingCategory.value) {
      await categoryApi.update(editingCategory.value.id, {
        name: formState.name,
        slug: formState.slug,
        sort_order: formState.sort_order,
      })
      message.success('更新成功')
    } else {
      await categoryApi.create({
        name: formState.name,
        slug: formState.slug,
        sort_order: formState.sort_order,
        type: 'music',
      })
      message.success('添加成功')
    }
    showAddModal.value = false
    emit('refresh')
  } catch {
    // 错误由拦截器处理
  } finally {
    saving.value = false
  }
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
