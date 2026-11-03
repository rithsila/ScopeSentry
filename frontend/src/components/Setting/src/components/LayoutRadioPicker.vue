<script setup lang="ts">
import { useAppStore } from '@/store/modules/app'
import { computed } from 'vue'
import { useDesign } from '@/hooks/web/useDesign'

const { getPrefixCls } = useDesign()

const prefixCls = getPrefixCls('layout-radio-picker')

const appStore = useAppStore()

const layout = computed(() => appStore.getLayout)
</script>

<template>
  <div :class="prefixCls" class="grid grid-cols-4 gap-2">
    <div class="flex flex-col items-center">
      <div
        :class="[
          `${prefixCls}__classic`,
          'relative w-full h-44px cursor-pointer bg-[#050c05] border rounded-2px transition-all',
          {
            'is-active': layout === 'classic'
          }
        ]"
        @click="appStore.setLayout('classic')"
      ></div>
      <span
        class="text-[10px] font-mono mt-1.5 transition-colors"
        :class="layout === 'classic' ? 'text-[#00ff41] font-700' : 'text-[#556855]'"
      >
        Classic
      </span>
    </div>

    <div class="flex flex-col items-center">
      <div
        :class="[
          `${prefixCls}__top-left`,
          'relative w-full h-44px cursor-pointer bg-[#050c05] border rounded-2px transition-all',
          {
            'is-active': layout === 'topLeft'
          }
        ]"
        @click="appStore.setLayout('topLeft')"
      ></div>
      <span
        class="text-[10px] font-mono mt-1.5 transition-colors"
        :class="layout === 'topLeft' ? 'text-[#00ff41] font-700' : 'text-[#556855]'"
      >
        TopLeft
      </span>
    </div>

    <div class="flex flex-col items-center">
      <div
        :class="[
          `${prefixCls}__top`,
          'relative w-full h-44px cursor-pointer bg-[#050c05] border rounded-2px transition-all',
          {
            'is-active': layout === 'top'
          }
        ]"
        @click="appStore.setLayout('top')"
      ></div>
      <span
        class="text-[10px] font-mono mt-1.5 transition-colors"
        :class="layout === 'top' ? 'text-[#00ff41] font-700' : 'text-[#556855]'"
      >
        Top
      </span>
    </div>

    <div class="flex flex-col items-center">
      <div
        :class="[
          `${prefixCls}__cut-menu`,
          'relative w-full h-44px cursor-pointer bg-[#050c05] border rounded-2px transition-all',
          {
            'is-active': layout === 'cutMenu'
          }
        ]"
        @click="appStore.setLayout('cutMenu')"
      >
        <div class="absolute h-full w-[20%] top-0 left-[15%] bg-[#132a13] z-2"></div>
      </div>
      <span
        class="text-[10px] font-mono mt-1.5 transition-colors"
        :class="layout === 'cutMenu' ? 'text-[#00ff41] font-700' : 'text-[#556855]'"
      >
        CutMenu
      </span>
    </div>
  </div>
</template>

<style lang="less" scoped>
@prefix-cls: ~'@{namespace}-layout-radio-picker';

.@{prefix-cls} {
  &__classic {
    border-color: #1b3d1b;

    &::before {
      position: absolute;
      top: 0;
      left: 0;
      z-index: 2;
      width: 32%;
      height: 100%;
      background-color: #1b3d1b;
      border-right: 1px solid #245224;
      content: '';
    }

    &::after {
      position: absolute;
      top: 0;
      left: 0;
      width: 100%;
      height: 28%;
      background-color: #0e1b0e;
      border-bottom: 1px solid #1b3d1b;
      content: '';
    }
  }

  &__top-left {
    border-color: #1b3d1b;

    &::before {
      position: absolute;
      top: 0;
      left: 0;
      z-index: 2;
      width: 100%;
      height: 28%;
      background-color: #132a13;
      border-bottom: 1px solid #245224;
      content: '';
    }

    &::after {
      position: absolute;
      top: 0;
      left: 0;
      width: 32%;
      height: 100%;
      background-color: #1b3d1b;
      border-right: 1px solid #245224;
      content: '';
    }
  }

  &__top {
    border-color: #1b3d1b;

    &::before {
      position: absolute;
      top: 0;
      left: 0;
      z-index: 2;
      width: 100%;
      height: 28%;
      background-color: #1b3d1b;
      border-bottom: 1px solid #00ff41;
      content: '';
    }
  }

  &__cut-menu {
    border-color: #1b3d1b;

    &::before {
      position: absolute;
      top: 0;
      left: 0;
      z-index: 1;
      width: 100%;
      height: 28%;
      background-color: #0e1b0e;
      border-bottom: 1px solid #1b3d1b;
      content: '';
    }

    &::after {
      position: absolute;
      top: 0;
      left: 0;
      z-index: 3;
      width: 15%;
      height: 100%;
      background-color: #1b3d1b;
      border-right: 1px solid #245224;
      content: '';
    }
  }

  .is-active,
  .is-acitve {
    border-color: #00ff41 !important;
    box-shadow: 0 0 10px rgba(0, 255, 65, 0.45) !important;
  }
}
</style>
