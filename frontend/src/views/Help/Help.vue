<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { Icon } from '@/components/Icon'
import {
  ElCard,
  ElRow,
  ElCol,
  ElInput,
  ElTag,
  ElButton,
  ElTabs,
  ElTabPane
} from 'element-plus'

const router = useRouter()
const activeTab = ref('all')
const searchQuery = ref('')

interface MenuItemGuide {
  id: string
  name: string
  icon: string
  category: 'core' | 'scanners' | 'system'
  path: string
  what: string
  how: string
}

const menuList: MenuItemGuide[] = [
  {
    id: 'dashboard',
    name: 'Dashboard',
    icon: 'ant-design:dashboard-filled',
    category: 'core',
    path: '/dashboard/index',
    what: 'The main home screen.',
    how: 'Look here to see total assets, active scan jobs, and system health.'
  },
  {
    id: 'asset',
    name: 'Asset Information',
    icon: 'carbon:view',
    category: 'core',
    path: '/asset-information/index',
    what: 'All discovered targets.',
    how: 'Search and filter subdomains, IP addresses, open ports, and URLs.'
  },
  {
    id: 'task',
    name: 'Task',
    icon: 'bi:list-task',
    category: 'scanners',
    path: '/task-management/ScanTask',
    what: 'Scan engine control center.',
    how: 'Create new scans, pause jobs, or set auto-run schedules.'
  },
  {
    id: 'plugins',
    name: 'Plugins Manager',
    icon: 'mingcute:plugin-2-fill',
    category: 'scanners',
    path: '/plugin-management/index',
    what: 'Scanning modules and scripts.',
    how: 'Enable, disable, or install community tools for port scan and web crawler.'
  },
  {
    id: 'node',
    name: 'Node',
    icon: 'material-symbols:network-node',
    category: 'system',
    path: '/node-management/index',
    what: 'Distributed scanning workers.',
    how: 'View node CPU, memory, connection status, and copy the worker connect key.'
  },
  {
    id: 'project',
    name: 'Project',
    icon: 'eos-icons:project-outlined',
    category: 'core',
    path: '/project-management/index',
    what: 'Target boundary definition.',
    how: 'Group targets by company or program. Define what is in-scope and out-of-scope.'
  },
  {
    id: 'poc',
    name: 'POC',
    icon: 'ant-design:bug-filled',
    category: 'scanners',
    path: '/poc-management/index',
    what: 'Vulnerability check templates.',
    how: 'Run security tests to confirm real bugs on found assets.'
  },
  {
    id: 'fingerprint',
    name: 'Fingerprint',
    icon: 'material-symbols:fingerprint',
    category: 'scanners',
    path: '/fingerprint-management/index',
    what: 'Technology identification rules.',
    how: 'Detect web frameworks, CMS, CDN, and server versions automatically.'
  },
  {
    id: 'sensitive',
    name: 'Sensitive Information',
    icon: 'carbon:deploy-rules',
    category: 'scanners',
    path: '/sensitive-information-rules/index',
    what: 'Leak and secret scanner.',
    how: 'Find exposed API keys, passwords, and tokens inside JavaScript and web pages.'
  },
  {
    id: 'dictionary',
    name: 'Dictionary',
    icon: 'material-symbols:dictionary',
    category: 'system',
    path: '/dictionary-management/manage',
    what: 'Wordlists for brute-forcing.',
    how: 'Upload your own subdomain lists, directory paths, and port lists.'
  },
  {
    id: 'configuration',
    name: 'Configuration',
    icon: 'tdesign:system-setting',
    category: 'system',
    path: '/configuration/system',
    what: 'System settings and API keys.',
    how: 'Add third-party keys (FOFA, Hunter, Shodan) and set alert bots (DingTalk, Telegram).'
  },
  {
    id: 'about',
    name: 'About',
    icon: 'carbon:deploy-rules',
    category: 'system',
    path: '/about/index',
    what: 'App details.',
    how: 'Check current version and update links.'
  }
]

const filteredMenus = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  if (!query) {
    return menuList
  }
  return menuList.filter((item) => {
    return (
      item.name.toLowerCase().includes(query) ||
      item.what.toLowerCase().includes(query) ||
      item.how.toLowerCase().includes(query) ||
      item.path.toLowerCase().includes(query)
    )
  })
})

const navigateTo = (path: string) => {
  router.push(path)
}
</script>

<template>
  <div class="help-page p-5 bg-[#050c05] text-[#d4d4d4] font-mono min-h-full">
    <!-- Terminal Header Banner -->
    <div
      class="header-banner mb-6 p-4 rounded border border-[#1b3d1b] bg-[#0a140a] shadow-[0_0_15px_rgba(0,255,65,0.1)]"
    >
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div class="flex items-center space-x-3">
          <div class="w-3 h-3 rounded-full bg-[#00ff41] shadow-[0_0_10px_#00ff41] animate-pulse"></div>
          <div>
            <h1 class="text-xl font-bold tracking-wider text-[#00ff41] m-0 flex items-center gap-2">
              <Icon icon="solar:help-bold" :size="24" class="text-[#00ff41]" />
              SCOPE_SENTRY // HELP_CENTER
            </h1>
            <p class="text-xs text-[#8a9e8a] m-0 mt-1">
              Beginner manual, menu guide, and recommended scanning steps.
            </p>
          </div>
        </div>
        <div class="flex items-center gap-2">
          <ElTag effect="plain" class="!bg-[#0e1b0e] !border-[#1b3d1b] !text-[#00ff41] !font-mono">
            STATUS: ONLINE
          </ElTag>
          <ElTag effect="plain" class="!bg-[#0e1b0e] !border-[#1b3d1b] !text-[#8a9e8a] !font-mono">
            MODE: PHOSPHOR_V2
          </ElTag>
        </div>
      </div>
    </div>

    <!-- Navigation Tabs -->
    <div class="mb-6">
      <ElTabs v-model="activeTab" class="terminal-tabs">
        <ElTabPane label="[ 00 // OVERVIEW ]" name="all" />
        <ElTabPane label="[ 01 // QUICK START ]" name="quickstart" />
        <ElTabPane label="[ 02 // ALL MENUS ]" name="menus" />
        <ElTabPane label="[ 03 // LEARNING PATH ]" name="learning" />
      </ElTabs>
    </div>

    <!-- SECTION 1: QUICK START -->
    <section
      v-if="activeTab === 'all' || activeTab === 'quickstart'"
      class="mb-8"
    >
      <div class="flex items-center justify-between mb-4 pb-2 border-b border-[#1b3d1b]">
        <div class="flex items-center gap-2">
          <span class="text-[#00ff41] font-bold text-base">SECTION 1:</span>
          <span class="text-white font-bold text-base">Quick Start (Step 1 to End)</span>
        </div>
        <span class="text-xs text-[#8a9e8a]">4 Simple Steps for Beginners</span>
      </div>

      <ElRow :gutter="16">
        <!-- Step 1 -->
        <ElCol :xs="24" :sm="12" :lg="6" class="mb-4">
          <div
            class="step-card h-full p-4 rounded border border-[#1b3d1b] bg-[#0a140a] hover:border-[#00ff41] transition-all duration-300 flex flex-col justify-between"
          >
            <div>
              <div class="flex items-center justify-between mb-2">
                <span class="step-badge px-2 py-0.5 text-xs font-bold rounded bg-[#1b3d1b] text-[#00ff41]">
                  STEP 01
                </span>
                <Icon icon="material-symbols:network-node" :size="20" class="text-[#00ff41]" />
              </div>
              <h3 class="text-sm font-bold text-[#00ff41] mb-2">Check Node</h3>
              <div class="text-xs text-[#8a9e8a] mb-2">Target Menu: Node Management</div>
              <p class="text-xs text-[#d4d4d4] leading-relaxed mb-4">
                Make sure at least one scan node is online with a green dot.
              </p>
            </div>
            <ElButton
              size="small"
              class="w-full !bg-[#0e1b0e] !border-[#1b3d1b] !text-[#00ff41] hover:!border-[#00ff41] !font-mono text-xs"
              @click="navigateTo('/node-management/index')"
            >
              Open Node Page &rarr;
            </ElButton>
          </div>
        </ElCol>

        <!-- Step 2 -->
        <ElCol :xs="24" :sm="12" :lg="6" class="mb-4">
          <div
            class="step-card h-full p-4 rounded border border-[#1b3d1b] bg-[#0a140a] hover:border-[#00ff41] transition-all duration-300 flex flex-col justify-between"
          >
            <div>
              <div class="flex items-center justify-between mb-2">
                <span class="step-badge px-2 py-0.5 text-xs font-bold rounded bg-[#1b3d1b] text-[#00ff41]">
                  STEP 02
                </span>
                <Icon icon="eos-icons:project-outlined" :size="20" class="text-[#00ff41]" />
              </div>
              <h3 class="text-sm font-bold text-[#00ff41] mb-2">Create Project</h3>
              <div class="text-xs text-[#8a9e8a] mb-2">Target Menu: Project</div>
              <p class="text-xs text-[#d4d4d4] leading-relaxed mb-4">
                Add your target domain (example: example.com) or IP range.
              </p>
            </div>
            <ElButton
              size="small"
              class="w-full !bg-[#0e1b0e] !border-[#1b3d1b] !text-[#00ff41] hover:!border-[#00ff41] !font-mono text-xs"
              @click="navigateTo('/project-management/index')"
            >
              Open Project Page &rarr;
            </ElButton>
          </div>
        </ElCol>

        <!-- Step 3 -->
        <ElCol :xs="24" :sm="12" :lg="6" class="mb-4">
          <div
            class="step-card h-full p-4 rounded border border-[#1b3d1b] bg-[#0a140a] hover:border-[#00ff41] transition-all duration-300 flex flex-col justify-between"
          >
            <div>
              <div class="flex items-center justify-between mb-2">
                <span class="step-badge px-2 py-0.5 text-xs font-bold rounded bg-[#1b3d1b] text-[#00ff41]">
                  STEP 03
                </span>
                <Icon icon="bi:list-task" :size="20" class="text-[#00ff41]" />
              </div>
              <h3 class="text-sm font-bold text-[#00ff41] mb-2">Launch Task</h3>
              <div class="text-xs text-[#8a9e8a] mb-2">Target Menu: Task</div>
              <p class="text-xs text-[#d4d4d4] leading-relaxed mb-4">
                Start a scan job. Pick what to run: Subdomains, Ports, Web titles, or Vulns.
              </p>
            </div>
            <ElButton
              size="small"
              class="w-full !bg-[#0e1b0e] !border-[#1b3d1b] !text-[#00ff41] hover:!border-[#00ff41] !font-mono text-xs"
              @click="navigateTo('/task-management/ScanTask')"
            >
              Open Task Page &rarr;
            </ElButton>
          </div>
        </ElCol>

        <!-- Step 4 -->
        <ElCol :xs="24" :sm="12" :lg="6" class="mb-4">
          <div
            class="step-card h-full p-4 rounded border border-[#1b3d1b] bg-[#0a140a] hover:border-[#00ff41] transition-all duration-300 flex flex-col justify-between"
          >
            <div>
              <div class="flex items-center justify-between mb-2">
                <span class="step-badge px-2 py-0.5 text-xs font-bold rounded bg-[#1b3d1b] text-[#00ff41]">
                  STEP 04
                </span>
                <Icon icon="carbon:view" :size="20" class="text-[#00ff41]" />
              </div>
              <h3 class="text-sm font-bold text-[#00ff41] mb-2">Review Findings</h3>
              <div class="text-xs text-[#8a9e8a] mb-2">Target Menu: Asset Information & POC</div>
              <p class="text-xs text-[#d4d4d4] leading-relaxed mb-4">
                Check discovered subdomains, open ports, web apps, and bugs found.
              </p>
            </div>
            <ElButton
              size="small"
              class="w-full !bg-[#0e1b0e] !border-[#1b3d1b] !text-[#00ff41] hover:!border-[#00ff41] !font-mono text-xs"
              @click="navigateTo('/asset-information/index')"
            >
              View Assets Page &rarr;
            </ElButton>
          </div>
        </ElCol>
      </ElRow>
    </section>

    <!-- SECTION 2: EVERY MENU EXPLAINED -->
    <section
      v-if="activeTab === 'all' || activeTab === 'menus'"
      class="mb-8"
    >
      <div class="flex flex-wrap items-center justify-between gap-3 mb-4 pb-2 border-b border-[#1b3d1b]">
        <div class="flex items-center gap-2">
          <span class="text-[#00ff41] font-bold text-base">SECTION 2:</span>
          <span class="text-white font-bold text-base">Every Menu Explained</span>
        </div>
        <div class="w-full sm:w-72">
          <ElInput
            v-model="searchQuery"
            placeholder="Search menu name, keyword, or purpose..."
            clearable
            class="search-input !font-mono text-xs"
          >
            <template #prefix>
              <Icon icon="ant-design:search-outlined" class="text-[#00ff41]" />
            </template>
          </ElInput>
        </div>
      </div>

      <div v-if="filteredMenus.length === 0" class="p-8 text-center border border-[#1b3d1b] rounded bg-[#0a140a]">
        <p class="text-[#8a9e8a] text-sm m-0">No menus matched "{{ searchQuery }}". Try another word.</p>
      </div>

      <ElRow :gutter="16">
        <ElCol
          v-for="item in filteredMenus"
          :key="item.id"
          :xs="24"
          :sm="12"
          :lg="8"
          class="mb-4"
        >
          <div
            class="menu-card h-full p-4 rounded border border-[#1b3d1b] bg-[#0a140a] hover:border-[#00ff41] transition-all duration-300 flex flex-col justify-between"
          >
            <div>
              <div class="flex items-center justify-between mb-3">
                <div class="flex items-center space-x-2">
                  <div class="p-1.5 rounded bg-[#0e1b0e] border border-[#1b3d1b]">
                    <Icon :icon="item.icon" :size="18" class="text-[#00ff41]" />
                  </div>
                  <span class="text-sm font-bold text-[#00ff41]">{{ item.name }}</span>
                </div>
                <span class="text-[10px] text-[#8a9e8a] font-mono px-1.5 py-0.5 rounded bg-[#0e1b0e] border border-[#1b3d1b]">
                  {{ item.path }}
                </span>
              </div>

              <!-- What -->
              <div class="mb-2">
                <div class="text-[11px] text-[#00ff41] font-bold uppercase tracking-wider mb-0.5">
                  [ What ]
                </div>
                <div class="text-xs text-[#d4d4d4] leading-relaxed pl-2 border-l border-[#1b3d1b]">
                  {{ item.what }}
                </div>
              </div>

              <!-- How -->
              <div class="mb-4">
                <div class="text-[11px] text-[#33ff67] font-bold uppercase tracking-wider mb-0.5">
                  [ How ]
                </div>
                <div class="text-xs text-[#8a9e8a] leading-relaxed pl-2 border-l border-[#1b3d1b]">
                  {{ item.how }}
                </div>
              </div>
            </div>

            <div class="pt-2 border-t border-[#122412] flex justify-end">
              <ElButton
                link
                size="small"
                class="!text-[#00ff41] hover:!text-[#33ff67] !font-mono text-xs !p-0"
                @click="navigateTo(item.path)"
              >
                Go to {{ item.name }} &rarr;
              </ElButton>
            </div>
          </div>
        </ElCol>
      </ElRow>
    </section>

    <!-- SECTION 3: RECOMMENDED LEARNING PATH -->
    <section
      v-if="activeTab === 'all' || activeTab === 'learning'"
      class="mb-6"
    >
      <div class="flex items-center justify-between mb-4 pb-2 border-b border-[#1b3d1b]">
        <div class="flex items-center gap-2">
          <span class="text-[#00ff41] font-bold text-base">SECTION 3:</span>
          <span class="text-white font-bold text-base">Recommended Learning Path</span>
        </div>
        <span class="text-xs text-[#8a9e8a]">Simple Advice for New Users</span>
      </div>

      <ElRow :gutter="16">
        <ElCol :xs="24" :sm="12" :lg="6" class="mb-4">
          <div class="h-full p-4 rounded border border-[#1b3d1b] bg-[#0a140a] hover:border-[#00ff41] transition-all">
            <div class="flex items-center space-x-2 mb-2">
              <span class="w-6 h-6 rounded-full bg-[#1b3d1b] text-[#00ff41] text-xs font-bold flex items-center justify-center">
                1
              </span>
              <span class="text-xs font-bold text-[#00ff41]">Single Domain First</span>
            </div>
            <p class="text-xs text-[#d4d4d4] leading-relaxed m-0">
              Start with single domain scans first before testing big IP blocks.
            </p>
          </div>
        </ElCol>

        <ElCol :xs="24" :sm="12" :lg="6" class="mb-4">
          <div class="h-full p-4 rounded border border-[#1b3d1b] bg-[#0a140a] hover:border-[#00ff41] transition-all">
            <div class="flex items-center space-x-2 mb-2">
              <span class="w-6 h-6 rounded-full bg-[#1b3d1b] text-[#00ff41] text-xs font-bold flex items-center justify-center">
                2
              </span>
              <span class="text-xs font-bold text-[#00ff41]">Keep Dictionaries Small</span>
            </div>
            <p class="text-xs text-[#d4d4d4] leading-relaxed m-0">
              Keep dictionaries small on day one to finish scans quickly.
            </p>
          </div>
        </ElCol>

        <ElCol :xs="24" :sm="12" :lg="6" class="mb-4">
          <div class="h-full p-4 rounded border border-[#1b3d1b] bg-[#0a140a] hover:border-[#00ff41] transition-all">
            <div class="flex items-center space-x-2 mb-2">
              <span class="w-6 h-6 rounded-full bg-[#1b3d1b] text-[#00ff41] text-xs font-bold flex items-center justify-center">
                3
              </span>
              <span class="text-xs font-bold text-[#00ff41]">Add API Keys Early</span>
            </div>
            <p class="text-xs text-[#d4d4d4] leading-relaxed m-0">
              Add third-party API keys in Configuration to speed up passive discovery.
            </p>
          </div>
        </ElCol>

        <ElCol :xs="24" :sm="12" :lg="6" class="mb-4">
          <div class="h-full p-4 rounded border border-[#1b3d1b] bg-[#0a140a] hover:border-[#00ff41] transition-all">
            <div class="flex items-center space-x-2 mb-2">
              <span class="w-6 h-6 rounded-full bg-[#1b3d1b] text-[#00ff41] text-xs font-bold flex items-center justify-center">
                4
              </span>
              <span class="text-xs font-bold text-[#00ff41]">Inspect Before POC</span>
            </div>
            <p class="text-xs text-[#d4d4d4] leading-relaxed m-0">
              Inspect discovered URLs before launching deep POC testing.
            </p>
          </div>
        </ElCol>
      </ElRow>

      <!-- Terminal Tip Banner -->
      <div class="mt-2 p-3 rounded border border-[#1b3d1b] bg-[#0e1b0e] flex items-center gap-3">
        <Icon icon="ant-design:info-circle-outlined" :size="18" class="text-[#00ff41] flex-shrink-0" />
        <span class="text-xs text-[#8a9e8a]">
          <strong class="text-[#00ff41]">Tip:</strong> ScopeSentry works best with at least 1 active Node connected. Configure your external keys in Configuration &gt; API Key Management to gather assets with zero bandwidth cost.
        </span>
      </div>
    </section>
  </div>
</template>

<style scoped>
.help-page {
  font-family: 'JetBrains Mono', 'Inter', monospace !important;
}

.step-card:hover,
.menu-card:hover {
  box-shadow: 0 0 12px rgba(0, 255, 65, 0.2);
  transform: translateY(-2px);
}

:deep(.terminal-tabs .el-tabs__nav-wrap::after) {
  background-color: #1b3d1b !important;
}

:deep(.terminal-tabs .el-tabs__item) {
  color: #8a9e8a !important;
  font-family: 'JetBrains Mono', 'Inter', monospace !important;
  font-size: 13px !important;
}

:deep(.terminal-tabs .el-tabs__item.is-active) {
  color: #00ff41 !important;
  font-weight: 700 !important;
}

:deep(.terminal-tabs .el-tabs__active-bar) {
  background-color: #00ff41 !important;
  box-shadow: 0 0 8px #00ff41 !important;
}

:deep(.search-input .el-input__wrapper) {
  background-color: #0e1b0e !important;
  box-shadow: 0 0 0 1px #1b3d1b inset !important;
}

:deep(.search-input .el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px #00ff41 inset, 0 0 8px rgba(0, 255, 65, 0.3) !important;
}

:deep(.search-input .el-input__inner) {
  color: #d4d4d4 !important;
}

:deep(.search-input .el-input__inner::placeholder) {
  color: #556855 !important;
}
</style>
