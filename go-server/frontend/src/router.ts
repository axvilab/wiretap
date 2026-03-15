import { createRouter, createWebHistory } from 'vue-router'

import Dashboard from './pages/Dashboard.vue'
import Activity from './pages/Activity.vue'
import Performance from './pages/Performance.vue'
import Requests from './pages/Requests.vue'
import Trace from './pages/Trace.vue'
import Exceptions from './pages/Exceptions.vue'
import Jobs from './pages/Jobs.vue'
import Mail from './pages/Mail.vue'
import MailDetail from './pages/MailDetail.vue'
import Commands from './pages/Commands.vue'
import Schedule from './pages/Schedule.vue'
import Broadcasts from './pages/Broadcasts.vue'
import BroadcastDetail from './pages/BroadcastDetail.vue'
import Logs from './pages/Logs.vue'
import Notifications from './pages/Notifications.vue'
import NotificationDetail from './pages/NotificationDetail.vue'
import LogDetail from './pages/LogDetail.vue'
import Settings from './pages/Settings.vue'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: Dashboard },
    { path: '/activity', component: Activity },
    { path: '/performance', component: Performance },
    { path: '/requests', component: Requests },
    { path: '/trace/:rid', component: Trace, props: true },
    { path: '/exceptions', component: Exceptions },
    { path: '/jobs', component: Jobs },
    { path: '/mail', component: Mail },
    { path: '/mail/:id', component: MailDetail, props: true },
    { path: '/commands', component: Commands },
    { path: '/schedule', component: Schedule },
    { path: '/broadcasts', component: Broadcasts },
    { path: '/broadcasts/:rid', component: BroadcastDetail, props: true },
    { path: '/logs', component: Logs },
    { path: '/logs/:rid/:index', component: LogDetail, props: true },
    { path: '/notifications', component: Notifications },
    { path: '/notifications/:rid/:index', component: NotificationDetail, props: true },
    { path: '/settings', component: Settings },
  ],
})
