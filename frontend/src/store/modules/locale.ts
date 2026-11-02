import { defineStore } from 'pinia'
import { store } from '../index'
import en from 'element-plus/es/locale/lang/en'
import { useStorage } from '@/hooks/web/useStorage'
import { LocaleDropdownType } from '@/components/LocaleDropdown'

const { setStorage } = useStorage()

const elLocaleMap = {
  en: en
}
interface LocaleState {
  currentLocale: LocaleDropdownType
  localeMap: LocaleDropdownType[]
}

export const useLocaleStore = defineStore('locales', {
  state: (): LocaleState => {
    return {
      currentLocale: {
        lang: 'en',
        elLocale: elLocaleMap.en
      },
      localeMap: [
        {
          lang: 'en',
          name: 'English'
        }
      ]
    }
  },
  getters: {
    getCurrentLocale(): LocaleDropdownType {
      return this.currentLocale
    },
    getLocaleMap(): LocaleDropdownType[] {
      return this.localeMap
    }
  },
  actions: {
    setCurrentLocale(localeMap: LocaleDropdownType) {
      this.currentLocale.lang = 'en'
      this.currentLocale.elLocale = elLocaleMap.en
      setStorage('lang', 'en')
    }
  }
})

export const useLocaleStoreWithOut = () => {
  return useLocaleStore(store)
}
