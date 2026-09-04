<template>
  <main class="group-detail-page">
    <div v-if="loadingGroup" class="group-page-state">
      <span class="loading-spinner"></span>
      Loading group...
    </div>

    <p v-else-if="groupError && !group" class="group-page-error">
      {{ groupError }}
    </p>

    <template v-else-if="group">
      <section class="group-hero">
        <div class="group-hero-icon">
          {{ group.title?.charAt(0)?.toUpperCase() || "G" }}
        </div>

        <div class="group-hero-content">
          <div class="group-hero-title-row">
            <h1>
              {{ group.title }}
            </h1>

            <span
              class="group-membership-badge"
              :class="`status-${group.membership_status}`"
            >
              {{
                group.membership_status === "owner"
                  ? "Owner"
                  : group.membership_status === "member"
                    ? "Member"
                    : group.membership_status === "pending"
                      ? "Request pending"
                      : group.membership_status === "invited"
                        ? "Invited"
                        : "Not a member"
              }}
            </span>
          </div>

          <p v-if="group.description" class="group-description">
            {{ group.description }}
          </p>

          <div class="group-summary">
            <span>
              Created by
              <strong>
                {{ group.creator_name }}
              </strong>
            </span>

            <span class="group-summary-divider"> - </span>

            <span>
              <strong>
                {{ group.member_count }}
              </strong>

              {{ group.member_count === 1 ? "member" : "members" }}
            </span>
          </div>

          <!-- REQUEST JOIN -->

          <div v-if="!isMemberOrOwner" class="group-join-actions">
            <button
              v-if="group.membership_status === 'none'"
              type="button"
              class="button-primary"
              @click="requestJoinGroup"
            >
              Request to join
            </button>

            <button
              v-else-if="group.membership_status === 'pending'"
              type="button"
              class="button group-pending-button"
              @click="cancelJoinRequest"
            >
              Cancel join request
            </button>

            <div
              v-else-if="group.membership_status === 'invited'"
              class="group-invited-notice"
            >
              You have a pending invitation to this group.

              <router-link to="/groups"> View invitations </router-link>
            </div>
          </div>

          <!-- LEAVE GROUP -->
          <div
            v-if="group.membership_status === 'member'"
            class="group-member-actions"
          >
            <button
              type="button"
              class="button group-leave-button"
              :disabled="leavingGroup"
              @click="leaveGroup"
            >
              {{ leavingGroup ? "Leaving..." : "Leave group" }}
            </button>
          </div>
        </div>
      </section>

      <!-- GROUP NAVIGATION -->

      <nav
        v-if="isMemberOrOwner"
        class="group-tabs"
        aria-label="Group sections"
      >
        <button
          type="button"
          class="group-tab"
          :class="{
            active: activeGroupTab === 'overview',
          }"
          @click="selectGroupTab('overview')"
        >
          Overview
        </button>

        <button
          type="button"
          class="group-tab"
          :class="{
            active: activeGroupTab === 'posts',
          }"
          @click="selectGroupTab('posts')"
        >
          Posts
        </button>

        <button
          type="button"
          class="group-tab"
          :class="{
            active: activeGroupTab === 'events',
          }"
          @click="selectGroupTab('events')"
        >
          Events
        </button>

        <button
          type="button"
          class="group-tab"
          :class="{
            active: activeGroupTab === 'members',
          }"
          @click="selectGroupTab('members')"
        >
          Members
        </button>

        <button
          type="button"
          class="group-tab"
          :class="{
            active: activeGroupTab === 'chat',
          }"
          @click="selectGroupTab('chat')"
        >
          Chat

          <span
            v-if="websocket.groupUnreadForGroup(Number(groupId)) > 0"
            class="group-tab-badge"
          >
            {{ websocket.groupUnreadForGroup(Number(groupId)) }}
          </span>
        </button>

        <button
          v-if="isOwner"
          type="button"
          class="group-tab"
          :class="{
            active: activeGroupTab === 'management',
          }"
          @click="selectGroupTab('management')"
        >
          Management

          <span v-if="joinRequests.length > 0" class="group-tab-badge">
            {{ joinRequests.length }}
          </span>
        </button>
      </nav>

      <!-- ABOUT -->

      <section
        v-if="!isMemberOrOwner || activeGroupTab === 'overview'"
        class="group-tab-panel"
      >
        <div class="group-panel-heading">
          <div>
            <h2>About</h2>
          </div>
        </div>

        <div class="group-overview-grid">
          <div class="group-overview-item">
            <span> Creator </span>

            <strong>
              {{ group.creator_name }}
            </strong>
          </div>

          <div class="group-overview-item">
            <span> Members </span>

            <strong>
              {{ group.member_count }}
            </strong>
          </div>

          <div class="group-overview-item">
            <span> Your role </span>

            <strong>
              {{ group.membership_status }}
            </strong>
          </div>

          <div class="group-overview-item">
            <span> Created </span>

            <strong>
              {{ formatDate(group.created_at) }}
            </strong>
          </div>
        </div>
      </section>

      <!-- POSTS -->

      <GroupPostsTab
        v-if="isMemberOrOwner"
        :group-id="groupId"
        :active="activeGroupTab === 'posts'"
      />

      <!-- EVENTS -->

      <GroupEventsTab
        v-if="isMemberOrOwner"
        :group-id="groupId"
        :active="activeGroupTab === 'events'"
      />

      <!-- MEMBERS -->
      <GroupMembersTab
        v-if="isMemberOrOwner"
        :group-id="groupId"
        :member-count="group.member_count"
        :active="activeGroupTab === 'members'"
        @invitation-sent="handleInvitationSent"
      />

      <!-- CHAT -->

      <GroupChatTab
        v-if="isMemberOrOwner"
        :group-id="groupId"
        :active="activeGroupTab === 'chat'"
      />

      <!-- MANAGEMENT -->
      <section
        v-if="isOwner && activeGroupTab === 'management'"
        class="group-tab-panel"
      >
        <div class="group-panel-heading">
          <div>
            <h2>Group management</h2>
          </div>
        </div>

        <section class="management-section">
          <div class="management-heading">
            <h3>Pending join requests</h3>

            <span v-if="joinRequests.length" class="group-tab-badge">
              {{ joinRequests.length }}
            </span>
          </div>

          <p v-if="loadingJoinRequests" class="group-section-state">
            Loading join requests...
          </p>

          <p v-if="joinRequestsError" class="group-page-error">
            {{ joinRequestsError }}
          </p>

          <p
            v-if="!loadingJoinRequests && joinRequests.length === 0"
            class="group-empty-state"
          >
            No pending join requests.
          </p>

          <article
            v-for="request in joinRequests"
            :key="request.id"
            class="management-request"
          >
            <div class="management-person">
              <UserAvatar
                :avatar-path="request.requester_avatar_path"
                :name="request.requester_name"
                class="management-avatar"
              />

              <div class="management-person-info">
                <strong>
                  {{ request.requester_name }}
                </strong>

                <span v-if="request.requester_nickname">
                  {{ request.requester_nickname }}
                </span>
              </div>
            </div>

            <div class="management-actions">
              <button
                type="button"
                class="button button-ghost"
                @click="declineJoinRequest(request.id)"
              >
                Decline
              </button>

              <button
                type="button"
                class="button-primary"
                @click="acceptJoinRequest(request.id)"
              >
                Accept
              </button>
            </div>
          </article>
        </section>

        <section class="management-section">
          <h3>Invitations sent</h3>

          <p v-if="loadingInvitations" class="group-section-state">
            Loading invitations...
          </p>

          <p v-if="invitationsError" class="group-page-error">
            {{ invitationsError }}
          </p>

          <p
            v-if="!loadingInvitations && groupInvitations.length === 0"
            class="group-empty-state"
          >
            No invitations found.
          </p>

          <article
            v-for="invitation in groupInvitations"
            :key="invitation.id"
            class="management-invitation"
          >
            <div class="management-person">
              <UserAvatar
                :avatar-path="invitation.invitee_avatar_path"
                :name="invitation.invitee_name"
                class="management-avatar"
              />

              <div class="management-person-info">
                <strong>
                  {{ invitation.invitee_name }}
                </strong>

                <span v-if="invitation.invitee_nickname">
                  {{ invitation.invitee_nickname }}
                </span>
              </div>
            </div>

            <span class="invitation-status">
              {{ invitation.status }}
            </span>
          </article>
        </section>
      </section>
    </template>
  </main>
</template>

<script setup>
import { computed, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiRequest } from "../services/api";
import { useWebSocketStore } from "../stores/websocket";
import { formatDate } from "../utils/date";
import GroupPostsTab from "../components/group/GroupPostsTab.vue";
import GroupEventsTab from "../components/group/GroupEventsTab.vue";
import GroupMembersTab from "../components/group/GroupMembersTab.vue";
import GroupChatTab from "../components/group/GroupChatTab.vue";
import UserAvatar from "../components/UserAvatar.vue";
const route = useRoute();
// const auth = useAuthStore();
const websocket = useWebSocketStore();
const group = ref(null);
const joinRequests = ref([]);
const groupInvitations = ref([]);
const loadingGroup = ref(false);
const loadingJoinRequests = ref(false);
const loadingInvitations = ref(false);
const leavingGroup = ref(false);
const groupError = ref("");
const joinRequestsError = ref("");
const invitationsError = ref("");
const activeGroupTab = ref("overview");
const loadedGroupTabs = ref({
  management: false,
});
const groupId = computed(() => route.params.id);
const isOwner = computed(() => {
  return group.value?.membership_status === "owner";
});
const isMemberOrOwner = computed(() => {
  return (
    group.value?.membership_status === "owner" ||
    group.value?.membership_status === "member"
  );
});

async function handleInvitationSent() {
  if (isOwner.value) {
    await loadGroupInvitations();
  }
}

async function loadGroup() {
  try {
    loadingGroup.value = true;
    groupError.value = "";
    group.value = await apiRequest(`/groups/${groupId.value}`);
    if (isOwner.value) {
      await loadJoinRequests();
    }
  } catch (err) {
    groupError.value = err.message;
  } finally {
    loadingGroup.value = false;
  }
}

async function selectGroupTab(tab) {
  if (tab !== "overview" && !isMemberOrOwner.value) {
    return;
  }
  if (tab === "management" && !isOwner.value) {
    return;
  }
  activeGroupTab.value = tab;
  switch (tab) {
    case "management":
      if (!loadedGroupTabs.value.management) {
        await Promise.all([loadJoinRequests(), loadGroupInvitations()]);
        loadedGroupTabs.value.management = true;
      }
      break;
  }
}

async function requestJoinGroup() {
  try {
    groupError.value = "";
    const result = await apiRequest(`/groups/${groupId.value}/join-request`, {
      method: "POST",
    });
    group.value.membership_status = result.status || "pending";
  } catch (err) {
    groupError.value = err.message;
  }
}

async function cancelJoinRequest() {
  try {
    groupError.value = "";
    await apiRequest(`/groups/${groupId.value}/cancel-join-request`, {
      method: "POST",
    });
    group.value.membership_status = "none";
  } catch (err) {
    groupError.value = err.message;
  }
}

async function leaveGroup() {
  if (leavingGroup.value || group.value?.membership_status !== "member") {
    return;
  }
  const confirmed = window.confirm(
    `Leave "${group.value.title}"?\n\n` +
      "You will lose access to the group's posts, events, members and chat",
  );
  if (!confirmed) {
    return;
  }
  try {
    leavingGroup.value = true;
    groupError.value = "";
    await apiRequest(`/groups/${groupId.value}/leave`, {
      method: "POST",
    });
    activeGroupTab.value = "overview";
    loadedGroupTabs.value = {
      management: false,
    };
    await loadGroup();
  } catch (err) {
    groupError.value = err.message;
  } finally {
    leavingGroup.value = false;
  }
}

async function loadJoinRequests() {
  try {
    loadingJoinRequests.value = true;
    joinRequestsError.value = "";
    joinRequests.value = await apiRequest(
      `/groups/${groupId.value}/join-requests`,
    );
  } catch (err) {
    joinRequestsError.value = err.message;
  } finally {
    loadingJoinRequests.value = false;
  }
}

async function acceptJoinRequest(requestId) {
  try {
    joinRequestsError.value = "";
    await apiRequest(
      `/groups/${groupId.value}` + `/join-requests/${requestId}/accept`,
      {
        method: "POST",
      },
    );
    joinRequests.value = joinRequests.value.filter(
      (request) => request.id !== requestId,
    );
    if (group.value) {
      group.value.member_count += 1;
    }
  } catch (err) {
    joinRequestsError.value = err.message;
  }
}

async function declineJoinRequest(requestId) {
  try {
    joinRequestsError.value = "";
    await apiRequest(
      `/groups/${groupId.value}` + `/join-requests/${requestId}/decline`,
      {
        method: "POST",
      },
    );
    joinRequests.value = joinRequests.value.filter(
      (request) => request.id !== requestId,
    );
  } catch (err) {
    joinRequestsError.value = err.message;
  }
}

async function loadGroupInvitations() {
  try {
    loadingInvitations.value = true;
    invitationsError.value = "";
    groupInvitations.value = await apiRequest(
      `/groups/${groupId.value}/invitations`,
    );
  } catch (err) {
    invitationsError.value = err.message;
  } finally {
    loadingInvitations.value = false;
  }
}

watch(
  () => route.fullPath,
  () => {
    activeGroupTab.value = "overview";
    loadedGroupTabs.value = {
      management: false,
    };
    group.value = null;
    joinRequests.value = [];
    groupInvitations.value = [];
    loadGroup();
  },
);
onMounted(() => {
  loadGroup();
});
</script>

<style scoped>
.group-detail-page {
  width: min(1000px, 100%);
}

.group-page-state {
  min-height: 220px;

  display: flex;
  align-items: center;
  justify-content: center;

  gap: 9px;

  color: var(--text-muted);
}

.group-page-error {
  margin: 0;

  padding: 11px 13px;

  border: 1px solid rgba(255, 95, 109, 0.2);

  border-radius: var(--radius-md);

  background: var(--danger-soft);

  color: var(--danger);

  font-size: 13px;
}

.group-hero {
  display: flex;

  gap: 20px;

  margin-bottom: 20px;
  padding: 24px;

  border: 1px solid var(--border-soft);

  border-radius: var(--radius-xl);

  background: var(--surface);

  box-shadow: var(--shadow-sm);
}

.group-hero-icon {
  width: 64px;
  height: 64px;

  flex: 0 0 64px;

  display: grid;
  place-items: center;

  border: 1px solid var(--primary-border);

  border-radius: 16px;

  background: var(--primary-soft);

  color: var(--primary);

  font-size: 22px;
  font-weight: 800;
}

.group-hero-content {
  min-width: 0;
  flex: 1;
}

.group-hero-title-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;

  gap: 16px;
}

.group-hero h1 {
  margin-bottom: 8px;
}

.group-description {
  max-width: 720px;

  color: var(--text-secondary);
}

.group-summary {
  display: flex;
  align-items: center;

  flex-wrap: wrap;

  gap: 8px;

  color: var(--text-muted);

  font-size: 12px;
}

.group-summary-divider {
  color: var(--border);
}

.group-membership-badge {
  flex-shrink: 0;

  padding: 4px 9px;

  border: 1px solid var(--border);

  border-radius: var(--radius-round);

  color: var(--text-secondary);

  font-size: 11px;
  font-weight: 700;
}

.status-owner {
  border-color: var(--primary-border);

  background: var(--primary-soft);

  color: var(--primary);
}

.status-member {
  border-color: rgba(54, 201, 143, 0.25);

  background: rgba(54, 201, 143, 0.08);

  color: var(--success);
}

.status-pending {
  color: var(--warning);
}

.group-join-actions {
  margin-top: 16px;
}

.group-member-actions {
  margin-top: 16px;
}

.group-leave-button {
  border-color: rgba(255, 95, 109, 0.28);

  color: var(--danger);
}

.group-leave-button:hover:not(:disabled) {
  border-color: rgba(255, 95, 109, 0.42);

  background: var(--danger-soft);

  color: var(--danger);
}

.group-invited-notice {
  padding: 11px 13px;

  border: 1px solid var(--primary-border);

  border-radius: var(--radius-md);

  background: var(--primary-soft);

  color: var(--text-secondary);

  font-size: 13px;
}

.group-tabs {
  position: sticky;
  top: 64px;

  z-index: 50;

  display: flex;

  gap: 4px;

  margin-bottom: 20px;
  padding: 6px;

  overflow-x: auto;

  border: 1px solid var(--border-soft);

  border-radius: var(--radius-lg);

  background: rgba(21, 24, 29, 0.94);

  backdrop-filter: blur(12px);

  scrollbar-width: none;
}

.group-tabs::-webkit-scrollbar {
  display: none;
}

.group-tab {
  position: relative;

  flex: 0 0 auto;

  min-height: 38px;

  padding: 7px 13px;

  border: 0;
  border-radius: var(--radius-md);

  background: transparent;

  color: var(--text-muted);

  font-size: 13px;
  font-weight: 650;
}

.group-tab:hover {
  background: var(--surface-2);

  color: var(--text);
}

.group-tab.active {
  background: var(--primary-soft);

  color: var(--primary);
}

.group-tab-badge {
  min-width: 18px;
  height: 18px;

  display: inline-flex;
  align-items: center;
  justify-content: center;

  margin-left: 5px;
  padding: 0 5px;

  border-radius: var(--radius-round);

  background: var(--primary);

  color: white;

  font-size: 10px;
}

.group-tab-panel {
  padding: 22px;

  border: 1px solid var(--border-soft);

  border-radius: var(--radius-lg);

  background: var(--surface);

  box-shadow: var(--shadow-sm);
}

.group-panel-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;

  gap: 18px;

  margin-bottom: 20px;
}

.group-panel-heading h2 {
  margin-bottom: 4px;
}

.group-overview-grid {
  display: grid;

  grid-template-columns: repeat(2, minmax(0, 1fr));

  gap: 12px;
}

.group-overview-item {
  display: flex;
  flex-direction: column;

  gap: 4px;

  padding: 15px;

  border: 1px solid var(--border-soft);

  border-radius: var(--radius-md);

  background: var(--bg-secondary);
}

.group-overview-item span {
  color: var(--text-muted);

  font-size: 11px;
}

.group-section-state,
.group-empty-state {
  color: var(--text-muted);

  font-size: 13px;

  text-align: center;
}

.management-person {
  min-width: 0;

  display: flex;
  align-items: center;
  gap: 10px;
}

.management-avatar {
  width: 40px;
  height: 40px;

  flex: 0 0 40px;

  display: grid;
  place-items: center;

  overflow: hidden;

  border: 1px solid var(--primary-border);
  border-radius: 50%;

  background: var(--primary-soft);
  color: var(--primary);

  font-size: 11px;
  font-weight: 800;
}

.management-person-info {
  min-width: 0;

  display: flex;
  flex-direction: column;
}

.management-person-info span {
  color: var(--text-muted);

  font-size: 11px;
}

.management-section + .management-section {
  margin-top: 28px;
  padding-top: 22px;

  border-top: 1px solid var(--border-soft);
}

.management-heading {
  display: flex;
  /* align-items: center; */

  gap: 8px;
}

.management-request,
.management-invitation {
  display: flex;
  align-items: center;
  justify-content: space-between;

  gap: 16px;

  padding: 14px;

  background: var(--bg-secondary);
}

.management-actions {
  flex-shrink: 0;

  display: flex;

  gap: 7px;
}

.invitation-status {
  color: var(--text-muted);

  font-size: 12px;
}

@media (max-width: 700px) {
  .group-hero {
    padding: 18px;

    flex-direction: column;
  }

  .group-hero-title-row {
    flex-direction: column;
  }

  .group-tabs {
    top: 55px;
  }

  .group-tab-panel {
    padding: 16px;
  }

  .group-panel-heading {
    flex-direction: column;
  }

  .group-overview-grid {
    grid-template-columns: 1fr;
  }

  .management-request,
  .management-invitation {
    align-items: stretch;

    flex-direction: column;
  }

  .management-actions {
    width: 100%;
  }

  .management-actions button {
    flex: 1;
  }
}
</style>
