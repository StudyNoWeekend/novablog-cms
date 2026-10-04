<template>
  <div class="role-picker">
    <div v-for="group in ROLE_GROUPS" :key="group.title" class="role-group">
      <div class="role-group__title">{{ group.title }}</div>
      <div class="role-grid">
        <button
          v-for="role in group.roles"
          :key="role.key"
          type="button"
          class="role-card"
          :class="{ selected: isSelected(role.key) }"
          :aria-pressed="isSelected(role.key)"
          @click="toggleRole(role.key)"
        >
          <span class="role-card__icon"><component :is="roleIconMap[role.key]" /></span>
          <span class="role-card__label">{{ role.label }}</span>
          <span class="role-card__desc">{{ role.description }}</span>
          <span v-if="isSelected(role.key)" class="role-card__check"><CheckOutlined /></span>
        </button>
      </div>
    </div>
    <div class="role-grid role-grid--wide">
      <button
        v-for="role in SPECIAL_ROLES"
        :key="role.key"
        type="button"
        class="role-card"
        :class="{ selected: isSelected(role.key) }"
        :aria-pressed="isSelected(role.key)"
        @click="toggleRole(role.key)"
      >
        <span class="role-card__icon"><component :is="roleIconMap[role.key]" /></span>
        <span class="role-card__label">{{ role.label }}</span>
        <span class="role-card__desc">{{ role.description }}</span>
        <span v-if="isSelected(role.key)" class="role-card__check"><CheckOutlined /></span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  AppstoreOutlined,
  BgColorsOutlined,
  BulbOutlined,
  CameraOutlined,
  CheckOutlined,
  CodeOutlined,
  CoffeeOutlined,
  CompassOutlined,
  CustomerServiceOutlined,
  HeartOutlined,
  LaptopOutlined,
  PlaySquareOutlined,
  ReadOutlined,
  RocketOutlined,
  SkinOutlined,
  SmileOutlined,
  ToolOutlined,
  VideoCameraOutlined,
} from '@ant-design/icons-vue'
import { ROLE_GROUPS, SPECIAL_ROLES, type RoleKey } from '@/constants/setupRoles'

const props = defineProps<{
  /** 已选角色 key 集合（v-model:selected） */
  selected: RoleKey[]
}>()

const emit = defineEmits<{
  (e: 'update:selected', value: RoleKey[]): void
}>()

const roleIconMap: Record<RoleKey, unknown> = {
  tech: CodeOutlined,
  digital: LaptopOutlined,
  photo: CameraOutlined,
  video: PlaySquareOutlined,
  music: CustomerServiceOutlined,
  travelvlog: VideoCameraOutlined,
  travel: CompassOutlined,
  food: CoffeeOutlined,
  fitness: HeartOutlined,
  fashion: SkinOutlined,
  pet: SmileOutlined,
  reading: ReadOutlined,
  gaming: RocketOutlined,
  designer: BgColorsOutlined,
  craft: ToolOutlined,
  lifestyle: BulbOutlined,
  all: AppstoreOutlined,
}

function isSelected(key: RoleKey): boolean {
  return props.selected.includes(key)
}

function toggleRole(key: RoleKey) {
  emit(
    'update:selected',
    isSelected(key) ? props.selected.filter((k) => k !== key) : [...props.selected, key],
  )
}
</script>

<style scoped>
.role-picker {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.role-group__title {
  font-size: 12px;
  font-weight: 500;
  color: #94a3b8;
  margin-bottom: 6px;
}

.role-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
}

.role-grid--wide {
  grid-template-columns: repeat(2, 1fr);
}

.role-card {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 2px;
  padding: 10px 12px;
  min-height: 44px;
  background: #f8f9fb;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  cursor: pointer;
  text-align: left;
  transition: border-color 0.2s, background 0.2s, box-shadow 0.2s;
}

.role-card:hover {
  border-color: #4a6cf7;
}

.role-card:focus-visible {
  outline: 2px solid #4a6cf7;
  outline-offset: 2px;
}

.role-card.selected {
  background: rgba(74, 108, 247, 0.06);
  border-color: #4a6cf7;
  box-shadow: 0 0 0 1px #4a6cf7 inset;
}

.role-card__icon {
  font-size: 18px;
  color: #64748b;
  line-height: 1.2;
}

.role-card.selected .role-card__icon {
  color: #4a6cf7;
}

.role-card__label {
  font-size: 13px;
  font-weight: 500;
  color: #1e293b;
}

.role-card__desc {
  font-size: 11px;
  color: #94a3b8;
  line-height: 1.4;
}

.role-card__check {
  position: absolute;
  top: 6px;
  right: 6px;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: #4a6cf7;
  color: #fff;
  font-size: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
