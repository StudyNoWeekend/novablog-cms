<template>
  <a-modal
    :open="visible"
    :title="category ? $t('category.edit') : $t('category.add')"
    :confirm-loading="submitting"
    @ok="handleSubmit"
    @cancel="handleCancel"
  >
    <a-form layout="vertical">
      <a-form-item :label="$t('category.form.name')" required>
        <a-input v-model:value="form.name" :placeholder="$t('category.form.namePlaceholder')" :maxlength="50" show-count />
      </a-form-item>
      <a-form-item :label="$t('common.field.description')">
        <a-input v-model:value="form.description" :placeholder="$t('category.form.descriptionPlaceholder')" />
      </a-form-item>
      <a-form-item :label="$t('common.field.sort')">
        <a-input-number
          v-model:value="form.sort_order"
          :min="0"
          :precision="0"
          style="width: 100%"
          :placeholder="$t('category.form.sortPlaceholder')"
        />
      </a-form-item>
    </a-form>
  </a-modal>
</template>
<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { message } from 'ant-design-vue'
import { useI18n } from 'vue-i18n'
import { categoryApi } from '@/api/category'
import type { Category } from '@/types/category'

const { t } = useI18n()

const props = defineProps<{
  /** 分类类型（categories.type），新增时写入该 type */
  type: string
  /** 编辑时传入分类，新增传 null */
  category?: Category | null
  visible: boolean
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'saved'): void
}>()

const submitting = ref(false)
const form = reactive({
  name: '',
  description: '',
  sort_order: 0,
})

watch(
  () => props.visible,
  (visible) => {
    if (visible) {
      form.name = props.category?.name || ''
      form.description = props.category?.description || ''
      form.sort_order = props.category?.sort_order ?? 0
    }
  },
)

function handleCancel() {
  emit('update:visible', false)
}

async function handleSubmit() {
  if (!form.name.trim()) {
    message.warning(t('category.form.nameRequired'))
    return
  }
  submitting.value = true
  try {
    const payload = {
      name: form.name.trim(),
      description: form.description || undefined,
      sort_order: form.sort_order || 0,
    }
    if (props.category) {
      await categoryApi.update(props.category.id, payload)
      message.success(t('category.message.updateSuccess'))
    } else {
      await categoryApi.create({ ...payload, type: props.type })
      message.success(t('category.message.createSuccess'))
    }
    emit('update:visible', false)
    emit('saved')
  } catch {
    // 错误由拦截器处理
  } finally {
    submitting.value = false
  }
}
</script>
