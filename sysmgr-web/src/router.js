// sysmgr-web/src/router.js
import { createRouter, createWebHistory } from 'vue-router';
import Setup from './views/Setup.vue';
import Login from './views/Login.vue';
import Dashboard from './views/Dashboard.vue';

const routes = [
    { path: '/setup', component: Setup },
    { path: '/login', component: Login },
    { path: '/dashboard', component: Dashboard, meta: { requiresAuth: true } },
    { path: '/', redirect: '/dashboard' }
];

const router = createRouter({
    history: createWebHistory(),
    routes
});

// The Security Guard
router.beforeEach(async (to, from, next) => {
    try {
        // 1. Check if the master database needs initialization
        const statusRes = await fetch('http://localhost:8088/api/setup/status');
        const statusData = await statusRes.json();

        if (statusData.setup_required && to.path !== '/setup') {
            return next('/setup'); // Force user to the Setup screen
        }

        if (!statusData.setup_required && to.path === '/setup') {
            return next('/login'); // Prevent accessing setup if already initialized
        }

        // 2. Check RAM Session Authentication for the Dashboard
        if (to.meta.requiresAuth) {
            const token = localStorage.getItem('sysmgr_token');
            if (!token) {
                return next('/login');
            }
        }

        next();
    } catch (error) {
        console.error("Router guard error:", error);
        // Fail safe: if the API is totally offline, go to login
        if (to.path !== '/login') next('/login');
        else next();
    }
});

export default router;