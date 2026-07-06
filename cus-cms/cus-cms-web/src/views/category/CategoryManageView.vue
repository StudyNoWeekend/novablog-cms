<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">分类标签管理</h1>
    </div>

    <a-row :gutter="24">
      <!-- 分类管理 -->
      <a-col :span="12">
        <a-card title="分类管理" :bordered="false">
          <template #extra>
            <a-button type="primary" size="small" @click="showAddCategory = true">
              <PlusOutlined /> 添加分类
            </a-button>
          </template>

          <a-table
            :columns="categoryColumns"
            :data-source="categories"
            :loading="categoriesLoading"
            :pagination="false"
            row-key="id"
            size="small"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'sort_order'">
                {{ record.sort_order || 0 }}
              </template>
              <template v-if="column.key === 'action'">
                <a-button type="text" size="small" @click="handleEditCategory(record)">
                  <EditOutlined />
                </a-button>
                <a-popconfirm
                  title="确定要删除该分类吗？"
                  ok-text="删除"
                  cancel-text="取消"
                  ok-type="danger"
                  @confirm="handleDeleteCategory(record.id)"
                >
                  <a-button type="text" size="small" danger>
                    <DeleteOutlined />
                  </a-button>
                </a-popconfirm>
              </template>
            </template>
          </a-table>
        </a-card>
      </a-col>

      <!-- 标签管理 -->
      <a-col :span="12">
        <a-card title="标签管理" :bordered="false">
          <template #extra>
            <a-button type="primary" size="small" @click="showAddTag = true">
              <PlusOutlined /> 添加标签
            </a-button>
          </template>

          <a-space wrap>
            <a-tag
              v-for="tag in tags"
              :key="tag.id"
              closable
              @close="handleDeleteTag(tag.id)"
              color="blue"
            >
              {{ tag.name }}
            </a-tag>
          </a-space>

          <a-empty v-if="!tagsLoading && tags.length === 0" description="暂无标签" />
        </a-card>
      </a-col>
    </a-row>

    <!-- 编辑分类弹窗 -->
    <a-modal
      v-model:open="showAddCategory"
      :title="editingCategory ? '编辑分类' : '添加分类'"
      @ok="handleCategorySubmit"
      @cancel="handleCancelCategory"
    >
      <a-form :model="categoryForm" layout="vertical">
        <a-form-item label="分类名称" required>
          <a-input v-model:value="categoryForm.name" placeholder="请输入分类名称" />
        </a-form-item>
        <a-form-item label="Slug">
          <a-input v-model:value="categoryForm.slug" placeholder="URL 别名（留空自动生成）" />
        </a-form-item>
        <a-form-item label="描述">
          <a-input v-model:value="categoryForm.description" placeholder="分类描述" />
        </a-form-item>
        <a-form-item label="排序">
          <a-input-number v-model:value="categoryForm.sort_order" :min="0" style="width: 100%" />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 添加标签弹窗 -->
    <a-modal
      v-model:open="showAddTag"
      title="添加标签"
      @ok="handleTagSubmit"
      @cancel="showAddTag = false"
    >
      <a-form :model="tagForm" layout="vertical">
        <a-form-item label="标签名称" required>
          <a-input v-model:value="tagForm.name" placeholder="请输入标签名称" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import { categoryApi } from '@/api/category'
import { tagApi } from '@/api/tag'
import type { Category } from '@/types/category'
import type { Tag } from '@/types/tag'

const categories = ref<Category[]>([])
const tags = ref<Tag[]>([])
const categoriesLoading = ref(false)
const tagsLoading = ref(false)

// 分类弹窗
const showAddCategory = ref(false)
const editingCategory = ref<Category | null>(null)
const categoryForm = reactive({
  name: '',
  slug: '',
  description: '',
  sort_order: 0,
})

// 标签弹窗
const showAddTag = ref(false)
const tagForm = reactive({ name: '' })

const categoryColumns = [
  { title: '名称', dataIndex: 'name', key: 'name' },
  { title: 'Slug', dataIndex: 'slug', key: 'slug' },
  { title: '排序', dataIndex: 'sort_order', key: 'sort_order' },
  { title: '操作', key: 'action', width: 100 },
]

onMounted(() => {
  loadCategories()
  loadTags()
})

async function loadCategories() {
  categoriesLoading.value = true
  try { categories.value = await categoryApi.getList() as unknown as Category[] }
  catch { message.error('加载分类失败') }
  finally { categoriesLoading.value = false }
}

async function loadTags() {
  tagsLoading.value = true
  try { tags.value = await tagApi.getList() as unknown as Tag[] }
  catch { message.error('加载标签失败') }
  finally { tagsLoading.value = false }
}

// 分类操作
function handleEditCategory(cat: Category) {
  editingCategory.value = cat
  categoryForm.name = cat.name
  categoryForm.slug = cat.slug || ''
  categoryForm.description = cat.description || ''
  categoryForm.sort_order = cat.sort_order || 0
  showAddCategory.value = true
}

function handleCancelCategory() {
  showAddCategory.value = false
  editingCategory.value = null
  categoryForm.name = ''
  categoryForm.slug = ''
  categoryForm.description = ''
  categoryForm.sort_order = 0
}

async function handleCategorySubmit() {
  if (!categoryForm.name.trim()) {
    message.warning('请输入分类名称')
    return
  }
  try {
    if (editingCategory.value) {
      await categoryApi.update(editingCategory.value.id, {
        name: categoryForm.name,
        slug: categoryForm.slug || undefined,
        description: categoryForm.description || undefined,
        sort_order: categoryForm.sort_order,
      })
      message.success('分类已更新')
    } else {
      await categoryApi.create({
        name: categoryForm.name,
        slug: categoryForm.slug || undefined,
        description: categoryForm.description || undefined,
        sort_order: categoryForm.sort_order,
      })
      message.success('分类已创建')
    }
    handleCancelCategory()
    loadCategories()
  } catch {}
}

async function handleDeleteCategory(id: string) {
  try {
    await categoryApi.remove(id)
    message.success('分类已删除')
    loadCategories()
  } catch {}
}

// 标签操作
async function handleTagSubmit() {
  if (!tagForm.name.trim()) {
    message.warning('请输入标签名称')
    return
  }
  try {
    await tagApi.create({ name: tagForm.name })
    message.success('标签已创建')
    showAddTag.value = false
    tagForm.name = ''
    loadTags()
  } catch {}
}

async function handleDeleteTag(id: string) {
  try {
    await tagApi.remove(id)
    message.success('标签已删除')
    loadTags()
  } catch {}
}
</script>