<template>
  <div class="page-container">
    <div class="page-header">
      <a-button type="text" @click="router.back()">
        <ArrowLeftOutlined /> {{ $t('common.actions.back') }}
      </a-button>
      <h1 class="page-title">{{ $t('travel.detail.title') }}</h1>
    </div>
    <a-card>
      <a-empty :description="$t('travel.detail.developing')">
        <template #image>
          <CompassOutlined style="font-size: 64px; color: #8A93A8;" />
        </template>
      </a-empty>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { ArrowLeftOutlined, CompassOutlined } from '@ant-design/icons-vue'

const router = useRouter()
</script>
