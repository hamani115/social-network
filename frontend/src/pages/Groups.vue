<template>
  <main class="groups-page">
    <header class="groups-header">
      <div>
        <h1>Groups</h1>
      </div>

      <button
        type="button"
        class="button-primary"
        @click="openCreateGroupModal"
      >
        <i class="fa-solid fa-plus" aria-hidden="true"></i>

        Create group
      </button>
    </header>

    <!-- PENDING INVITATIONS -->

    <section
      v-if="loadingInvitations || invitationsError || myInvitations.length > 0"
      class="group-invitations-card"
    >
      <div class="groups-section-heading">
        <h2>Group invitations</h2>

        <span v-if="myInvitations.length > 0" class="group-invitation-count">
          {{ myInvitations.length }}
        </span>
      </div>

      <div v-if="loadingInvitations" class="groups-state">
        <span class="loading-spinner"></span>

        Loading invitations...
      </div>

      <p v-else-if="invitationsError" class="groups-error">
        {{ invitationsError }}
      </p>

      <div v-else class="group-invitation-list">
        <article
          v-for="invitation in myInvitations"
          :key="invitation.id"
          class="group-invitation-item"
        >
          <router-link
            :to="`/groups/${invitation.group_id}`"
            class="group-invitation-main"
          >
            <div class="group-icon">
              {{ groupInitials(invitation.group_title) }}
            </div>

            <div class="group-invitation-info">
              <strong>
                {{ invitation.group_title }}
              </strong>

              <span>
                Invited by
                {{ invitation.inviter_name }}
              </span>
            </div>
          </router-link>

          <div class="group-invitation-actions">
            <button
              type="button"
              class="button button-ghost button-small"
              :disabled="changingInvitationId === invitation.id"
              @click="declineInvitation(invitation)"
            >
              Decline
            </button>

            <button
              type="button"
              class="button button-primary button-small"
              :disabled="changingInvitationId === invitation.id"
              @click="acceptInvitation(invitation)"
            >
              {{
                changingInvitationId === invitation.id ? "Working..." : "Accept"
              }}
            </button>
          </div>
        </article>
      </div>
    </section>

    <section   class="groups-directory">
      <div class="groups-section-heading">
        <h2>Browse groups</h2>
      </div>

      <!-- SEARCH -->
      <div class="group-search">
        <span class="group-search-icon" aria-hidden="true">
          <i class="fa-solid fa-magnifying-glass"></i>
        </span>

        <label for="group-search" class="visually-hidden">
          Search groups
        </label>

        <input
          id="group-search"
          v-model="searchQuery"
          type="search"
          placeholder="Search groups by title or description..."
          autocomplete="off"
        />

        <button
          v-if="searchQuery"
          type="button"
          class="group-search-clear"
          aria-label="Clear group search"
          @click="clearSearch"
        >
          <i class="fa-solid fa-xmark" aria-hidden="true"></i>
        </button>
      </div>

      <div v-if="loadingGroups" class="groups-state">
        <span class="loading-spinner"></span>

        Loading groups...
      </div>

      <p v-else-if="groupsError" class="groups-error">
        {{ groupsError }}
      </p>

      <div
        v-else-if="groups.length === 0 && debouncedSearchQuery"
        class="groups-empty"
      >
        <div class="groups-empty-icon">
          <i class="fa-solid fa-magnifying-glass" aria-hidden="true"></i>
        </div>

        <h3>No groups found</h3>

        <p>No groups match "{{ debouncedSearchQuery }}"</p>
      </div>

      <div v-else-if="groups.length === 0" class="groups-empty">
        <div class="groups-empty-icon">
          <i class="fa-solid fa-user-group" aria-hidden="true"></i>
        </div>

        <h3>No groups yet</h3>
      </div>

      <!-- GROUPS -->

      <div v-else class="groups-grid">
        <article v-for="group in groups" :key="group.id" class="group-card">
          <div class="group-card-top">
            <router-link :to="`/groups/${group.id}`" class="group-card-icon">
              {{ groupInitials(group.title) }}
            </router-link>

            <div class="group-card-badges">
              <span
                v-if="websocket.groupUnreadForGroup(group.id) > 0"
                class="group-unread-badge"
              >
                {{ websocket.groupUnreadForGroup(group.id) }}
                new
              </span>

              <span
                v-if="membershipLabel(group.membership_status) !== ''"
                class="group-status-badge"
                :class="`status-${group.membership_status}`"
              >
                {{ membershipLabel(group.membership_status) }}
              </span>
            </div>
          </div>

          <router-link :to="`/groups/${group.id}`" class="group-card-title">
            {{ group.title }}
          </router-link>

          <p v-if="group.description" class="group-card-description">
            {{ group.description }}
          </p>

          <div class="group-card-meta">
            <span>
              Created by
              <strong>
                {{ group.creator_name }}
              </strong>
            </span>

            <span>
              {{ group.member_count }}
              {{ group.member_count === 1 ? "member" : "members" }}
            </span>
          </div>

          <div class="group-card-actions">
            <!-- NOT MEMBER -->

            <button
              v-if="group.membership_status === 'none'"
              type="button"
              class="button-primary"
              :disabled="changingGroupId === group.id"
              @click="requestJoinGroup(group)"
            >
              {{
                changingGroupId === group.id ? "Working..." : "Request to join"
              }}
            </button>

            <button
              v-else-if="group.membership_status === 'pending'"
              type="button"
              class="button group-pending-button"
              :disabled="changingGroupId === group.id"
              @click="cancelJoinRequest(group)"
            >
              {{
                changingGroupId === group.id ? "Working..." : "Request pending"
              }}
            </button>

            <span
              v-else-if="group.membership_status === 'invited'"
              class="group-invited-message"
            >
              Invitation waiting above
            </span>

            <!-- MEMBER + OWNER -->
            <router-link
              v-else
              :to="`/groups/${group.id}`"
              class="button button-primary group-open-button"
            >
              Open group
            </router-link>
          </div>
        </article>
      </div>

      <!-- PAGINATION -->
      <div v-if="groups.length > 0" class="groups-pagination">
        <div v-if="loadingMoreGroups" class="groups-load-more">
          <span class="loading-spinner"></span>

          Loading more groups...
        </div>

        <div v-else-if="groupsLoadMoreError" class="groups-load-more-error">
          <span>
            {{ groupsLoadMoreError }}
          </span>

          <button
            type="button"
            class="button button-ghost button-small"
            @click="loadGroups()"
          >
            Try again
          </button>
        </div>

        <div
          v-else-if="hasMoreGroups"
          ref="groupsLoadTrigger"
          class="groups-load-trigger"
          aria-hidden="true"
        ></div>

        <p v-else class="groups-end-message">
          {{
            debouncedSearchQuery
              ? "End of search results"
              : "You've reached the end"
          }}
        </p>
      </div>
    </section>

    <Teleport to="body">
      <div
        v-if="createGroupModalOpen"
        class="group-modal-overlay"
        @click.self="closeCreateGroupModal"
      >
        <section
          class="group-modal"
          role="dialog"
          aria-modal="true"
          aria-labelledby="create-group-title"
        >
          <header class="group-modal-header">
            <h2 id="create-group-title">Create group</h2>

            <button
              type="button"
              class="group-modal-close"
              aria-label="Close create group"
              :disabled="creatingGroup"
              @click="closeCreateGroupModal"
            >
              <i class="fa-solid fa-xmark" aria-hidden="true"></i>
            </button>
          </header>

          <form class="group-modal-form" @submit.prevent="createGroup">
            <div class="group-modal-body">
              <div>
                <label for="group-title"> Group title </label>

                <input
                  id="group-title"
                  v-model="newGroup.title"
                  type="text"
                  placeholder="Enter a group title"
                  required
                />
              </div>

              <div>
                <label for="group-description"> Description </label>

                <textarea
                  id="group-description"
                  v-model="newGroup.description"
                  rows="5"
                  placeholder="What is this group about?"
                ></textarea>
              </div>

              <p v-if="createError" class="groups-error">
                {{ createError }}
              </p>
            </div>

            <footer class="group-modal-footer">
              <button
                type="button"
                class="button button-ghost"
                :disabled="creatingGroup"
                @click="closeCreateGroupModal"
              >
                Cancel
              </button>

              <button
                type="submit"
                class="button-primary"
                :disabled="creatingGroup"
              >
                {{ creatingGroup ? "Creating..." : "Create group" }}
              </button>
            </footer>
          </form>
        </section>
      </div>
    </Teleport>
  </main>
</template>

<script setup>
import { onBeforeUnmount, onMounted, ref, watch } from "vue";
import { apiRequest } from "../services/api";
import { useWebSocketStore } from "../stores/websocket";
const websocket = useWebSocketStore();
const groups = ref([]);
const myInvitations = ref([]);
const loadingGroups = ref(false);
const loadingMoreGroups = ref(false);
const groupsError = ref("");
const groupsLoadMoreError = ref("");
const GROUPS_PAGE_SIZE = 20;
const groupOffset = ref(0);
const hasMoreGroups = ref(true);
const groupsLoadTrigger = ref(null);
let groupsObserver = null;
const searchQuery = ref("");
const debouncedSearchQuery = ref("");
let searchDebounceTimer = null;
let groupsRequestVersion = 0;
const loadingInvitations = ref(false);
const invitationsError = ref("");
const changingInvitationId = ref(null);
const changingGroupId = ref(null);
const createGroupModalOpen = ref(false);
const creatingGroup = ref(false);
const createError = ref("");
const newGroup = ref({
  title: "",
  description: "",
});

function groupInitials(title) {
  if (!title) {
    return "G";
  }
  const parts = title.trim().split(/\s+/);
  if (parts.length === 1) {
    return parts[0].charAt(0).toUpperCase();
  }
  return (parts[0].charAt(0) + parts[1].charAt(0)).toUpperCase();
}

function membershipLabel(status) {
  switch (status) {
    case "owner":
      return "Owner";
    case "member":
      return "Member";
    case "pending":
      return "Pending";
    case "invited":
      return "Invited";
    default:
      return "";
  }
}

async function loadGroups(reset = false) {
  if (
    !reset &&
    (loadingGroups.value || loadingMoreGroups.value || !hasMoreGroups.value)
  ) {
    return;
  }
  if (reset) {
    groupsRequestVersion += 1;
    groups.value = [];
    groupOffset.value = 0;
    hasMoreGroups.value = true;
    groupsLoadMoreError.value = "";
  }
  const requestVersion = groupsRequestVersion;
  const initialLoad = groupOffset.value === 0;
  try {
    if (initialLoad) {
      loadingGroups.value = true;
      groupsError.value = "";
    } else {
      loadingMoreGroups.value = true;
      groupsLoadMoreError.value = "";
    }
    const params = new URLSearchParams();
    params.set("limit", String(GROUPS_PAGE_SIZE));
    params.set("offset", String(groupOffset.value));
    if (debouncedSearchQuery.value) {
      params.set("q", debouncedSearchQuery.value);
    }
    const result = await apiRequest(`/groups?${params.toString()}`);
    if (requestVersion !== groupsRequestVersion) {
      return;
    }
    const incomingGroups = result.groups || [];
    if (reset) {
      groups.value = incomingGroups;
    } else {
      const existingIDs = new Set(groups.value.map((group) => group.id));
      groups.value.push(
        ...incomingGroups.filter((group) => !existingIDs.has(group.id)),
      );
    }
    groupOffset.value =
      result.next_offset ?? groupOffset.value + incomingGroups.length;
    hasMoreGroups.value = Boolean(result.has_more);
  } catch (err) {
    if (requestVersion !== groupsRequestVersion) {
      return;
    }
    if (initialLoad) {
      groupsError.value = err.message;
    } else {
      groupsLoadMoreError.value = err.message;
    }
  } finally {
    if (requestVersion === groupsRequestVersion) {
      loadingGroups.value = false;
      loadingMoreGroups.value = false;
    }
  }
}

watch(searchQuery, (value) => {
  if (searchDebounceTimer) {
    clearTimeout(searchDebounceTimer);
  }
  searchDebounceTimer = setTimeout(async () => {
    debouncedSearchQuery.value = value.trim();
    await loadGroups(true);
  }, 300);
});

async function clearSearch() {
  if (searchDebounceTimer) {
    clearTimeout(searchDebounceTimer);
    searchDebounceTimer = null;
  }
  searchQuery.value = "";
  debouncedSearchQuery.value = "";
  await loadGroups(true);
}

function observeGroupsTrigger(element) {
  if (groupsObserver) {
    groupsObserver.disconnect();
    groupsObserver = null;
  }
  if (!element) {
    return;
  }
  groupsObserver = new IntersectionObserver(
    (entries) => {
      const entry = entries[0];
      if (
        entry.isIntersecting &&
        hasMoreGroups.value &&
        !loadingGroups.value &&
        !loadingMoreGroups.value
      ) {
        loadGroups();
      }
    },
    {
      root: null,
      rootMargin: "250px 0px",
      threshold: 0,
    },
  );
  groupsObserver.observe(element);
}

watch(groupsLoadTrigger, (element) => {
  observeGroupsTrigger(element);
});

async function loadMyInvitations() {
  try {
    loadingInvitations.value = true;
    invitationsError.value = "";
    myInvitations.value = await apiRequest("/group-invitations");
  } catch (err) {
    invitationsError.value = err.message;
  } finally {
    loadingInvitations.value = false;
  }
}

async function acceptInvitation(invitation) {
  try {
    changingInvitationId.value = invitation.id;
    invitationsError.value = "";
    await apiRequest(`/group-invitations/${invitation.id}/accept`, {
      method: "POST",
    });
    myInvitations.value = myInvitations.value.filter(
      (item) => item.id !== invitation.id,
    );
    const group = groups.value.find((item) => item.id === invitation.group_id);
    if (group) {
      group.membership_status = "member";
      group.member_count += 1;
    }
  } catch (err) {
    invitationsError.value = err.message;
  } finally {
    changingInvitationId.value = null;
  }
}

async function declineInvitation(invitation) {
  try {
    changingInvitationId.value = invitation.id;
    invitationsError.value = "";
    await apiRequest(`/group-invitations/${invitation.id}/decline`, {
      method: "POST",
    });
    myInvitations.value = myInvitations.value.filter(
      (item) => item.id !== invitation.id,
    );
    const group = groups.value.find((item) => item.id === invitation.group_id);
    if (group && group.membership_status === "invited") {
      group.membership_status = "none";
    }
  } catch (err) {
    invitationsError.value = err.message;
  } finally {
    changingInvitationId.value = null;
  }
}

async function requestJoinGroup(group) {
  try {
    changingGroupId.value = group.id;
    groupsError.value = "";
    const result = await apiRequest(`/groups/${group.id}/join-request`, {
      method: "POST",
    });
    group.membership_status = result.status || "pending";
  } catch (err) {
    groupsError.value = err.message;
  } finally {
    changingGroupId.value = null;
  }
}

async function cancelJoinRequest(group) {
  try {
    changingGroupId.value = group.id;
    groupsError.value = "";
    await apiRequest(`/groups/${group.id}/cancel-join-request`, {
      method: "POST",
    });
    group.membership_status = "none";
  } catch (err) {
    groupsError.value = err.message;
  } finally {
    changingGroupId.value = null;
  }
}

function openCreateGroupModal() {
  createError.value = "";
  createGroupModalOpen.value = true;
}

function resetCreateGroupForm() {
  newGroup.value.title = "";
  newGroup.value.description = "";
}

function closeCreateGroupModal() {
  if (creatingGroup.value) {
    return;
  }
  createGroupModalOpen.value = false;
  createError.value = "";
  resetCreateGroupForm();
}

async function createGroup() {
  try {
    creatingGroup.value = true;
    createError.value = "";
    await apiRequest("/groups", {
      method: "POST",

      body: JSON.stringify({
        title: newGroup.value.title,
        description: newGroup.value.description,
      }),
    });
    resetCreateGroupForm();
    createGroupModalOpen.value = false;
    if (searchDebounceTimer) {
      clearTimeout(searchDebounceTimer);
      searchDebounceTimer = null;
    }
    searchQuery.value = "";
    debouncedSearchQuery.value = "";
    await loadGroups(true);
  } catch (err) {
    createError.value = err.message;
  } finally {
    creatingGroup.value = false;
  }
}

onMounted(async () => {
  await Promise.all([loadGroups(true), loadMyInvitations()]);
});
onBeforeUnmount(() => {
  if (searchDebounceTimer) {
    clearTimeout(searchDebounceTimer);
  }
  if (groupsObserver) {
    groupsObserver.disconnect();
  }
});
</script>

<style scoped>
.groups-page {
  width: min(1000px, 100%);
}

.groups-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;

  gap: 24px;

  margin-bottom: 26px;
}

.groups-header h1 {
  margin-bottom: 7px;
}

.groups-section-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;

  gap: 16px;

  margin-bottom: 18px;
}

.groups-section-heading h2 {
  margin-bottom: 4px;
}

.group-invitations-card {
  margin-bottom: 24px;

  padding: 20px;

  border: 1px solid var(--primary-border);

  border-radius: var(--radius-lg);

  background: rgba(79, 156, 255, 0.045);

  box-shadow: var(--shadow-sm);
}

.group-invitation-count {
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

.group-invitation-list {
  display: grid;
}

.group-invitation-item {
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

.group-invitation-main {
  min-width: 0;

  display: flex;
  align-items: center;

  gap: 12px;

  color: inherit;
}

.group-invitation-info {
  min-width: 0;

  display: flex;
  flex-direction: column;

  gap: 3px;
}

.group-invitation-info strong {
  overflow: hidden;

  text-overflow: ellipsis;
  white-space: nowrap;
}

.group-invitation-info span {
  color: var(--text-muted);
  font-size: 12px;
}

.group-invitation-actions {
  flex-shrink: 0;
  display: flex;
  gap: 8px;
}

.groups-directory {
  padding: 20px;
  border: 1px solid var(--border-soft);
  border-radius: var(--radius-lg);
  background: var(--surface);
  box-shadow: var(--shadow-sm);
}

.group-search {
  position: relative;

  margin-bottom: 20px;
}

.group-search input {
  padding-left: 42px;
  padding-right: 42px;

  background: var(--bg-secondary);
}

.group-search-icon {
  position: absolute;
  left: 14px;
  top: 50%;
  z-index: 1;

  color: var(--text-muted);
  font-size: 14px;

  pointer-events: none;
  transform: translateY(-50%);
}

.group-search-clear {
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
  font-size: 14px;

  transform: translateY(-50%);
}

.group-search-clear:hover {
  background: var(--surface-2);
  color: var(--text);
}

.groups-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.group-card {
  min-width: 0;

  display: flex;
  flex-direction: column;

  gap: 13px;
  margin: 0;
  padding: 18px;

  border: 1px solid var(--border-soft);
  border-radius: var(--radius-lg);
  background: var(--bg-secondary);

  box-shadow: none;
}

.group-card:hover {
  border-color: var(--primary-border);
  background: var(--surface-2);
}

.group-card-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;

  gap: 12px;
}

.group-icon,
.group-card-icon {
  width: 44px;
  height: 44px;

  flex: 0 0 44px;

  display: grid;
  place-items: center;

  border: 1px solid var(--primary-border);
  border-radius: 12px;
  background: var(--primary-soft);

  color: var(--primary);

  font-size: 13px;
  font-weight: 800;
}

.group-card-badges {
  display: flex;
  align-items: center;
  flex-wrap: wrap;

  gap: 6px;
}

.group-status-badge,
.group-unread-badge {
  display: inline-flex;
  align-items: center;

  min-height: 23px;
  padding: 2px 8px;

  border-radius: var(--radius-round);

  font-size: 10px;
  font-weight: 750;
}

.group-status-badge {
  border: 1px solid var(--border);
  /* background: var(--surface);
  color: var(--text-secondary); */
}

.status-owner {
  border-color: var(--primary-border);
  background: var(--primary-soft);
  color: var(--primary);
}

.status-member {
  border-color: rgba(54, 201, 143, 0.24);
  background: rgba(54, 201, 143, 0.08);
  color: var(--success);
}

.status-pending {
  border-color: rgba(244, 185, 66, 0.25);
  background: rgba(244, 185, 66, 0.08);
  color: var(--warning);
}

.status-invited {
  border-color: var(--primary-border);
  background: var(--primary-soft);
  color: var(--primary);
}

.group-unread-badge {
  background: var(--primary);
  color: white;
}

.group-card-title {
  width: fit-content;

  color: var(--text);

  font-size: 17px;
  font-weight: 750;
}

.group-card-title:hover {
  color: var(--primary);
}

.group-card-description {
  min-height: 42px;

  margin: 0;

  overflow: hidden;

  color: var(--text-secondary);

  font-size: 13px;
  line-height: 1.55;

  display: -webkit-box;

  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.group-card-meta {
  display: flex;
  flex-direction: column;

  gap: 4px;

  color: var(--text-muted);

  font-size: 11px;
}

.group-card-meta strong {
  color: var(--text-secondary);
}

.group-card-actions {
  margin-top: auto;

  padding-top: 5px;
}

.group-card-actions > * {
  width: 100%;
}

.group-open-button {
  display: flex;
  align-items: center;
  justify-content: center;
}

.group-pending-button {
  border-color: rgba(244, 185, 66, 0.28);

  background: rgba(244, 185, 66, 0.08);

  color: var(--warning);
}

.group-pending-button:hover {
  border-color: rgba(255, 95, 109, 0.3);

  background: var(--danger-soft);

  color: var(--danger);
}

.group-invited-message {
  min-height: 40px;

  display: flex;
  align-items: center;
  justify-content: center;

  border: 1px solid var(--primary-border);

  border-radius: var(--radius-md);

  background: var(--primary-soft);

  color: var(--primary);

  font-size: 12px;
  font-weight: 650;
}

.groups-state {
  min-height: 80px;

  display: flex;
  align-items: center;
  justify-content: center;

  gap: 9px;

  color: var(--text-muted);

  font-size: 13px;
}

.groups-error {
  margin: 0;

  padding: 12px 14px;

  border: 1px solid rgba(255, 95, 109, 0.2);

  border-radius: var(--radius-md);

  background: var(--danger-soft);

  color: var(--danger);

  font-size: 13px;
}

.groups-empty {
  padding: 36px 18px;

  text-align: center;
}

.groups-empty-icon {
  width: 48px;
  height: 48px;

  display: grid;
  place-items: center;

  margin: 0 auto 12px;

  border-radius: 50%;

  background: var(--primary-soft);

  color: var(--primary);

  font-size: 18px;
}

.groups-empty h3 {
  margin-bottom: 5px;
}

.groups-empty p {
  margin-bottom: 14px;

  color: var(--text-muted);

  font-size: 13px;
}

.groups-pagination {
  padding: 18px 0 2px;

  text-align: center;
}

.groups-load-more {
  min-height: 44px;

  display: flex;
  align-items: center;
  justify-content: center;

  gap: 8px;

  color: var(--text-muted);

  font-size: 12px;
}

.groups-load-more-error {
  display: flex;
  align-items: center;
  justify-content: center;

  flex-wrap: wrap;

  gap: 10px;

  color: var(--danger);

  font-size: 12px;
}

.groups-load-trigger {
  height: 2px;
}

.groups-end-message {
  margin: 0;

  padding: 10px;

  color: var(--text-muted);

  font-size: 12px;
}

.group-modal-overlay {
  position: fixed;
  inset: 0;

  z-index: 2000;

  display: flex;
  align-items: center;
  justify-content: center;

  padding: 24px;

  background: rgba(3, 7, 18, 0.76);

  backdrop-filter: blur(5px);
}

.group-modal {
  width: min(580px, 100%);

  max-height: calc(100vh - 48px);

  overflow-y: auto;

  border: 1px solid rgba(255, 255, 255, 0.09);

  border-radius: var(--radius-xl);

  background: var(--surface);

  box-shadow: 0 24px 70px rgba(0, 0, 0, 0.55);
}

.group-modal-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;

  gap: 20px;

  padding: 20px 22px;

  border-bottom: 1px solid var(--border-soft);
}

.group-modal-header h2 {
  margin: 0;
}

.group-modal-close {
  width: 36px;
  height: 36px;
  min-height: 0;

  padding: 0;

  border: none;
  border-radius: 50%;

  background: transparent;

  color: var(--text-secondary);

  font-size: 16px;
}

.group-modal-close:hover {
  background: var(--surface-2);

  color: var(--text);
}

.group-modal-body {
  display: flex;
  flex-direction: column;

  gap: 17px;

  padding: 22px;
}

.group-modal-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;

  gap: 8px;

  padding: 16px 22px;

  border-top: 1px solid var(--border-soft);
}

@media (max-width: 700px) {
  .groups-header {
    flex-direction: column;
  }

  .groups-header > button {
    width: 100%;
  }

  .groups-grid {
    grid-template-columns: 1fr;
  }

  .group-invitations-card,
  .groups-directory {
    padding: 16px;
  }

  .group-invitation-item {
    align-items: stretch;

    flex-direction: column;
  }

  .group-invitation-actions {
    width: 100%;
  }

  .group-invitation-actions button {
    flex: 1;
  }

  .group-modal-overlay {
    padding: 12px;
  }

  .group-modal {
    max-height: calc(100vh - 24px);
  }

  .group-modal-footer button {
    flex: 1;
  }
}
</style>
