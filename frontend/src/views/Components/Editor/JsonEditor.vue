<script setup lang="ts">
import { ContentWrap } from '@/components/ContentWrap'
import { JsonEditor } from '@/components/JsonEditor'
import { useI18n } from '@/hooks/web/useI18n'
import { ref, watch } from 'vue'

const { t } = useI18n()

const defaultData = ref({
  title: 'Title',
  content: 'Content'
})

watch(
  () => defaultData.value,
  (val) => {
    console.log(val)
  },
  {
    deep: true
  }
)

setTimeout(() => {
  defaultData.value = {
    title: 'Async Title',
    content: 'Async Content'
  }
}, 4000)
</script>

<template>
  <ContentWrap :title="t('richText.jsonEditor')" :message="t('richText.jsonEditorDes')">
    <JsonEditor v-model="defaultData" />
  </ContentWrap>
</template>
