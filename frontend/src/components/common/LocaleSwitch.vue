<template>
  <a-dropdown placement="bottomRight" :trigger="['click']">
    <button
      type="button"
      class="locale-switch"
      :title="$t('common.actions.language')"
      :aria-label="$t('common.actions.language')"
    >
      <GlobalOutlined />
      <span class="locale-name">{{ LOCALE_META[localeStore.locale].name }}</span>
    </button>
    <template #overlay>
      <a-menu :selected-keys="[localeStore.locale]" @click="handleSelect">
        <a-menu-item v-for="item in SUPPORTED_LOCALES" :key="item">
          <span class="locale-option">
            <span>{{ LOCALE_META[item].name }}</span>
            <CheckOutlined v-if="item === localeStore.locale" class="locale-check" />
          </span>
        </a-menu-item>
      </a-menu>
    </template>
  </a-dropdown>
</template>

<script setup lang="ts">
import { GlobalOutlined, CheckOutlined } from '@ant-design/icons-vue'
import { SUPPORTED_LOCALES, LOCALE_META, isSupportedLocale } from '@/locales'
import { useLocaleStore } from '@/stores/locale'

const localeStore = useLocaleStore()

function handleSelect({ key }: { key: string | number }) {
  if (isSupportedLocale(key)) localeStore.setLocale(key)
}
</script>

<style scoped>
.locale-switch {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 20px;
  color: #667085;
  cursor: pointer;
  padding: 4px;
  background: none;
  border: none;
  transition: color var(--transition-fast);
}

.locale-switch:hover {
  color: #526FE8;
}

.locale-name {
  font-size: 13px;
  line-height: 1;
}

.locale-option {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-width: 96px;
}

.locale-check {
  color: #526FE8;
  font-size: 12px;
}

@media (max-width: 768px) {
  .locale-name {
    display: none;
  }
}
</style>
