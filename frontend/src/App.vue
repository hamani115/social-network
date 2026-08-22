<template>
  <div>
    <nav>
      <router-link v-if="auth.user" to="/">Feed</router-link>
      <span v-if="auth.user">| </span>
      <router-link v-if="auth.user" to="/users">Users</router-link>
      <span v-if="auth.user">| </span>
      <router-link v-if="auth.user" to="/groups">Groups
        <span v-if="websocket.groupUnreadTotal > 0">
          ({{ websocket.groupUnreadTotal }})
        </span>
      </router-link>
      <span v-if="auth.user">| </span>
      <router-link v-if="auth.user" to="/chat">Chat
        <span v-if="websocket.privateUnreadTotal > 0">
          ({{ websocket.privateUnreadTotal }})
        </span>
      </router-link>
      <span v-if="auth.user">| </span>
      <router-link v-if="auth.user" to="/profile/me">My Profile</router-link>
      <span v-if="auth.user">| </span>
      <router-link v-if="auth.user" to="/notifications">
        Notifications
        <span v-if="notifications.unreadCount > 0">
          ({{ notifications.unreadCount }})
        </span>
      </router-link>
      <span v-if="auth.user">| </span>
      <router-link v-if="!auth.user" to="/login">Login</router-link>
      <span v-if="!auth.user">| </span>
      <router-link v-if="!auth.user" to="/register">Register</router-link>

      <span v-if="auth.user">
        Logged in as {{ auth.user.first_name }} {{ auth.user.last_name }}
        <button @click="handleLogout">Logout</button>
      </span>
    </nav>

    <hr />
    <span v-if="auth.user">
      |
      WS:
      {{ websocket.connected ? "Connected" : "Disconnected" }}
    </span>

    <router-view />
  </div>
</template>

<script setup>
import { onMounted } from "vue";
import { watch } from "vue";
import { useAuthStore } from "./stores/auth";
import { useNotificationsStore } from "./stores/notifications";
import { useWebSocketStore } from "./stores/websocket";

const auth = useAuthStore();
const notifications = useNotificationsStore();
const websocket = useWebSocketStore();

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
      websocket.connect(user.id);
      await notifications.fetchNotifications();
    } else {
      notifications.clear();
      websocket.disconnect();
    }
  }
);

watch(
  () => websocket.eventVersion,
  () => {
    const event =
      websocket.lastEvent;

    if (!event) {
      return;
    }

    if (
      event.type === "notification" &&
      event.data
    ) {
      notifications.addNotification(
        event.data
      );
    }
  }
);

</script>