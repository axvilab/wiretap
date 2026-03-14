import { createRouter, createWebHistory } from 'vue-router'

import Dashboard from './pages/Dashboard.vue'
import Requests from './pages/Requests.vue'
import Trace from './pages/Trace.vue'
import Exceptions from './pages/Exceptions.vue'
import Jobs from './pages/Jobs.vue'
import Mail from './pages/Mail.vue'
import MailDetail from './pages/MailDetail.vue'
import Broadcasts from './pages/Broadcasts.vue'
import Settings from './pages/Settings.vue'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: Dashboard },
    { path: '/requests', component: Requests },
    { path: '/trace/:rid', component: Trace, props: true },
    { path: '/exceptions', component: Exceptions },
    { path: '/jobs', component: Jobs },
    { path: '/mail', component: Mail },
    { path: '/mail/:id', component: MailDetail, props: true },
    { path: '/broadcasts', component: Broadcasts },
    { path: '/settings', component: Settings },
  ],
})
