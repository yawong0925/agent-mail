<!-- sysmgr-web/src/views/Dashboard.vue -->
<template>
  <div class="min-h-screen bg-slate-50 flex flex-col">
    
    <!-- Top Navigation Bar -->
    <nav class="bg-indigo-900 text-white shadow-md z-10">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between h-16 items-center">
          <div class="flex items-center space-x-3">
            <!-- Icon -->
            <svg class="w-6 h-6 text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"></path></svg>
            <span class="font-bold text-xl tracking-wide">SysMgr Console</span>
          </div>
          <button @click="logout" class="bg-indigo-800 hover:bg-red-600 px-4 py-2 rounded text-sm font-semibold transition-colors duration-200">
            Secure Logout
          </button>
        </div>
      </div>
    </nav>

    <div class="flex-grow flex max-w-7xl w-full mx-auto">
      
      <!-- Left Sidebar Menu -->
      <aside class="w-64 bg-white border-r border-slate-200 py-6 pr-6 hidden md:block">
        <nav class="space-y-2">
          <!-- Switch tabs by changing the 'activeTab' ref -->
          <button @click="activeTab = 'telemetry'" :class="['w-full text-left px-4 py-3 rounded-lg font-medium transition-colors', activeTab === 'telemetry' ? 'bg-indigo-50 text-indigo-700' : 'text-slate-600 hover:bg-slate-50']">
            System Telemetry
          </button>
          <button @click="activeTab = 'users'" :class="['w-full text-left px-4 py-3 rounded-lg font-medium transition-colors', activeTab === 'users' ? 'bg-indigo-50 text-indigo-700' : 'text-slate-600 hover:bg-slate-50']">
            User Management
          </button>
          <button @click="activeTab = 'settings'" :class="['w-full text-left px-4 py-3 rounded-lg font-medium transition-colors', activeTab === 'settings' ? 'bg-indigo-50 text-indigo-700' : 'text-slate-600 hover:bg-slate-50']">
            Global Settings
          </button>
        </nav>
      </aside>

      <!-- Main Content Area -->
      <main class="flex-grow p-6 md:p-8">
        
        <!-- Error Banner -->
        <div v-if="error" class="bg-red-50 border-l-4 border-red-500 text-red-700 p-4 mb-6 rounded shadow-sm">
          <p class="font-bold">System Warning</p>
          <p>{{ error }}</p>
        </div>

        <!-- ============================== -->
        <!-- TAB 1: SYSTEM TELEMETRY        -->
        <!-- ============================== -->
        <div v-if="activeTab === 'telemetry'">
          <h2 class="text-2xl font-bold text-slate-800 mb-6">System Overview</h2>
          
          <!-- Key Metrics Row -->
          <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
            <div class="bg-white p-6 rounded-xl shadow-sm border border-slate-200">
              <h3 class="text-slate-500 text-xs font-bold uppercase tracking-wider mb-2">Total Ingested Emails</h3>
              <p class="text-4xl font-black text-slate-800">{{ stats.total_emails?.toLocaleString() || 0 }}</p>
            </div>
            
            <div class="bg-white p-6 rounded-xl shadow-sm border border-slate-200">
              <h3 class="text-slate-500 text-xs font-bold uppercase tracking-wider mb-2">Attachment Storage</h3>
              <p class="text-4xl font-black text-slate-800">{{ stats.disk_usage_mb?.toFixed(2) || '0.00' }} <span class="text-lg text-slate-400 font-medium">MB</span></p>
            </div>
            
            <div class="bg-white p-6 rounded-xl shadow-sm border border-slate-200">
              <h3 class="text-slate-500 text-xs font-bold uppercase tracking-wider mb-2">Active Email Nodes</h3>
              <p class="text-4xl font-black text-emerald-600">{{ stats.active_accounts || 0 }} <span class="text-lg text-emerald-400 font-medium">Syncing</span></p>
            </div>
          </div>

          <!-- Chart Placeholders (To be implemented with vue-chartjs later) -->
          <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div class="bg-white p-6 rounded-xl shadow-sm border border-slate-200 min-h-[300px] flex flex-col">
              <h3 class="font-bold text-slate-700 mb-4">Bandwidth Usage (24H)</h3>
              <div class="flex-grow flex items-center justify-center border-2 border-dashed border-slate-100 rounded-lg bg-slate-50">
                <span class="text-slate-400 text-sm">[ Line Chart Visualization Area ]</span>
              </div>
            </div>

            <div class="bg-white p-6 rounded-xl shadow-sm border border-slate-200 min-h-[300px] flex flex-col">
              <h3 class="font-bold text-slate-700 mb-4">Disk Usage by Provider</h3>
              <div class="flex-grow flex items-center justify-center border-2 border-dashed border-slate-100 rounded-lg bg-slate-50">
                <span class="text-slate-400 text-sm">[ Bar Chart Visualization Area ]</span>
              </div>
            </div>
          </div>
        </div>

        <!-- ============================== -->
        <!-- TAB 2: USER MANAGEMENT         -->
        <!-- ============================== -->
        <div v-if="activeTab === 'users'">
          <div class="flex justify-between items-center mb-6">
            <h2 class="text-2xl font-bold text-slate-800">Local Users</h2>
            <button class="bg-indigo-600 hover:bg-indigo-700 text-white px-4 py-2 rounded-lg text-sm font-semibold transition">
              + Create User
            </button>
          </div>

          <div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
            <table class="min-w-full divide-y divide-slate-200">
              <thead class="bg-slate-50">
                <tr>
                  <th class="px-6 py-3 text-left text-xs font-bold text-slate-500 uppercase tracking-wider">Username</th>
                  <th class="px-6 py-3 text-left text-xs font-bold text-slate-500 uppercase tracking-wider">Status</th>
                  <th class="px-6 py-3 text-right text-xs font-bold text-slate-500 uppercase tracking-wider">Actions</th>
                </tr>
              </thead>
              <tbody class="bg-white divide-y divide-slate-200">
                <!-- Placeholder row until we build the User List API endpoint -->
                <tr>
                  <td class="px-6 py-4 whitespace-nowrap">
                    <div class="text-sm font-medium text-slate-900">admin</div>
                    <div class="text-sm text-slate-500">Master Account</div>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap">
                    <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-emerald-100 text-emerald-800">Active</span>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    <button class="text-indigo-600 hover:text-indigo-900 mr-3">Edit</button>
                    <button class="text-slate-400 cursor-not-allowed" disabled>Lock</button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- ============================== -->
        <!-- TAB 3: GLOBAL SETTINGS         -->
        <!-- ============================== -->
        <div v-if="activeTab === 'settings'">
          <h2 class="text-2xl font-bold text-slate-800 mb-6">System Configuration</h2>
          <div class="bg-white p-6 rounded-xl shadow-sm border border-slate-200 max-w-2xl">
            
            <div class="flex items-center justify-between py-4 border-b border-slate-100">
              <div>
                <h4 class="font-bold text-slate-800">Public Registration</h4>
                <p class="text-sm text-slate-500">Allow users to sign up via the public portal (:8080).</p>
              </div>
              <button class="bg-slate-200 px-4 py-2 rounded-full text-sm font-bold text-slate-500">Disabled</button>
            </div>

            <div class="flex items-center justify-between py-4">
              <div>
                <h4 class="font-bold text-slate-800">Global API Rate Limit</h4>
                <p class="text-sm text-slate-500">Maximum Agent requests per minute per token.</p>
              </div>
              <input type="number" value="10" class="w-20 px-3 py-2 border rounded-lg text-center font-bold text-slate-700" />
            </div>

          </div>
        </div>

      </main>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import { api } from '../api.js';

const router = useRouter();

// Tab state management
const activeTab = ref('telemetry'); 

// Data state
const stats = ref({});
const error = ref('');
let pollingInterval = null;

// Fetches the live stats from the Go backend
const fetchTelemetry = async () => {
  try {
    const data = await api.get('/system/telemetry');
    stats.value = data;
    error.value = ''; // Clear any previous errors
  } catch (err) {
    error.value = err.message;
    // If the RAM token expired, the backend returns Unauthorized (401).
    if (err.message.includes('Unauthorized')) {
      logout();
    }
  }
};

// Completely wipes the session token from the browser and redirects
const logout = () => {
  localStorage.removeItem('sysmgr_token');
  router.push('/login');
};

// Start fetching data immediately when the component renders
onMounted(() => {
  fetchTelemetry();
  // Set up a background timer to silently refresh the stats every 15 seconds
  pollingInterval = setInterval(fetchTelemetry, 15000);
});

// Clean up the timer if the admin navigates away from this component
onUnmounted(() => {
  if (pollingInterval) clearInterval(pollingInterval);
});
</script>