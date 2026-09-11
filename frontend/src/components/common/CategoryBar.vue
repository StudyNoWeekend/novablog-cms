<template>
  <div class="category-bar">
    <div class="category-chips">
      <span
        class="category-chip"
        :class="{ active: !selectedId }"
        @click="handleSelect(undefined)"
      >
        全部
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
            <a-menu-item key="edit">编辑</a-menu-item>
            <a-menu-item key="delete">删除</a-menu-item>
          </a-menu>
        </template>
      </a-dropdown>
      <a-button type="dashed" size="small" @click="showAddModal" class="add-btn">
        <PlusOutlined /> 新增分类
      </a-button>
    </div>

    <!-- Add/Edit Category Modal -->
    <a-modal
      v-model:open="modalVisible"
      :title="editingCategory ? '编辑分类' : '新增分类'"
      :confirm-loading="submitting"
      @ok="handleSubmit"
      @cancel="modalVisible = false"
    >
      <a-form layout="vertical">
        <a-form-item label="分类名称" required>
          <a-input v-model:value="form.name" placeholder="请输入分类名称" :maxlength="50" show-count />
        </a-form-item>
        <a-form-item label="描述">
          <a-input v-model:value="form.description" placeholder="分类描述（可选）" />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- Delete Confirmation -->
    <a-modal
      v-model:open="deleteVisible"
      title="确认删除"
      ok-text="删除"
      ok-type="danger"
      cancel-text="取消"
      @ok="handleDelete"
    >
      <p>确定要删除分类「{{ deletingCategory?.name }}」吗？</p>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import { categoryApi } from '@/api/category'
import type { Category } from '@/types/category'

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
const submitting = ref(false)
const editingCategory = ref<Category | null>(null)
const form = reactive({
  name: '',
  description: '',
})

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
  form.name = ''
  form.description = ''
  modalVisible.value = true
}

function handleMenuClick(e: { key: string }, cat: Category) {
  if (e.key === 'edit') {
    editingCategory.value = cat
    form.name = cat.name
    form.description = cat.description || ''
    modalVisible.value = true
  } else if (e.key === 'delete') {
    deletingCategory.value = cat
    deleteVisible.value = true
  }
}

async function handleSubmit() {
  if (!form.name.trim()) {
    message.warning('请输入分类名称')
    return
  }
  submitting.value = true
  try {
    if (editingCategory.value) {
      await categoryApi.update(editingCategory.value.id, {
        name: form.name.trim(),
        description: form.description || undefined,
      })
      message.success('分类已更新')
    } else {
      await categoryApi.create({
        name: form.name.trim(),
        description: form.description || undefined,
        type: props.type,
      })
      message.success('分类已创建')
    }
    modalVisible.value = false
    loadCategories()
  } catch {
    // error handled by interceptor
  } finally {
    submitting.value = false
  }
}

async function handleDelete() {
  if (!deletingCategory.value) return
  try {
    await categoryApi.remove(deletingCategory.value.id)
    message.success('分类已删除')
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
  border: 1px solid var(--border-color, #e8e8e8);
  color: var(--text-secondary, #64748b);
  transition: all 0.2s;
  user-select: none;
  white-space: nowrap;
}

.category-chip:hover {
  border-color: var(--primary, #4a6cf7);
  color: var(--primary, #4a6cf7);
}

.category-chip.active {
  background: var(--primary, #4a6cf7);
  border-color: var(--primary, #4a6cf7);
  color: #fff;
}

.add-btn {
  flex-shrink: 0;
  cursor: pointer;
}
</style>
