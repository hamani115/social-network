<template>
  <section v-show="active" class="group-tab-panel group-members-panel">
    

    <div class="group-panel-heading">
      <h2>Members</h2>

      <span class="group-members-count">
        {{ memberCount }}
        {{ memberCount === 1 ? "member" : "members" }}
      </span>
    </div>

    <!-- Search -->
    <div class="group-member-search">
      <span class="group-member-search-icon" aria-hidden="true">
        <i class="fa-solid fa-magnifying-glass"></i>
      </span>

      <input
        v-model="memberSearchQuery"
        type="search"
        placeholder="Search group members..."
        autocomplete="off"
      />

      <button
        v-if="memberSearchQuery"
        type="button"
        class="group-member-search-clear"
        aria-label="Clear member search"
        @click="clearMemberSearch"
      >
        <i class="fa-solid fa-xmark" aria-hidden="true"></i>
      </button>
    </div>

    <!-- Loading -->

    <div v-if="loadingGroupMembers" class="group-section-state">
      <span class="loading-spinner"></span>

      Loading members...
    </div>

    <p v-else-if="groupMembersError" class="group-page-error">
      {{ groupMembersError }}
    </p>

    <!-- No results -->
    <div
      v-else-if="groupMembers.length === 0 && debouncedMemberSearchQuery"
      class="group-members-empty"
    >
      <h3>No members found</h3>

      <p>No group members match "{{ debouncedMemberSearchQuery }}"</p>
    </div>

    <!-- Members -->
    <div v-else class="group-member-list">
      <article
        v-for="member in groupMembers"
        :key="member.id"
        class="group-member-card"
      >
        <router-link :to="`/profiles/${member.id}`" class="group-member-main">
          <UserAvatar
            :avatar-path="member.avatar_path"
            :name="`${member.first_name} ${member.last_name}`"
            class="group-member-avatar"
          />

          <div class="group-member-identity">
            <strong>
              {{ member.first_name }}
              {{ member.last_name }}
            </strong>

            <span v-if="member.nickname">
              {{ member.nickname }}
            </span>
          </div>
        </router-link>

        <span
          class="group-member-role"
          :class="{
            owner: member.role === 'owner',
          }"
        >
          {{ member.role === "owner" ? "Owner" : "Member" }}
        </span>
      </article>

      <!-- Pagination -->
      <div class="group-member-pagination">
        <div v-if="loadingMoreGroupMembers" class="group-member-loading-more">
          <span class="loading-spinner"></span>

          Loading more members...
        </div>

        <div
          v-else-if="groupMembersLoadMoreError"
          class="group-member-load-error"
        >
          <span>
            {{ groupMembersLoadMoreError }}
          </span>

          <button
            type="button"
            class="button button-ghost button-small"
            @click="loadGroupMembers()"
          >
            Try again
          </button>
        </div>

        <div
          v-else-if="hasMoreGroupMembers"
          ref="groupMembersLoadTrigger"
          class="group-member-load-trigger"
          aria-hidden="true"
        ></div>

        <p v-else class="group-member-end">End of members.</p>
      </div>
    </div>

    <!-- Invite -->
    <section class="group-invite-section">
      <div class="group-invite-heading">
        <h3>Invite someone</h3>
      </div>

      <div class="group-invite-search">
        <span class="group-member-search-icon" aria-hidden="true">
          <i class="fa-solid fa-magnifying-glass"></i>
        </span>

        <input
          v-model="inviteSearchQuery"
          type="search"
          placeholder="Search by name or nickname..."
          autocomplete="off"
        />
      </div>

      <p v-if="inviteError" class="group-page-error">
        {{ inviteError }}
      </p>

      <p v-if="inviteMessage" class="group-success">
        {{ inviteMessage }}
      </p>

      <p v-if="inviteCandidatesError" class="group-page-error">
        {{ inviteCandidatesError }}
      </p>

      <div v-else-if="loadingInviteCandidates" class="group-section-state">
        <span class="loading-spinner"></span>

        Searching...
      </div>

      <!-- No results -->

      <p
        v-else-if="debouncedInviteSearchQuery && inviteCandidates.length === 0"
        class="group-invite-hint"
      >
        No available users found.
      </p>

      <!-- LIST -->
      <div v-else class="group-invite-candidate-list">
        <article
          v-for="user in inviteCandidates"
          :key="user.id"
          class="group-invite-candidate"
        >
          <router-link
            :to="`/profiles/${user.id}`"
            class="group-invite-candidate-main"
          >
            <UserAvatar
              :avatar-path="user.avatar_path"
              :name="`${user.first_name} ${user.last_name}`"
              class="group-member-avatar"
            />

            <div class="group-member-identity">
              <strong>
                {{ user.first_name }}
                {{ user.last_name }}
              </strong>

              <span v-if="user.nickname">
                {{ user.nickname }}
              </span>
            </div>
          </router-link>

          <button
            type="button"
            class="button button-primary button-small"
            :disabled="invitingUserID === user.id"
            @click="sendInvitation(user)"
          >
            {{ invitingUserID === user.id ? "Sending..." : "Invite" }}
          </button>
        </article>

        <button
          v-if="hasMoreInviteCandidates"
          type="button"
          class="button button-ghost group-invite-more"
          :disabled="loadingMoreInviteCandidates"
          @click="loadInviteCandidates()"
        >
          {{ loadingMoreInviteCandidates ? "Loading..." : "Show more people" }}
        </button>
      </div>
    </section>
  </section>
</template>

<script setup>
import UserAvatar from "../UserAvatar.vue";

import { onUnmounted, ref, watch } from "vue";

import { apiRequest } from "../../services/api";

const props = defineProps({
  groupId: {
    type: [String, Number],
    required: true,
  },

  memberCount: {
    type: Number,
    required: true,
  },

  active: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(["invitation-sent"]);

const groupMembers = ref([]);

const loadingGroupMembers = ref(false);
const loadingMoreGroupMembers = ref(false);

const groupMembersError = ref("");
const groupMembersLoadMoreError = ref("");

const GROUP_MEMBERS_PAGE_SIZE = 20;

const groupMemberOffset = ref(0);
const hasMoreGroupMembers = ref(true);

const groupMembersLoadTrigger = ref(null);

let groupMembersObserver = null;
let groupMembersRequestVersion = 0;

const membersLoaded = ref(false);

// search

const memberSearchQuery = ref("");
const debouncedMemberSearchQuery = ref("");

let memberSearchTimer = null;

const inviteCandidates = ref([]);

const inviteSearchQuery = ref("");
const debouncedInviteSearchQuery = ref("");

const loadingInviteCandidates = ref(false);
const loadingMoreInviteCandidates = ref(false);

const inviteCandidatesError = ref("");

const INVITE_CANDIDATE_PAGE_SIZE = 10;

const inviteCandidateOffset = ref(0);
const hasMoreInviteCandidates = ref(false);

const invitingUserID = ref(null);

const inviteError = ref("");
const inviteMessage = ref("");

let inviteSearchTimer = null;
let inviteRequestVersion = 0;

async function loadGroupMembers(reset = false) {
  if (
    !reset &&
    (loadingGroupMembers.value ||
      loadingMoreGroupMembers.value ||
      !hasMoreGroupMembers.value)
  ) {
    return;
  }

  if (reset) {
    groupMembersRequestVersion += 1;

    groupMembers.value = [];

    groupMemberOffset.value = 0;

    hasMoreGroupMembers.value = true;

    groupMembersLoadMoreError.value = "";
  }

  const requestVersion = groupMembersRequestVersion;

  const initialLoad = groupMemberOffset.value === 0;

  try {
    if (initialLoad) {
      loadingGroupMembers.value = true;

      groupMembersError.value = "";
    } else {
      loadingMoreGroupMembers.value = true;

      groupMembersLoadMoreError.value = "";
    }

    const params = new URLSearchParams();

    params.set("limit", String(GROUP_MEMBERS_PAGE_SIZE));

    params.set("offset", String(groupMemberOffset.value));

    if (debouncedMemberSearchQuery.value) {
      params.set("q", debouncedMemberSearchQuery.value);
    }

    const result = await apiRequest(
      `/groups/${props.groupId}/members` + `?${params.toString()}`,
    );

    if (requestVersion !== groupMembersRequestVersion) {
      return;
    }

    const incomingMembers = result.members || [];

    if (reset) {
      groupMembers.value = incomingMembers;
    } else {
      const existingIDs = new Set(
        groupMembers.value.map((member) => member.id),
      );

      groupMembers.value.push(
        ...incomingMembers.filter((member) => !existingIDs.has(member.id)),
      );
    }

    groupMemberOffset.value =
      result.next_offset ?? groupMemberOffset.value + incomingMembers.length;

    hasMoreGroupMembers.value = Boolean(result.has_more);
  } catch (err) {
    if (requestVersion !== groupMembersRequestVersion) {
      return;
    }

    if (initialLoad) {
      groupMembersError.value = err.message;
    } else {
      groupMembersLoadMoreError.value = err.message;
    }
  } finally {
    if (requestVersion === groupMembersRequestVersion) {
      loadingGroupMembers.value = false;

      loadingMoreGroupMembers.value = false;
    }
  }
}

async function clearMemberSearch() {
  if (memberSearchTimer) {
    clearTimeout(memberSearchTimer);

    memberSearchTimer = null;
  }

  memberSearchQuery.value = "";

  debouncedMemberSearchQuery.value = "";

  await loadGroupMembers(true);
}

// infinite scroll

function observeGroupMembersTrigger(element) {
  if (groupMembersObserver) {
    groupMembersObserver.disconnect();

    groupMembersObserver = null;
  }

  if (!element) {
    return;
  }

  groupMembersObserver = new IntersectionObserver(
    (entries) => {
      const entry = entries[0];

      if (
        entry.isIntersecting &&
        props.active &&
        hasMoreGroupMembers.value &&
        !loadingGroupMembers.value &&
        !loadingMoreGroupMembers.value
      ) {
        loadGroupMembers();
      }
    },
    {
      root: null,
      rootMargin: "250px 0px",
      threshold: 0,
    },
  );

  groupMembersObserver.observe(element);
}

async function loadInviteCandidates(reset = false) {
  const query = debouncedInviteSearchQuery.value.trim();

  if (!query) {
    inviteCandidates.value = [];

    hasMoreInviteCandidates.value = false;

    return;
  }

  if (
    !reset &&
    (loadingInviteCandidates.value ||
      loadingMoreInviteCandidates.value ||
      !hasMoreInviteCandidates.value)
  ) {
    return;
  }

  if (reset) {
    inviteRequestVersion += 1;

    inviteCandidates.value = [];

    inviteCandidateOffset.value = 0;

    hasMoreInviteCandidates.value = true;

    inviteCandidatesError.value = "";
  }

  const requestVersion = inviteRequestVersion;

  const initialLoad = inviteCandidateOffset.value === 0;

  try {
    if (initialLoad) {
      loadingInviteCandidates.value = true;
    } else {
      loadingMoreInviteCandidates.value = true;
    }

    inviteCandidatesError.value = "";

    const params = new URLSearchParams();

    params.set("q", query);

    params.set("limit", String(INVITE_CANDIDATE_PAGE_SIZE));

    params.set("offset", String(inviteCandidateOffset.value));

    const result = await apiRequest(
      `/groups/${props.groupId}` +
        `/invite-candidates` +
        `?${params.toString()}`,
    );

    if (requestVersion !== inviteRequestVersion) {
      return;
    }

    const incomingUsers = result.users || [];

    if (reset) {
      inviteCandidates.value = incomingUsers;
    } else {
      const existingIDs = new Set(
        inviteCandidates.value.map((user) => user.id),
      );

      inviteCandidates.value.push(
        ...incomingUsers.filter((user) => !existingIDs.has(user.id)),
      );
    }

    inviteCandidateOffset.value =
      result.next_offset ?? inviteCandidateOffset.value + incomingUsers.length;

    hasMoreInviteCandidates.value = Boolean(result.has_more);
  } catch (err) {
    if (requestVersion === inviteRequestVersion) {
      inviteCandidatesError.value = err.message;
    }
  } finally {
    if (requestVersion === inviteRequestVersion) {
      loadingInviteCandidates.value = false;

      loadingMoreInviteCandidates.value = false;
    }
  }
}

async function sendInvitation(user) {
  try {
    invitingUserID.value = user.id;

    inviteError.value = "";
    inviteMessage.value = "";

    await apiRequest(`/groups/${props.groupId}/invitations`, {
      method: "POST",

      body: JSON.stringify({
        invitee_id: user.id,
      }),
    });

    inviteMessage.value =
      `Invitation sent to ` + `${user.first_name} ` + `${user.last_name}.`;

    inviteCandidates.value = inviteCandidates.value.filter(
      (candidate) => candidate.id !== user.id,
    );

    emit("invitation-sent");
  } catch (err) {
    inviteError.value = err.message;
  } finally {
    invitingUserID.value = null;
  }
}

function resetMembers() {
  groupMembersRequestVersion += 1;
  inviteRequestVersion += 1;

  if (memberSearchTimer) {
    clearTimeout(memberSearchTimer);
    memberSearchTimer = null;
  }

  if (inviteSearchTimer) {
    clearTimeout(inviteSearchTimer);
    inviteSearchTimer = null;
  }

  membersLoaded.value = false;

  groupMembers.value = [];

  loadingGroupMembers.value = false;
  loadingMoreGroupMembers.value = false;

  groupMembersError.value = "";
  groupMembersLoadMoreError.value = "";

  groupMemberOffset.value = 0;
  hasMoreGroupMembers.value = true;

  memberSearchQuery.value = "";
  debouncedMemberSearchQuery.value = "";

  inviteCandidates.value = [];

  inviteSearchQuery.value = "";
  debouncedInviteSearchQuery.value = "";

  loadingInviteCandidates.value = false;

  loadingMoreInviteCandidates.value = false;

  inviteCandidatesError.value = "";

  inviteCandidateOffset.value = 0;

  hasMoreInviteCandidates.value = false;

  invitingUserID.value = null;

  inviteError.value = "";
  inviteMessage.value = "";
}

watch(
  () => props.active,
  async (active) => {
    if (active && !membersLoaded.value) {
      await loadGroupMembers(true);

      membersLoaded.value = true;
    }
  },
  {
    immediate: true,
  },
);

watch(
  () => props.groupId,
  async () => {
    resetMembers();

    if (props.active) {
      await loadGroupMembers(true);

      membersLoaded.value = true;
    }
  },
);

watch(
  () => props.memberCount,
  async (newMemberCount, oldMemberCount) => {
    if (newMemberCount === oldMemberCount || !membersLoaded.value) {
      return;
    }

    await loadGroupMembers(true);
  },
);

watch(memberSearchQuery, (value) => {
  if (memberSearchTimer) {
    clearTimeout(memberSearchTimer);
  }

  memberSearchTimer = setTimeout(async () => {
    debouncedMemberSearchQuery.value = value.trim();

    await loadGroupMembers(true);
  }, 300);
});

watch(inviteSearchQuery, (value) => {
  if (inviteSearchTimer) {
    clearTimeout(inviteSearchTimer);
  }

  const trimmed = value.trim();

  if (!trimmed) {
    inviteRequestVersion += 1;

    debouncedInviteSearchQuery.value = "";

    inviteCandidates.value = [];

    hasMoreInviteCandidates.value = false;

    return;
  }

  inviteSearchTimer = setTimeout(async () => {
    debouncedInviteSearchQuery.value = trimmed;

    await loadInviteCandidates(true);
  }, 300);
});

watch(groupMembersLoadTrigger, (element) => {
  observeGroupMembersTrigger(element);
});

onUnmounted(() => {
  if (groupMembersObserver) {
    groupMembersObserver.disconnect();
  }

  if (memberSearchTimer) {
    clearTimeout(memberSearchTimer);
  }

  if (inviteSearchTimer) {
    clearTimeout(inviteSearchTimer);
  }
});
</script>

<style scoped>
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

.group-section-state {
  color: var(--text-muted);

  font-size: 13px;

  text-align: center;
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

.group-success {
  color: var(--success);
}

.group-members-count {
  flex-shrink: 0;

  padding: 5px 10px;

  border-radius: var(--radius-round);

  background: var(--primary-soft);

  color: var(--primary);

  font-size: 11px;
  font-weight: 700;
}

.group-member-search,
.group-invite-search {
  position: relative;

  margin-bottom: 18px;
}

.group-member-search input,
.group-invite-search input {
  padding-left: 42px;
  padding-right: 40px;

  background: var(--bg-secondary);
}

.group-member-search-icon {
  position: absolute;

  left: 14px;
  top: 50%;

  color: var(--text-muted);

  pointer-events: none;

  transform: translateY(-50%);

  font-size: 14px;
}

.group-member-search-clear {
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

.group-member-list {
  display: grid;
}

.group-member-card,
.group-invite-candidate {
  display: flex;
  align-items: center;
  justify-content: space-between;

  gap: 14px;

  padding: 13px 2px;

  border-top: 1px solid var(--border-soft);
}

.group-member-main,
.group-invite-candidate-main {
  min-width: 0;

  display: flex;
  align-items: center;

  gap: 11px;

  color: inherit;
}

.group-member-avatar {
  width: 42px;
  height: 42px;

  flex: 0 0 42px;
  overflow: hidden;

  display: grid;
  place-items: center;

  border: 1px solid var(--primary-border);
  border-radius: 50%;

  background: var(--primary-soft);

  color: var(--primary);

  font-size: 12px;
  font-weight: 800;
}

.group-member-identity {
  min-width: 0;

  display: flex;
  flex-direction: column;
}

.group-member-identity span {
  color: var(--text-muted);

  font-size: 11px;
}

.group-member-role {
  flex-shrink: 0;

  padding: 3px 8px;

  border: 1px solid var(--border);
  border-radius: var(--radius-round);

  color: var(--text-muted);

  font-size: 10px;
  font-weight: 700;
}

.group-member-role.owner {
  border-color: var(--primary-border);

  background: var(--primary-soft);

  color: var(--primary);
}

.group-member-pagination {
  padding: 14px 0 2px;

  text-align: center;
}

.group-member-loading-more,
.group-member-load-error {
  min-height: 40px;

  display: flex;
  align-items: center;
  justify-content: center;

  gap: 8px;

  color: var(--text-muted);

  font-size: 12px;
}

.group-member-load-trigger {
  height: 2px;
}

.group-member-end {
  margin: 0;

  color: var(--text-muted);

  font-size: 11px;
}

.group-invite-section {
  margin-top: 26px;
  padding-top: 22px;

  border-top: 1px solid var(--border-soft);
}

.group-invite-hint {
  color: var(--text-muted);

  font-size: 12px;
}

.group-invite-candidate-list {
  display: grid;
}

.group-invite-more {
  width: 100%;

  margin-top: 8px;
}

@media (max-width: 700px) {
  .group-tab-panel {
    padding: 16px;
  }

  .group-panel-heading {
    flex-direction: column;
  }

  .group-member-card,
  .group-invite-candidate {
    align-items: stretch;

    flex-direction: column;
  }

  .group-invite-candidate button {
    width: 100%;
  }
}
</style>
