<template>
  <div class="page-container">
    <div class="page-header">
      <a-button type="text" @click="router.back()">
        <ArrowLeftOutlined /> 返回
      </a-button>
      <h1 class="page-title">攻略详情</h1>
    </div>
    <a-card>
      <a-empty description="攻略详情页开发中">
        <template #image>
          <CompassOutlined style="font-size: 64px; color: #bfbfbf;" />
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
