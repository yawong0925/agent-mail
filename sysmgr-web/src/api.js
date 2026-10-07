const API_BASE = 'http://localhost:8088/api'; // Pointing to SysMgr Port

export const api = {
    getToken() {
        return localStorage.getItem('sysmgr_token'); // Different token name to avoid conflicts
    },

    async get(endpoint, requiresAuth = true) {
        return this.request('GET', endpoint, null, requiresAuth);
    },

    async post(endpoint, data, requiresAuth = true) {
        return this.request('POST', endpoint, data, requiresAuth);
    },

    async request(method, endpoint, data = null, requiresAuth = true) {
        const headers = { 'Content-Type': 'application/json' };

        if (requiresAuth) {
            const token = this.getToken();
            if (token) headers['Authorization'] = `Bearer ${token}`;
        }

        const config = { method, headers };
        if (data) config.body = JSON.stringify(data);

        const response = await fetch(`${API_BASE}${endpoint}`, config);
        const json = await response.json();

        if (!response.ok) throw new Error(json.error || 'API Error');
        return json;
    }
};