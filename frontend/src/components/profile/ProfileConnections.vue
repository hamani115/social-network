<template>
  <section class="info-card connections-card">
    <div class="card-heading">
      <h2>Connections</h2>
    </div>

    <!-- Tabs -->
    <div class="connection-tabs">
      <button
        type="button"
        class="connection-tab"
        :class="{ active: activeConnectionsTab === 'followers' }"
        @click="selectConnectionsTab('followers')"
      >
        Followers
        <span>{{ profile.followers_count }}</span>
      </button>

      <button
        type="button"
        class="connection-tab"
        :class="{ active: activeConnectionsTab === 'following' }"
        @click="selectConnectionsTab('following')"
      >
        Following
        <span>{{ profile.following_count }}</span>
      </button>
    </div>

    <p v-if="loadingFollowLists" class="connections-status">
      Loading connections...
    </p>

    <p v-else-if="followListsError" class="connections-error">
      {{ followListsError }}
    </p>

    <div v-else-if="activeConnections.length > 0" class="connections-content">
      <div
        class="connections-list"
        :class="{
          'connections-list-expanded': showAllConnections,
        }"
      >
        <article
          v-for="user in visibleConnections"
          :key="user.id"
          class="connection-person"
        >
          <router-link
            :to="`/profiles/${user.id}`"
            class="connection-user-link"
          >
            <div class="connection-avatar">
              {{ connectionInitials(user) }}
            </div>

            <div class="connection-user-info">
              <strong>
                {{ user.first_name }}
                {{ user.last_name }}
              </strong>

              <span v-if="user.nickname">
                {{ user.nickname }}
              </span>
            </div>
          </router-link>
        </article>
      </div>

      <button
        v-if="activeConnections.length > connectionPreviewLimit"
        type="button"
        class="connections-more-button"
        @click="showAllConnections = !showAllConnections"
      >
        {{
          showAllConnections
            ? "Show less"
            : `Show all ${activeConnections.length}`
        }}
      </button>
    </div>

    <p v-else class="empty-text connections-empty">
      {{
        activeConnectionsTab === "followers"
          ? "No followers yet"
          : "Not following anyone yet"
      }}
    </p>
  </section>
</template>

<script setup>
import { computed, ref, watch } from "vue";

import { apiRequest } from "../../services/api";

const props = defineProps({
  profile: {
    type: Object,
    required: true,
  },
});

const followers = ref([]);
const following = ref([]);

const loadingFollowLists = ref(false);
const followListsError = ref("");

const showAllConnections = ref(false);
const connectionPreviewLimit = 4;

const activeConnectionsTab = ref("followers");

const activeConnections = computed(() => {
  if (activeConnectionsTab.value === "followers") {
    return followers.value;
  }

  return following.value;
});

const visibleConnections = computed(() => {
  if (showAllConnections.value) {
    return activeConnections.value;
  }

  return activeConnections.value.slice(0, connectionPreviewLimit);
});

// Connections

async function loadFollowLists() {
  if (!props.profile?.id) {
    return;
  }

  try {
    loadingFollowLists.value = true;
    followListsError.value = "";

    followers.value = await apiRequest(`/users/${props.profile.id}/followers`);

    following.value = await apiRequest(`/users/${props.profile.id}/following`);
  } catch (err) {
    followers.value = [];
    following.value = [];

    followListsError.value = err.message;
  } finally {
    loadingFollowLists.value = false;
  }
}

function selectConnectionsTab(tab) {
  activeConnectionsTab.value = tab;
  showAllConnections.value = false;
}

// Helpers

function connectionInitials(user) {
  const firstInitial = user.first_name?.charAt(0) || "";

  const lastInitial = user.last_name?.charAt(0) || "";

  return (firstInitial + lastInitial).toUpperCase();
}

// Reload when a different/new profile is received

watch(
  () => props.profile,
  () => {
    loadFollowLists();
  },
  {
    immediate: true,
  },
);
</script>

<style scoped>
.info-card {
  margin-top: 20px;
  padding: 26px 30px;

  border: 1px solid rgba(255, 255, 255, 0.08);

  border-radius: 16px;

  background: rgba(20, 25, 35, 0.95);

  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.18);
}

.card-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;

  gap: 16px;

  margin-bottom: 20px;
}

.card-heading h2 {
  margin: 0;

  color: #f8fafc;

  font-size: 22px;
}

.empty-text {
  margin: 0 0 24px;

  color: #64748b;

  font-style: italic;
}

/* Connection tabs */

.connection-tabs {
  display: flex;

  gap: 10px;

  margin-bottom: 22px;
}

.connection-tab {
  display: flex;
  align-items: center;

  gap: 8px;

  padding: 9px 16px;

  border: 1px solid rgba(255, 255, 255, 0.08);

  border-radius: 999px;

  color: #94a3b8;

  background: rgba(255, 255, 255, 0.025);

  font-size: 14px;
  font-weight: 600;

  cursor: pointer;

  transition:
    color 0.2s,
    border-color 0.2s,
    background 0.2s;
}

.connection-tab span {
  color: #64748b;
}

.connection-tab:hover {
  color: #cbd5e1;

  border-color: rgba(79, 156, 255, 0.35);
}

.connection-tab.active {
  color: #80b7ff;

  border-color: rgba(79, 156, 255, 0.55);

  background: rgba(79, 156, 255, 0.1);
}

.connection-tab.active span {
  color: #80b7ff;
}

/* Connection list */

.connections-list {
  display: flex;
  flex-direction: column;
}

.connection-person {
  display: flex;
  align-items: center;

  margin: 0;
  padding: 0;

  border: 0;

  border-bottom: 1px solid rgba(255, 255, 255, 0.06);

  border-radius: 0;

  background: transparent;
}

.connection-person:last-child {
  border-bottom: none;
}

.connection-user-link {
  width: 100%;

  display: flex;
  align-items: center;

  gap: 14px;

  padding: 14px 4px;

  border-radius: 8px;

  color: inherit;

  text-decoration: none;

  transition: background 0.2s;
}

.connection-user-link:hover {
  background: rgba(255, 255, 255, 0.035);
}

.connection-avatar {
  width: 48px;
  height: 48px;

  flex-shrink: 0;

  display: flex;
  align-items: center;
  justify-content: center;

  border-radius: 50%;

  color: #eaf3ff;

  background: linear-gradient(135deg, #4f9cff, #2867d6);

  font-size: 15px;
  font-weight: 700;
}

.connection-user-info {
  min-width: 0;

  display: flex;
  flex-direction: column;

  gap: 4px;
}

.connection-user-info strong {
  color: #e2e8f0;

  font-size: 15px;
}

.connection-user-link:hover strong {
  color: #80b7ff;

  text-decoration: underline;
}

.connection-user-info span {
  overflow: hidden;

  color: #94a3b8;

  font-size: 13px;

  text-overflow: ellipsis;
  white-space: nowrap;
}

/* States */

.connections-status {
  color: #94a3b8;
}

.connections-error {
  color: #f87171;
}

.connections-empty {
  margin-bottom: 0;
}

/* Show all */

.connections-more-button {
  width: 100%;

  margin-top: 12px;
  padding: 10px;

  border: none;
  border-radius: 8px;

  color: #80b7ff;

  background: transparent;

  font-size: 14px;
  font-weight: 600;

  cursor: pointer;
}

.connections-more-button:hover {
  background: rgba(79, 156, 255, 0.08);
}

.connections-list-expanded {
  max-height: 380px;

  overflow-y: auto;

  padding-right: 6px;

  scrollbar-width: thin;

  scrollbar-color: rgba(148, 163, 184, 0.3) transparent;
}

.connections-list-expanded::-webkit-scrollbar {
  width: 8px;
}

.connections-list-expanded::-webkit-scrollbar-track {
  background: transparent;
}

.connections-list-expanded::-webkit-scrollbar-thumb {
  border-radius: 999px;

  background: rgba(148, 163, 184, 0.25);
}

.connections-list-expanded::-webkit-scrollbar-thumb:hover {
  background: rgba(148, 163, 184, 0.4);
}

@media (max-width: 700px) {
  .connections-list-expanded {
    max-height: 50vh;

    overflow-y: auto;

    overscroll-behavior: contain;
  }
}
</style>
