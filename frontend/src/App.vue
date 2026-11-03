<script setup lang="ts">
import { computed } from 'vue'
import { useAppStore } from '@/store/modules/app'
import { ConfigGlobal } from '@/components/ConfigGlobal'
import { useDesign } from '@/hooks/web/useDesign'
import { useStorage } from '@/hooks/web/useStorage'
import { getCssVar, setCssVar } from './utils'
import { isDark } from '@/utils/is'

const { getPrefixCls } = useDesign()

const prefixCls = getPrefixCls('app')

const appStore = useAppStore()

const currentSize = computed(() => appStore.getCurrentSize)

const greyMode = computed(() => appStore.getGreyMode)

const { getStorage } = useStorage('localStorage')

// Set default system theme based on Dark Phosphor Terminal
const setDefaultTheme = () => {
  // Force reset cached theme tokens in localStorage
  try {
    const rawApp = localStorage.getItem('app')
    if (rawApp) {
      const parsed = JSON.parse(rawApp)
      if (parsed) {
        if (!parsed.theme) parsed.theme = {}
        parsed.theme.elColorPrimary = '#00ff41'
        parsed.theme.leftMenuBgColor = '#0a140a'
        parsed.theme.leftMenuBorderColor = '#1b3d1b'
        parsed.theme.leftMenuBgActiveColor = '#00ff41'
        parsed.theme.leftMenuCollapseBgActiveColor = '#00ff41'
        parsed.theme.leftMenuTextActiveColor = '#050c05'
        parsed.theme.leftMenuTextColor = '#8a9e8a'
        localStorage.setItem('app', JSON.stringify(parsed))
      }
    }
    const rawTheme = localStorage.getItem('theme')
    if (rawTheme) {
      const parsedTheme = JSON.parse(rawTheme)
      if (parsedTheme) {
        parsedTheme.elColorPrimary = '#00ff41'
        localStorage.setItem('theme', JSON.stringify(parsedTheme))
      }
    }
  } catch (e) {
    console.error('Failed to reset cached theme tokens in localStorage', e)
  }

  const isDarkValue = getStorage('isDark') !== null ? getStorage('isDark') : true
  appStore.setIsDark(isDarkValue)
  const color = getCssVar('--el-bg-color') || '#0a140a'
  appStore.setMenuTheme(color)
  appStore.setHeaderTheme(color)

  // Force reset cached theme tokens on startup
  appStore.setTheme({
    elColorPrimary: '#00ff41',
    leftMenuBgColor: '#0a140a',
    leftMenuBorderColor: '#1b3d1b',
    leftMenuBgActiveColor: '#00ff41',
    leftMenuCollapseBgActiveColor: '#00ff41',
    leftMenuTextColor: '#8a9e8a',
    leftMenuTextActiveColor: '#050c05'
  })
  appStore.setCssVarTheme()
  setCssVar('--el-color-primary', '#00ff41')

  // Initialize tags view height
  const tagsViewEnabled = appStore.getTagsView
  setCssVar('--tags-view-height', tagsViewEnabled ? '35px' : '0px')
}

setDefaultTheme()
</script>

<template>
  <ConfigGlobal :size="currentSize">
    <RouterView :class="greyMode ? `${prefixCls}-grey-mode` : ''" />
  </ConfigGlobal>
</template>

<style lang="less">
@prefix-cls: ~'@{namespace}-app';

.size {
  width: 100%;
  height: 100%;
}

html,
body {
  padding: 0 !important;
  margin: 0;
  overflow: hidden;
  .size;

  #app {
    .size;
  }
}

.@{prefix-cls}-grey-mode {
  filter: grayscale(100%);
}
</style>
