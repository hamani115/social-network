<template>
  <div>
    <nav>
      <router-link to="/">Feed</router-link>
      |
      <router-link v-if="!auth.user" to="/login">Login</router-link>
      |
      <router-link v-if="!auth.user" to="/register">Register</router-link>

      <span v-if="auth.user">
        Logged in as {{ auth.user.first_name }} {{ auth.user.last_name }}
        <button @click="handleLogout">Logout</button>
      </span>
    </nav>

    <hr />

    <router-view />
  </div>
</template>

<script setup>
import { onMounted } from "vue";
import { useAuthStore } from "./stores/auth";

const auth = useAuthStore();

async function handleLogout() {
  try {
    await auth.logout();
  } catch (err) {
    console.error(err);
  }
}

onMounted(async () => {
  try {
    await auth.fetchMe();
  } catch {
    auth.user = null;
  }
});
</script>