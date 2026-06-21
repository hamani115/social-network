<template>
  <main>
    <h1>Notifications</h1>

    <p v-if="notifications.loading">Loading notifications...</p>
    <p v-if="notifications.error">{{ notifications.error }}</p>

    <button v-if="notifications.notifications.length > 0" @click="notifications.markAllAsRead">
      Mark All as Read
    </button>

    <p v-if="notifications.notifications.length === 0">
      No notifications yet.
    </p>

    <article v-for="notification in notifications.notifications" :key="notification.id">
      <p>
        <strong v-if="!notification.is_read">Unread</strong>
        <span v-else>Read</span>
      </p>

      <p>{{ notification.message }}</p>

      <p>
        Type:
        <strong>{{ notification.type }}</strong>
      </p>

      <p>{{ notification.created_at }}</p>

      <button v-if="!notification.is_read" @click="notifications.markAsRead(notification.id)">
        Mark as Read
      </button>

      <button @click="openNotification(notification)">
        Open
      </button>

      <hr />
    </article>
  </main>
</template>

<script setup>
import { onMounted } from "vue";
import { useRouter } from "vue-router";
import { useNotificationsStore } from "../stores/notifications";

const router = useRouter();
const notifications = useNotificationsStore();

async function openNotification(notification) {
  if (!notification.is_read) {
    await notifications.markAsRead(notification.id);
  }

  if (notification.link_path) {
    router.push(notification.link_path);
  }
}

onMounted(() => {
  notifications.fetchNotifications();
});
</script>