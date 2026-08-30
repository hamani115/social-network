<template>
  <main class="notifications-page">
    

    <header class="notifications-header">
      <h1>Notifications</h1>

      <button
        v-if="notifications.unreadCount > 0"
        type="button"
        class="button button-ghost button-small"
        @click="notifications.markAllAsRead()"
      >
        Mark all as read
      </button>
    </header>

    <!-- FILTERS -->

    <div class="notification-filters">
      <button
        type="button"
        class="notification-filter"
        :class="{
          active: notifications.filter === 'all',
        }"
        @click="setNotificationFilter('all')"
      >
        All
      </button>

      <button
        type="button"
        class="notification-filter"
        :class="{
          active: notifications.filter === 'unread',
        }"
        @click="setNotificationFilter('unread')"
      >
        Unread

        <span
          v-if="notifications.unreadCount > 0"
          class="notification-filter-count"
        >
          {{ notifications.unreadCount }}
        </span>
      </button>
    </div>

    

    <p v-if="notifications.error" class="notifications-error">
      {{ notifications.error }}
    </p>

    <!-- INITIAL LOADING -->

    <div v-if="notifications.loading" class="notifications-state">
      <span class="loading-spinner"></span>

      Loading notifications...
    </div>

    

    <section
      v-else-if="notifications.notifications.length === 0"
      class="notifications-empty"
    >
      <div class="notifications-empty-icon">
        <i class="fa-solid fa-bell" aria-hidden="true"></i>
      </div>

      <h2>
        {{
          notifications.filter === "unread"
            ? "You're all caught up"
            : "No notifications yet"
        }}
      </h2>
    </section>

    <!-- LIST -->

    <section v-else class="notifications-list">
      <button
        v-for="notification in notifications.notifications"
        :key="notification.id"
        type="button"
        class="notification-card"
        :class="{
          unread: !notification.is_read,
        }"
        @click="openNotification(notification)"
      >
        <!-- UNREAD DOT -->

        <span class="notification-unread-column">
          <span
            v-if="!notification.is_read"
            class="notification-unread-dot"
          ></span>
        </span>

        <!-- ICON -->

        <span
          class="notification-icon"
          :class="`type-${notification.type}`"
          aria-hidden="true"
        >
          <i :class="notificationMeta(notification.type).icon"></i>
        </span>

        <!-- CONTENT -->

        <span class="notification-content">
          <span class="notification-meta-row">
            <strong>
              {{ notificationMeta(notification.type).label }}
            </strong>

            <span>
              {{ formatNotificationTime(notification.created_at) }}
            </span>
          </span>

          <span class="notification-message">
            {{ notification.message }}
          </span>
        </span>

        <!-- OPEN -->

        <span
          v-if="notification.link_path"
          class="notification-arrow"
          aria-hidden="true"
        >
          <i class="fa-solid fa-chevron-right"></i>
        </span>
      </button>

      <!-- PAGINATION -->

      <div class="notifications-pagination">
        <div
          v-if="notifications.loadingMore"
          class="notifications-loading-more"
        >
          <span class="loading-spinner"></span>

          Loading more...
        </div>

        <div
          v-else-if="notifications.loadMoreError"
          class="notifications-load-error"
        >
          <span>
            {{ notifications.loadMoreError }}
          </span>

          <button
            type="button"
            class="button button-ghost button-small"
            @click="notifications.fetchNotifications(false)"
          >
            Try again
          </button>
        </div>

        <div
          v-else-if="notifications.hasMore"
          ref="notificationsLoadTrigger"
          class="notifications-load-trigger"
          aria-hidden="true"
        ></div>

        <p v-else class="notifications-end">You're all caught up</p>
      </div>
    </section>
  </main>
</template>

<script setup>
import { onMounted, onUnmounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import { useNotificationsStore } from "../stores/notifications";
import { formatNotificationTime } from "../utils/date";

const router = useRouter();
const notifications = useNotificationsStore();
const notificationsLoadTrigger = ref(null);
let notificationsObserver = null;

async function setNotificationFilter(filter) {
  await notifications.setFilter(filter);
}

async function openNotification(notification) {
  if (!notification) {
    return;
  }

  if (!notification.is_read) {
    try {
      await notifications.markAsRead(notification.id);
    } catch {}
  }

  if (notification.link_path) {
    router.push(notification.link_path);
  }
}

function notificationMeta(type) {
  switch (type) {
    case "follow_request":
      return {
        icon: "fa-solid fa-user-plus",
        label: "Follow request",
      };
    case "follow_accepted":
      return {
        icon: "fa-solid fa-user-check",
        label: "Follow accepted",
      };
    case "group_invitation":
      return {
        icon: "fa-solid fa-envelope",
        label: "Group invitation",
      };
    case "group_invitation_accepted":
      return {
        icon: "fa-solid fa-check",
        label: "Invitation accepted",
      };
    case "group_join_request":
      return {
        icon: "fa-solid fa-user-group",
        label: "Join request",
      };
    case "group_join_accepted":
      return {
        icon: "fa-solid fa-circle-check",
        label: "Join accepted",
      };
    case "group_join_declined":
      return {
        icon: "fa-solid fa-circle-xmark",
        label: "Join declined",
      };
    case "group_event_created":
      return {
        icon: "fa-solid fa-calendar-days",
        label: "Group event",
      };
    default:
      return {
        icon: "fa-solid fa-bell",
        label: "Notification",
      };
  }
}

function observeNotificationsTrigger(element) {
  if (notificationsObserver) {
    notificationsObserver.disconnect();

    notificationsObserver = null;
  }

  if (!element) {
    return;
  }

  notificationsObserver = new IntersectionObserver(
    (entries) => {
      const entry = entries[0];

      if (
        entry.isIntersecting &&
        notifications.hasMore &&
        !notifications.loading &&
        !notifications.loadingMore
      ) {
        notifications.fetchNotifications(false);
      }
    },
    {
      root: null,
      rootMargin: "300px 0px",
      threshold: 0,
    },
  );

  notificationsObserver.observe(element);
}

watch(notificationsLoadTrigger, (element) => {
  observeNotificationsTrigger(element);
});


onMounted(async () => {
  if (!notifications.initialized) {
    await notifications.fetchNotifications(true);
  }
});

onUnmounted(() => {
  if (notificationsObserver) {
    notificationsObserver.disconnect();
  }
});
</script>

<style scoped>
.notifications-page {
  width: min(820px, 100%);
}

.notifications-header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;

  gap: 20px;

  margin-bottom: 20px;
}

.notifications-header h1 {
  margin-bottom: 6px;
}

.notification-filters {
  width: fit-content;

  display: flex;

  gap: 4px;

  margin-bottom: 18px;
  padding: 4px;

  border: 1px solid var(--border-soft);
  border-radius: var(--radius-round);
  background: var(--surface);
}

.notification-filter {
  min-height: 34px;

  display: inline-flex;
  align-items: center;

  gap: 6px;

  padding: 5px 13px;
  border: 0;
  border-radius: var(--radius-round);

  background: transparent;

  color: var(--text-muted);

  font-size: 12px;
  font-weight: 650;
}

.notification-filter:hover {
  background: var(--surface-2);

  color: var(--text);
}

.notification-filter.active {
  background: var(--primary-soft);

  color: var(--primary);
}

.notification-filter-count {
  min-width: 18px;
  height: 18px;

  display: inline-flex;
  align-items: center;
  justify-content: center;

  padding: 0 5px;

  border-radius: var(--radius-round);

  background: var(--primary);

  color: white;

  font-size: 9px;
}

.notifications-list {
  overflow: hidden;

  border: 1px solid var(--border-soft);

  border-radius: var(--radius-xl);

  background: var(--surface);

  box-shadow: var(--shadow-sm);
}

.notification-card {
  width: 100%;
  min-height: 84px;

  display: flex;
  align-items: center;

  gap: 12px;

  padding: 14px 16px;

  border: 0;
  border-bottom: 1px solid var(--border-soft);

  border-radius: 0;

  background: transparent;

  color: inherit;

  text-align: left;
}

.notification-card:hover {
  background: var(--surface-2);
}

.notification-card.unread {
  background: rgba(79, 156, 255, 0.055);
}

.notification-card.unread:hover {
  background: rgba(79, 156, 255, 0.09);
}

.notification-unread-column {
  width: 8px;

  flex: 0 0 8px;

  display: flex;
  justify-content: center;
}

.notification-unread-dot {
  width: 7px;
  height: 7px;

  border-radius: 50%;

  background: var(--primary);
}

.notification-icon {
  width: 42px;
  height: 42px;

  flex: 0 0 42px;

  display: grid;
  place-items: center;

  border: 1px solid var(--border-soft);

  border-radius: 50%;

  background: var(--bg-secondary);

  color: var(--text-secondary);

  font-size: 16px;
}

.notification-card.unread .notification-icon {
  border-color: var(--primary-border);
  background: var(--primary-soft);
  color: var(--primary);
}

.notification-content {
  min-width: 0;
  flex: 1;

  display: flex;
  flex-direction: column;

  gap: 4px;
}

.notification-meta-row {
  display: flex;
  align-items: center;
  justify-content: space-between;

  gap: 12px;
}

.notification-meta-row strong {
  font-size: 11px;
}

.notification-meta-row > span {
  flex-shrink: 0;

  color: var(--text-muted);

  font-size: 10px;
}

.notification-message {
  color: var(--text-secondary);

  font-size: 13px;

  line-height: 1.45;
}

.notification-card.unread .notification-message {
  color: var(--text);
}

.notification-arrow {
  flex-shrink: 0;

  color: var(--text-muted);

  font-size: 12px;
}

.notifications-state {
  min-height: 260px;

  display: flex;
  align-items: center;
  justify-content: center;

  gap: 8px;

  color: var(--text-muted);
}

.notifications-empty {
  min-height: 300px;

  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;

  padding: 30px;

  border: 1px solid var(--border-soft);

  border-radius: var(--radius-xl);

  background: var(--surface);

  color: var(--text-muted);

  text-align: center;
}

.notifications-empty-icon {
  width: 54px;
  height: 54px;

  display: grid;
  place-items: center;

  margin-bottom: 12px;

  border-radius: 50%;

  background: var(--primary-soft);

  font-size: 21px;
}

.notifications-empty h2 {
  margin-bottom: 5px;
}


.notifications-error {
  margin-bottom: 15px;

  padding: 10px 12px;

  border: 1px solid rgba(255, 95, 109, 0.22);

  border-radius: var(--radius-md);

  background: var(--danger-soft);

  color: var(--danger);

  font-size: 12px;
}

.notifications-pagination {
  padding: 15px;

  text-align: center;
}

.notifications-loading-more,
.notifications-load-error {
  min-height: 38px;

  display: flex;
  align-items: center;
  justify-content: center;

  gap: 8px;

  color: var(--text-muted);

  font-size: 11px;
}

.notifications-load-error {
  color: var(--danger);
}

.notifications-load-trigger {
  height: 2px;
}

.notifications-end {
  margin: 0;

  color: var(--text-muted);

  font-size: 11px;
}

@media (max-width: 700px) {
  .notifications-header {
    align-items: stretch;

    flex-direction: column;
  }

  .notifications-header button {
    width: fit-content;
  }

  .notification-card {
    align-items: flex-start;

    padding: 13px 11px;
  }

  .notification-icon {
    width: 38px;
    height: 38px;

    flex-basis: 38px;
  }

  .notification-meta-row {
    align-items: flex-start;

    flex-direction: column;

    gap: 2px;
  }

  .notification-arrow {
    display: none;
  }
}
</style>
