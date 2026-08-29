<template>
  <div class="app-shell">
    <header class="topbar">
      <div class="topbar-inner">
        <!-- Left Side -->
        <router-link to="/" class="brand">
          <img src="/social_network_logo.png" alt="" class="brand-logo" />

          <span class="brand-name">Social Network</span>
        </router-link>

        <nav v-if="auth.user" class="nav-links">
          <router-link class="nav-link" to="/">
            <i class="fa-solid fa-house" aria-hidden="true"></i>

            Feed
          </router-link>

          <router-link class="nav-link" to="/users">
            <i class="fa-solid fa-users" aria-hidden="true"></i>

            Users
          </router-link>

          <router-link class="nav-link" to="/groups">
            <i class="fa-solid fa-user-group" aria-hidden="true"></i>

            Groups

            <span v-if="websocket.groupUnreadTotal > 0" class="nav-badge">
              {{ websocket.groupUnreadTotal }}
            </span>
          </router-link>

          <router-link class="nav-link" to="/chat">
            <i class="fa-solid fa-comments" aria-hidden="true"></i>

            Chat

            <span v-if="websocket.privateUnreadTotal > 0" class="nav-badge">
              {{ websocket.privateUnreadTotal }}
            </span>
          </router-link>

          <router-link class="nav-link" to="/notifications">
            <i class="fa-solid fa-bell" aria-hidden="true"></i>

            Notifications

            <span v-if="notifications.unreadCount > 0" class="nav-badge">
              {{ notifications.unreadCount }}
            </span>
          </router-link>
        </nav>

        <!-- Right side -->
        <div class="topbar-actions">
          <template v-if="auth.user">
            <span
              class="connection-status"
              :class="{
                'is-connected': websocket.connected,
              }"
            >
              <span class="status-dot"></span>

              {{ websocket.connected ? "Connected" : "Disconnected" }}
            </span>

            <router-link class="user-chip" to="/profile/me">
              <UserAvatar
                :avatar-path="auth.user.avatar_path"
                :name="`${auth.user.first_name} ${auth.user.last_name}`"
                class="user-chip-avatar"
              />

              <span>
                {{ auth.user.first_name }}
                {{ auth.user.last_name }}
              </span>
            </router-link>

            <button class="button-ghost button-small" @click="handleLogout">
              <i class="fa-solid fa-right-from-bracket" aria-hidden="true"></i>

              Logout
            </button>
          </template>

          <template v-else>
            <router-link class="button button-ghost" to="/login">
              Log in
            </router-link>

            <router-link class="button button-primary" to="/register">
              Sign up
            </router-link>
          </template>
        </div>
      </div>
    </header>

    <div class="app-content">
      <router-view />
    </div>
  </div>
</template>

<script setup>
import { onMounted, onUnmounted, watch } from "vue";
import { useRouter } from "vue-router";
import { useAuthStore } from "./stores/auth";
import { useNotificationsStore } from "./stores/notifications";
import { useWebSocketStore } from "./stores/websocket";

import UserAvatar from "./components/UserAvatar.vue";

const auth = useAuthStore();
const notifications = useNotificationsStore();
const websocket = useWebSocketStore();
const router = useRouter();

async function handleLogout() {
  try {
    await auth.logout();

    await router.replace("/login");
  } catch (err) {
    console.error(err);
  }
}

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
  },
  {
    immediate: true,
  },
);

async function handleGlobalApiError(event) {
  const detail = event?.detail;
  if (!detail) {
    return;
  }

  const currentRoute = router.currentRoute.value;

  // SESSION EXPIRED
  if (detail.type === "unauthorized") {
    if (currentRoute.path === "/login" || currentRoute.path === "/register") {
      return;
    }

    const redirect = currentRoute.fullPath;
    auth.user = null;
    await router.replace({
      path: "/login",
      query: {
        redirect,
      },
    });
    return;
  }

  // BACKEND UNAVAILABLE
  if (detail.type === "server-unavailable") {
    if (currentRoute.path === "/server-unavailable") {
      return;
    }

    await router.replace({
      path: "/server-unavailable",
      query: {
        from: currentRoute.fullPath,
      },
    });
    return;
  }

  // GENERIC
  if (detail.type === "generic") {
    if (currentRoute.path === "/error") {
      return;
    }

    await router.replace({
      path: "/error",
      query: {
        from: currentRoute.fullPath,
      },
    });
  }
}

onMounted(() => {
  window.addEventListener("api-global-error", handleGlobalApiError);
});

onUnmounted(() => {
  window.removeEventListener("api-global-error", handleGlobalApiError);
});
</script>
