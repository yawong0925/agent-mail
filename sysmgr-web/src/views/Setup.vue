<!-- sysmgr-web/src/views/Setup.vue -->
<template>
  <div class="min-h-screen bg-slate-100 flex items-center justify-center p-4">
    <div class="bg-white p-8 rounded-xl shadow-lg w-full max-w-md border border-slate-200">
      
      <!-- Header -->
      <div class="text-center mb-8">
        <h2 class="text-2xl font-black text-indigo-900 mb-2">System Initialization</h2>
        <p class="text-sm text-slate-500">Create the master administrator account to secure the Agent Mail system.</p>
      </div>
      
      <!-- Form bound to the 'createAdmin' function -->
      <form @submit.prevent="createAdmin" class="space-y-5">
        <div>
          <label class="block text-slate-700 text-sm font-bold mb-2">Admin Username</label>
          <input 
            v-model="form.username" 
            type="text" 
            required 
            placeholder="e.g. admin"
            class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:bg-white transition-all" 
          />
        </div>

        <div>
          <label class="block text-slate-700 text-sm font-bold mb-2">Admin Email</label>
          <input 
            v-model="form.email" 
            type="email" 
            required 
            placeholder="admin@yourdomain.com"
            class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:bg-white transition-all" 
          />
        </div>

        <div>
          <label class="block text-slate-700 text-sm font-bold mb-2">Master Password</label>
          <input 
            v-model="form.password" 
            type="password" 
            required 
            placeholder="••••••••"
            class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:bg-white transition-all" 
          />
        </div>

        <button 
          type="submit" 
          :disabled="isLoading"
          class="w-full bg-indigo-600 text-white font-bold py-3 px-4 rounded-lg hover:bg-indigo-700 transition-colors disabled:opacity-50"
        >
          {{ isLoading ? 'Initializing...' : 'Initialize System' }}
        </button>
      </form>

      <!-- Error Display -->
      <div v-if="error" class="mt-6 p-4 bg-red-50 text-red-600 rounded-lg text-sm border border-red-100 text-center">
        {{ error }}
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { api } from '../api.js';

const router = useRouter();

// Reactive state variables
const form = ref({ username: '', email: '', password: '' });
const error = ref('');
const isLoading = ref(false);

const createAdmin = async () => {
  error.value = '';
  isLoading.value = true;

  try {
    // The 'false' parameter tells our API wrapper NOT to attach a Bearer token, 
    // since we don't have one yet!
    const res = await api.post('/setup/admin', form.value, false);
    
    if (res.status === 'success') {
      alert('System initialized successfully! Please log in.');
      // Instantly route them to the login screen
      router.push('/login');
    }
  } catch (err) {
    error.value = err.message;
  } finally {
    isLoading.value = false;
  }
};
</script>