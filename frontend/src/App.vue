<template>
  <div>
    <nav>
      <router-link to="/">Feed</router-link>
      |
      <router-link v-if="auth.user" to="/users">Users</router-link>
      |
      <router-link v-if="auth.user" to="/groups">Groups</router-link>
      |
      <router-link v-if="auth.user" to="/profile/me">My Profile</router-link>
      |
      <router-link v-if="auth.user" to="/notifications">
        Notifications
        <span v-if="notifications.unreadCount > 0">
          ({{ notifications.unreadCount }})
        </span>
      </router-link>
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
import { watch } from "vue";
import { useAuthStore } from "./stores/auth";
import { useNotificationsStore } from "./stores/notifications";

const auth = useAuthStore();
const notifications = useNotificationsStore();

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

watch(
  () => auth.user,
  async (user) => {
    if (user) {
      await notifications.fetchNotifications();
    } else {
      notifications.clear();
    }
  },
  { immediate: true }
);

</script>