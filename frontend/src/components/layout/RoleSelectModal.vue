<template>
  <a-modal
    v-model:open="visible"
    :title="modalTitle"
    :width="760"
    :confirm-loading="saving"
    ok-text="保存"
    :cancel-text="hasRole ? '取消' : '稍后再选'"
    :ok-button-props="{ disabled: selected.length === 0 }"
    @ok="handleSave"
  >
    <p class="role-modal__desc">{{ modalDesc }}</p>

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
import { ALL_ROLES, computeModulePreset, type RoleKey } from '@/constants/setupRoles'

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

// 是否已选过创作方向：已选（编辑模式）与未选（首装跳过/旧版本用户）用不同标题与文案
const hasRole = computed(() => !!authStore.user?.role)

const modalTitle = computed(() => (hasRole.value ? '修改创作方向' : '选择你的创作方向'))

const modalDesc = computed(() =>
  hasRole.value
    ? '重新选择你的创作方向（可多选），保存后系统将按新方向更新左侧菜单模块，之后可在「模块管理」中随时调整。'
    : '检测到你还没有选择创作方向。选择爱好（可多选）后，系统将按方向定制左侧菜单模块，之后可在「模块管理」中随时调整。',
)

// 解析服务端逗号分隔的 role 为合法角色 key 数组（未选过或含失效 key 时自动过滤）
function parseUserRole(role?: string): RoleKey[] {
  if (!role) return []
  return role
    .split(',')
    .map((key) => key.trim())
    .filter((key): key is RoleKey => ALL_ROLES.some((r) => r.key === key))
}

// 打开时预填当前已选方向：编辑模式带上现有选择，未选过则为空
watch(
  () => props.open,
  (open) => {
    if (open) selected.value = parseUserRole(authStore.user?.role)
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
  color: #667085;
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
  background: #F5F7FC;
  border: 1px solid #E5E9F2;
  border-radius: 8px;
}

.role-preset__title {
  font-size: 13px;
  font-weight: 500;
  color: #29365C;
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
  background: rgba(82, 111, 232, 0.08);
  color: #526FE8;
  transition: opacity 0.2s;
}

.module-chip--off {
  background: #F5F7FC;
  color: #8A93A8;
}

.module-chip__check {
  font-size: 10px;
}

.module-chip__badge {
  font-style: normal;
  font-size: 10px;
  color: #8A93A8;
}
</style>
