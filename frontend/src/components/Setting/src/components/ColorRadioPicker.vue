<script setup lang="ts">
import { PropType, watch, unref, ref } from 'vue'
import { propTypes } from '@/utils/propTypes'
import { useDesign } from '@/hooks/web/useDesign'
import { colorIsDark } from '@/utils/color'

const { getPrefixCls } = useDesign()

const prefixCls = getPrefixCls('color-radio-picker')

const props = defineProps({
  schema: {
    type: Array as PropType<string[]>,
    default: () => []
  },
  modelValue: propTypes.string.def('')
})

const emit = defineEmits(['update:modelValue', 'change'])

const colorVal = ref(props.modelValue)

watch(
  () => props.modelValue,
  (val: string) => {
    if (val === unref(colorVal)) return
    colorVal.value = val
  }
)

// 监听
watch(
  () => colorVal.value,
  (val: string) => {
    emit('update:modelValue', val)
    emit('change', val)
  }
)
</script>

<template>
  <div :class="prefixCls" class="flex flex-wrap gap-2.5 justify-center">
    <span
      v-for="(item, i) in schema"
      :key="`radio-${i}`"
      class="w-22px h-22px cursor-pointer rounded-2px border border-[#1b3d1b] flex items-center justify-center transition-all hover:scale-110"
      :class="{ 'is-active': colorVal === item }"
      :style="{
        background: item
      }"
      @click="colorVal = item"
    >
      <Icon
        v-if="colorVal === item"
        :color="colorIsDark(item) ? '#00ff41' : '#050c05'"
        icon="ep:check"
        :size="14"
      />
    </span>
  </div>
</template>

<style lang="less" scoped>
@prefix-cls: ~'@{namespace}-color-radio-picker';

.@{prefix-cls} {
  .is-active {
    border: 2px solid #ffffff !important;
    box-shadow: 0 0 8px rgba(0, 255, 65, 0.6) !important;
    transform: scale(1.1);
  }
}
</style>
