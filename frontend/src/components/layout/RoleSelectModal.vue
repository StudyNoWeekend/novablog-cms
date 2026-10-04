<template>
  <a-modal
    v-model:open="visible"
    title="选择你的创作方向"
    :width="760"
    :confirm-loading="saving"
    ok-text="保存"
    cancel-text="稍后再选"
    :ok-button-props="{ disabled: selected.length === 0 }"
    @ok="handleSave"
  >
    <p class="role-modal__desc">
      检测到你还没有选择创作方向。选择爱好（可多选）后，系统将按方向定制左侧菜单模块，之后可在「模块管理」中随时调整。
    </p>

    <div class="role-modal__picker">
      <RolePickerCards v-model:selected="selected" />
    </div>

    <div class="role-preset">
      <div class="role-preset__title">将为你开启的模块</div>
      <div class="role-preset__chips">
        <span
          v-for="item in moduleChips"
          :key="item.key"
          class="module-chip"
          :class="{ 'module-chip--off': !item.enabled }"
        >
          <CheckOutlined v-if="item.enabled" class="module-chip__check" />
          {{ item.label }}
          <em v-if="item.common" class="module-chip__badge">通用</em>
        </span>
      </div>
    </div>
  </a-modal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { CheckOutlined } from '@ant-design/icons-vue'
import RolePickerCards from '@/components/common/RolePickerCards.vue'
import { authApi } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'
import { useModuleStore } from '@/stores/module'
import { pathModuleMap } from '@/router/guards'
import { MODULE_DEFS, COMMON_MODULE_KEYS } from '@/constants/modules'
import { computeModulePreset, type RoleKey } from '@/constants/setupRoles'

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
}>()

const visible = computed({
  get: () => props.open,
  set: (val: boolean) => emit('update:open', val),
})

const router = useRouter()
const authStore = useAuthStore()
const moduleStore = useModuleStore()

const selected = ref<RoleKey[]>([])
const saving = ref(false)

// 每次打开重置选择（弹窗仅在服务端 role 为空时出现，不记忆上次关闭时的选择）
watch(
  () => props.open,
  (open) => {
    if (open) selected.value = []
  },
)

const presetModules = computed(() => computeModulePreset(selected.value))

const moduleChips = computed(() =>
  MODULE_DEFS.map((def) => ({
    key: def.key,
    label: def.label,
    common: COMMON_MODULE_KEYS.includes(def.key),
    enabled: presetModules.value[def.key],
  })),
)

/** 保存补选：更新 role 并全量应用模块预设，侧边栏菜单即时刷新 */
async function handleSave() {
  if (selected.value.length === 0) return
  saving.value = true
  try {
    const modules = computeModulePreset(selected.value)
    const res = await authApi.updateRoles({
      roles: selected.value,
      modules: { ...modules },
    })

    // 同步本地用户信息，本会话不再弹窗（关闭不持久化，刷新后以服务端 role 为准）
    if (authStore.user) {
      authStore.user = { ...authStore.user, role: res.role || selected.value.join(',') }
    }

    // 强刷模块配置，AppSidebar 的菜单 computed 会自动重算
    await moduleStore.fetchConfig(true)
    message.success('已按创作方向更新功能模块')
    visible.value = false

    // 当前页面所属模块被关闭时回到工作台
    const moduleKey = pathModuleMap[router.currentRoute.value.path.split('/')[1]]
    if (moduleKey && !moduleStore.isEnabled(moduleKey)) {
      router.replace('/dashboard')
    }
  } catch (e: any) {
    message.error(e?.message || '保存失败，请重试')
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.role-modal__desc {
  font-size: 13px;
  color: #64748b;
  line-height: 1.6;
  margin: 0 0 12px;
}

.role-modal__picker {
  max-height: 340px;
  overflow-y: auto;
  padding-right: 4px;
}

.role-preset {
  margin-top: 14px;
  padding: 12px 14px;
  background: #f8f9fb;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
}

.role-preset__title {
  font-size: 13px;
  font-weight: 500;
  color: #1e293b;
  margin-bottom: 8px;
}

.role-preset__chips {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.module-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 10px;
  font-size: 12px;
  border-radius: 999px;
  background: rgba(74, 108, 247, 0.08);
  color: #4a6cf7;
  transition: opacity 0.2s;
}

.module-chip--off {
  background: #eef1f5;
  color: #94a3b8;
}

.module-chip__check {
  font-size: 10px;
}

.module-chip__badge {
  font-style: normal;
  font-size: 10px;
  color: #94a3b8;
}
</style>
