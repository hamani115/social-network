<template>
  <main class="users-page">
    <!-- PAGE HEADER -->

    <header class="users-header">
      <h1>Users</h1>
    </header>

    <!-- FOLLOW REQUESTS -->

    <section
      v-if="loadingRequests || requestsError || followRequests.length > 0"
      class="follow-requests-card"
    >
      <div class="users-section-heading">
        <div>
          <h2>Follow requests</h2>
        </div>

        <span v-if="followRequests.length > 0" class="request-count">
          {{ followRequests.length }}
        </span>
      </div>

      <div v-if="loadingRequests" class="users-state">
        <span class="loading-spinner"></span>

        Loading follow requests...
      </div>

      <p v-else-if="requestsError" class="users-error">
        {{ requestsError }}
      </p>

      <div v-else class="follow-request-list">
        <article
          v-for="request in followRequests"
          :key="request.id"
          class="follow-request-item"
        >
          <router-link
            :to="`/profiles/${request.requester_id}`"
            class="request-user"
          >
            <UserAvatar
              :avatar-path="request.requester_avatar_path"
              :name="request.requester_name"
              class="users-avatar"
            />

            <div class="request-user-info">
              <strong>
                {{ request.requester_name }}
              </strong>

              <span v-if="request.requester_nickname">
                {{ request.requester_nickname }}
              </span>

              <span v-else> Wants to follow you </span>
            </div>
          </router-link>

          <div class="request-actions">
            <button
              type="button"
              class="button button-ghost button-small"
              :disabled="changingRequestId === request.id"
              @click="declineRequest(request.id)"
            >
              Decline
            </button>

            <button
              type="button"
              class="button button-primary button-small"
              :disabled="changingRequestId === request.id"
              @click="acceptRequest(request.id)"
            >
              {{ changingRequestId === request.id ? "Working..." : "Accept" }}
            </button>
          </div>
        </article>
      </div>
    </section>

    <!-- PEOPLE DIRECTORY -->

    <section class="users-section">
      <div class="users-section-heading">
        <div>
          <h2>Find users</h2>
        </div>
      </div>

      <!-- SEARCH -->

      <div class="users-search">
        <span class="users-search-icon" aria-hidden="true">
          <i class="fa-solid fa-magnifying-glass"></i>
        </span>

        <label for="users-search" class="visually-hidden"> Search users </label>

        <input
          id="users-search"
          v-model="searchQuery"
          type="search"
          placeholder="Search users..."
          autocomplete="off"
        />

        <button
          v-if="searchQuery"
          type="button"
          class="users-search-clear"
          aria-label="Clear users search"
          @click="clearSearch"
        >
          <i class="fa-solid fa-xmark" aria-hidden="true"></i>
        </button>
      </div>

      <!-- LOADING -->

      <div v-if="loadingUsers" class="users-state">
        <span class="loading-spinner"></span>

        Loading users...
      </div>

      <!-- ERROR -->

      <p v-else-if="usersError" class="users-error">
        {{ usersError }}
      </p>

      <!-- SEARCH HAS NO RESULTS -->

      <div
        v-else-if="users.length === 0 && debouncedSearchQuery"
        class="users-empty"
      >
        <div class="users-empty-icon">
          <i class="fa-solid fa-magnifying-glass" aria-hidden="true"></i>
        </div>

        <h3>No users found</h3>

        <p>No users match "{{ searchQuery }}"</p>
      </div>

      <!-- NO USERS AT ALL -->

      <div v-else-if="users.length === 0" class="users-empty">
        <div class="users-empty-icon">
          <i class="fa-solid fa-users" aria-hidden="true"></i>
        </div>

        <h3>No other users yet</h3>
      </div>

      <!-- USERS -->

      <div v-else class="users-list">
        <article v-for="user in users" :key="user.id" class="person-card">
          <!-- USER -->

          <router-link :to="`/profiles/${user.id}`" class="person-main">
            <UserAvatar
              :avatar-path="user.avatar_path"
              :name="`${user.first_name} ${user.last_name}`"
              class="users-avatar"
            />

            <div class="person-identity">
              <div class="person-name-row">
                <strong>
                  {{ user.first_name }}
                  {{ user.last_name }}
                </strong>

                <span
                  class="person-privacy"
                  :class="{
                    private: !user.is_public,
                  }"
                >
                  {{ user.is_public ? "Public" : "Private" }}
                </span>
              </div>

              <span v-if="user.nickname" class="person-nickname">
                {{ user.nickname }}
              </span>

              <span v-if="user.email" class="person-email">
                {{ user.email }}
              </span>
            </div>
          </router-link>

          <!-- ACTION -->

          <div class="person-actions">
            <button
              v-if="user.follow_status === 'none'"
              type="button"
              class="button button-primary"
              :disabled="changingUserId === user.id"
              @click="followUser(user.id)"
            >
              <span v-if="changingUserId === user.id"> Working... </span>

              <template v-else>
                <i class="fa-solid fa-user-plus" aria-hidden="true"></i>

                Follow
              </template>
            </button>

            <button
              v-else-if="user.follow_status === 'following'"
              type="button"
              class="button person-following-butto"
              :disabled="changingUserId === user.id"
              @click="unfollowUser(user.id)"
            >
              <span v-if="changingUserId === user.id"> Working... </span>

              <template v-else>
                <i class="fa-solid fa-user-check" aria-hidden="true"></i>

                Following
              </template>
            </button>

            <button
              v-else-if="user.follow_status === 'pending'"
              type="button"
              class="button person-pending-button"
              :disabled="changingUserId === user.id"
              title="Cancel follow request"
              @click="unfollowUser(user.id)"
            >
              {{
                changingUserId === user.id ? "Working..." : "Request pending"
              }}
            </button>
          </div>
        </article>
      </div>

      <div v-if="users.length > 0" class="users-pagination">
        <!-- LOADING MORE -->

        <div v-if="loadingMoreUsers" class="users-load-more">
          <span class="loading-spinner"></span>

          Loading more users...
        </div>

        <!-- ERROR -->

        <div v-else-if="usersLoadMoreError" class="users-load-more-error">
          <span>
            {{ usersLoadMoreError }}
          </span>

          <button
            type="button"
            class="button button-ghost button-small"
            @click="loadUsers()"
          >
            Try again
          </button>
        </div>

        <!-- SCROLL TARGET -->

        <div
          v-else-if="hasMoreUsers"
          ref="usersLoadTrigger"
          class="users-load-trigger"
          aria-hidden="true"
        ></div>

        <p v-else class="users-end-message">
          {{
            debouncedSearchQuery
              ? "End of search results"
              : "You've reached the end"
          }}
        </p>
      </div>
    </section>
  </main>
</template>

<script setup>
import { onBeforeUnmount, computed, onMounted, watch, ref } from "vue";
import { apiRequest } from "../services/api";

import UserAvatar from "../components/UserAvatar.vue";

const users = ref([]);
const followRequests = ref([]);

const loadingUsers = ref(false);
const loadingRequests = ref(false);

const usersError = ref("");
const requestsError = ref("");

const searchQuery = ref("");
const debouncedSearchQuery = ref("");

const USERS_PAGE_SIZE = 20;

const userOffset = ref(0);
const hasMoreUsers = ref(true);
const loadingMoreUsers = ref(false);
const usersLoadMoreError = ref("");
const usersLoadTrigger = ref(null);
let usersObserver = null;
let searchDebounceTimer = null;

let usersRequestVersion = 0;
const changingUserId = ref(null);
const changingRequestId = ref(null);

const filteredUsers = computed(() => {
  const query = searchQuery.value.trim().toLowerCase();

  if (!query) {
    return users.value;
  }

  return users.value.filter((user) => {
    const firstName = user.first_name || "";
    const lastName = user.last_name || "";
    const fullName = `${firstName} ${lastName}`;
    const nickname = user.nickname || "";
    const email = user.email || "";

    return [firstName, lastName, fullName, nickname, email].some((value) =>
      value.toLowerCase().includes(query),
    );
  });
});

function observeUsersTrigger(element) {
  if (usersObserver) {
    usersObserver.disconnect();
    usersObserver = null;
  }

  if (!element) {
    return;
  }

  usersObserver = new IntersectionObserver(
    (entries) => {
      const entry = entries[0];

      if (
        entry.isIntersecting &&
        hasMoreUsers.value &&
        !loadingUsers.value &&
        !loadingMoreUsers.value
      ) {
        loadUsers();
      }
    },
    {
      root: null,
      rootMargin: "250px 0px",
      threshold: 0,
    },
  );

  usersObserver.observe(element);
}

async function loadUsers(reset = false) {
  if (
    !reset &&
    (loadingUsers.value || loadingMoreUsers.value || !hasMoreUsers.value)
  ) {
    return;
  }

  if (reset) {
    usersRequestVersion += 1;
    users.value = [];
    userOffset.value = 0;
    hasMoreUsers.value = true;
    usersLoadMoreError.value = "";
  }

  const requestVersion = usersRequestVersion;
  const initialLoad = userOffset.value === 0;

  try {
    if (initialLoad) {
      loadingUsers.value = true;
      usersError.value = "";
    } else {
      loadingMoreUsers.value = true;
      usersLoadMoreError.value = "";
    }

    const params = new URLSearchParams();
    params.set("limit", String(USERS_PAGE_SIZE));
    params.set("offset", String(userOffset.value));

    if (debouncedSearchQuery.value) {
      params.set("q", debouncedSearchQuery.value);
    }

    const result = await apiRequest(`/users?${params.toString()}`);
    if (requestVersion !== usersRequestVersion) {
      return;
    }

    const incomingUsers = result.users || [];

    if (reset) {
      users.value = incomingUsers;
    } else {
      const existingIDs = new Set(users.value.map((user) => user.id));
      users.value.push(
        ...incomingUsers.filter((user) => !existingIDs.has(user.id)),
      );
    }

    userOffset.value =
      result.next_offset ?? userOffset.value + incomingUsers.length;

    hasMoreUsers.value = Boolean(result.has_more);
  } catch (err) {
    if (requestVersion !== usersRequestVersion) {
      return;
    }

    if (initialLoad) {
      usersError.value = err.message;
    } else {
      usersLoadMoreError.value = err.message;
    }
  } finally {
    if (requestVersion === usersRequestVersion) {
      loadingUsers.value = false;
      loadingMoreUsers.value = false;
    }
  }
}

async function loadFollowRequests() {
  try {
    loadingRequests.value = true;
    requestsError.value = "";
    followRequests.value = await apiRequest("/follow-requests");
  } catch (err) {
    requestsError.value = err.message;
  } finally {
    loadingRequests.value = false;
  }
}

async function followUser(userId) {
  try {
    changingUserId.value = userId;
    usersError.value = "";

    const result = await apiRequest(`/users/${userId}/follow`, {
      method: "POST",
    });

    const user = users.value.find((item) => item.id === userId);

    if (user) {
      user.follow_status = result.status;
    }
  } catch (err) {
    usersError.value = err.message;
  } finally {
    changingUserId.value = null;
  }
}

async function unfollowUser(userId) {
  try {
    changingUserId.value = userId;
    usersError.value = "";

    await apiRequest(`/users/${userId}/unfollow`, {
      method: "POST",
    });

    const user = users.value.find((item) => item.id === userId);

    if (user) {
      user.follow_status = "none";
      if (!user.is_public) {
        user.email = "";
      }
    }
  } catch (err) {
    usersError.value = err.message;
  } finally {
    changingUserId.value = null;
  }
}

async function acceptRequest(requestId) {
  try {
    changingRequestId.value = requestId;
    requestsError.value = "";

    await apiRequest(`/follow-requests/${requestId}/accept`, {
      method: "POST",
    });

    followRequests.value = followRequests.value.filter(
      (request) => request.id !== requestId,
    );
  } catch (err) {
    requestsError.value = err.message;
  } finally {
    changingRequestId.value = null;
  }
}

async function declineRequest(requestId) {
  try {
    changingRequestId.value = requestId;
    requestsError.value = "";

    await apiRequest(`/follow-requests/${requestId}/decline`, {
      method: "POST",
    });

    followRequests.value = followRequests.value.filter(
      (request) => request.id !== requestId,
    );
  } catch (err) {
    requestsError.value = err.message;
  } finally {
    changingRequestId.value = null;
  }
}

async function clearSearch() {
  if (searchDebounceTimer) {
    clearTimeout(searchDebounceTimer);
    searchDebounceTimer = null;
  }

  searchQuery.value = "";
  debouncedSearchQuery.value = "";
  await loadUsers(true);
}

watch(usersLoadTrigger, (element) => {
  observeUsersTrigger(element);
});

watch(searchQuery, (value) => {
  if (searchDebounceTimer) {
    clearTimeout(searchDebounceTimer);
  }

  searchDebounceTimer = setTimeout(async () => {
    debouncedSearchQuery.value = value.trim();
    await loadUsers(true);
  }, 300);
});

onMounted(async () => {
  await Promise.all([loadUsers(true), loadFollowRequests()]);
});

onBeforeUnmount(() => {
  if (usersObserver) {
    usersObserver.disconnect();
    usersObserver = null;
  }

  if (searchDebounceTimer) {
    clearTimeout(searchDebounceTimer);
    searchDebounceTimer = null;
  }
});
</script>

<style scoped>
.users-page {
  width: min(820px, 100%);
}

.users-header {
  margin-bottom: 26px;
}

.users-header h1 {
  margin-bottom: 7px;
}

.users-section-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 18px;
}

.users-section-heading h2 {
  margin-bottom: 4px;
}

.follow-requests-card {
  margin-bottom: 24px;
  padding: 20px;

  border: 1px solid var(--primary-border);
  border-radius: var(--radius-lg);

  background: rgba(79, 156, 255, 0.045);
  box-shadow: var(--shadow-sm);
}

.request-count {
  min-width: 26px;
  height: 26px;

  display: inline-flex;
  align-items: center;
  justify-content: center;

  padding: 0 7px;
  border-radius: var(--radius-round);
  background: var(--primary);

  color: white;
  font-size: 12px;
  font-weight: 750;
}

.follow-request-list {
  display: grid;
}

.follow-request-item {
  display: flex;
  align-items: center;
  justify-content: space-between;

  gap: 18px;
  margin: 0;
  padding: 14px 0;

  border: 0;
  border-top: 1px solid var(--border-soft);
  border-radius: 0;

  background: transparent;
}

.follow-request-item:last-child {
  padding-bottom: 0;
}

.request-user {
  min-width: 0;

  display: flex;
  align-items: center;
  gap: 12px;
  color: inherit;
}

.request-user:hover strong {
  color: var(--primary);
}

.request-user-info {
  min-width: 0;

  display: flex;
  flex-direction: column;
  gap: 2px;
}

.request-user-info strong {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.request-user-info span {
  color: var(--text-muted);
  font-size: 12px;
}

.request-actions {
  flex-shrink: 0;
  display: flex;
  gap: 8px;
}

.users-section {
  padding: 20px;
  border: 1px solid var(--border-soft);
  border-radius: var(--radius-lg);

  background: var(--surface);
  box-shadow: var(--shadow-sm);
}

.users-search {
  position: relative;
  margin-bottom: 18px;
}

.users-search input {
  padding-left: 42px;
  padding-right: 42px;

  background: var(--bg-secondary);
}

.users-search-icon {
  position: absolute;
  left: 14px;
  top: 50%;
  z-index: 1;

  color: var(--text-muted);
  font-size: 18px;
  pointer-events: none;
  transform: translateY(-50%);
}

.users-search-clear {
  position: absolute;

  right: 7px;
  top: 50%;

  width: 30px;
  height: 30px;
  min-height: 0;

  padding: 0;
  border: 0;
  background: transparent;
  color: var(--text-muted);
  font-size: 20px;
  transform: translateY(-50%);
}

.users-search-clear:hover {
  background: var(--surface-2);
  color: var(--text);
}

.users-list {
  display: grid;
}

.person-card {
  display: flex;
  align-items: center;
  justify-content: space-between;

  gap: 20px;

  margin: 0;
  padding: 16px 2px;

  border: 0;
  border-top: 1px solid var(--border-soft);

  border-radius: 0;
  background: transparent;
}

.person-card:hover {
  border-color: var(--border-soft);
  background: transparent;
}

.person-main {
  min-width: 0;
  flex: 1;
  display: flex;
  align-items: center;
  gap: 14px;
  color: inherit;
}

.users-avatar {
  width: 48px;
  height: 48px;

  flex: 0 0 48px;
  display: grid;
  place-items: center;
  border: 1px solid var(--primary-border);
  border-radius: 50%;
  background: var(--primary-soft);
  color: var(--primary);
  font-size: 13px;
  font-weight: 800;
  overflow: hidden;
}

.person-identity {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.person-name-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.person-name-row strong {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.person-main:hover .person-name-row strong {
  color: var(--primary);
}

.person-privacy {
  padding: 2px 7px;
  border: 1px solid rgba(54, 201, 143, 0.22);
  border-radius: var(--radius-round);
  background: rgba(54, 201, 143, 0.08);
  color: var(--success);
  font-size: 10px;
  font-weight: 700;
}

.person-privacy.private {
  border-color: rgba(148, 163, 184, 0.2);
  background: rgba(148, 163, 184, 0.07);
  color: var(--text-secondary);
}

.person-nickname,
.person-email {
  overflow: hidden;
  color: var(--text-muted);
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.person-actions {
  flex-shrink: 0;
}

.person-actions button {
  min-width: 118px;
}

.person-pending-button {
  border-color: rgba(244, 185, 66, 0.28);
  background: rgba(244, 185, 66, 0.08);
  color: var(--warning);
}

.person-pending-button:hover {
  border-color: rgba(255, 95, 109, 0.3);
  background: var(--danger-soft);

  color: var(--danger);
}

.users-pagination {
  width: 100%;
  padding: 14px 0 2px;
  text-align: center;
}

.users-load-more {
  min-height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: var(--text-muted);
  font-size: 12px;
}

.users-load-more-error {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-wrap: wrap;
  gap: 10px;
  padding: 10px;
  color: var(--danger);
  font-size: 12px;
}

.users-load-trigger {
  width: 100%;
  height: 2px;
}

.users-end-message {
  margin: 0;
  padding: 10px;
  color: var(--text-muted);
  font-size: 12px;
}

.users-state {
  min-height: 80px;

  display: flex;
  align-items: center;
  justify-content: center;
  gap: 9px;

  color: var(--text-muted);
  font-size: 13px;
}

.users-error {
  margin: 0;
  padding: 12px 14px;
  border: 1px solid rgba(255, 95, 109, 0.2);
  border-radius: var(--radius-md);

  background: var(--danger-soft);
  color: var(--danger);
  font-size: 13px;
}

.users-empty {
  padding: 34px 18px;
  text-align: center;
}

.users-empty-icon {
  width: 48px;
  height: 48px;

  display: grid;
  place-items: center;
  margin: 0 auto 12px;

  border-radius: 50%;

  background: var(--primary-soft);
  color: var(--primary);

  font-size: 20px;
}

.users-empty h3 {
  margin-bottom: 5px;
}

.users-empty p {
  margin-bottom: 14px;
  color: var(--text-muted);
  font-size: 13px;
}

@media (max-width: 620px) {
  .follow-requests-card,
  .users-section {
    padding: 16px;
  }

  .follow-request-item,
  .person-card {
    align-items: stretch;
    flex-direction: column;
  }

  .request-actions,
  .person-actions {
    width: 100%;
  }

  .request-actions button {
    flex: 1;
  }

  .person-actions button {
    width: 100%;
  }
}
</style>
