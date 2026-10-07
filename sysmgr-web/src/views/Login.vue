<!-- sysmgr-web/src/views/Login.vue -->
<template>
  <div class="min-h-screen bg-slate-100 flex items-center justify-center p-4">
    <div class="bg-white p-8 rounded-xl shadow-lg w-full max-w-md border border-slate-200">
      
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-indigo-100 text-indigo-600 mb-4">
          <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"></path></svg>
        </div>
        <h2 class="text-2xl font-black text-indigo-900 mb-2">SysMgr Login</h2>
        <p class="text-sm text-slate-500">Master administrative access only.</p>
      </div>
      
      <form @submit.prevent="requestLogin" class="space-y-5">
        <div>
          <label class="block text-slate-700 text-sm font-bold mb-2">Username</label>
          <input v-model="username" type="text" required class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500" />
        </div>
        <div>
          <label class="block text-slate-700 text-sm font-bold mb-2">Password</label>
          <input v-model="password" type="password" required class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500" />
        </div>
        <button type="submit" :disabled="isLoading" class="w-full bg-indigo-600 text-white font-bold py-3 px-4 rounded-lg hover:bg-indigo-700 transition-colors disabled:opacity-50">
          {{ isLoading ? 'Authenticating...' : 'Sign In' }}
        </button>
      </form>

      <!-- Global Error Banner -->
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

const isLoading = ref(false);

const username = ref('');
const password = ref('');
const error = ref('');

const requestLogin = async () => {
  error.value = '';
  isLoading.value = true;
  try {
    const res = await api.post('/auth/login', {
      username: username.value,
      password: password.value
    }, false); // false = bypass token verification

    if (res.status === 'success') {
      // Security: Save the token specific to the SysMgr, preventing crossover with the public portal
      localStorage.setItem('sysmgr_token', res.token);
      router.push('/dashboard');
    }
  } catch (err) {
    error.value = err.message;
  } finally {
    isLoading.value = false;
  }
};
</script>
