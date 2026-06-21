<template>
  <main>
    <h1>Groups</h1>

    <section>
      <h2>Create Group</h2>

      <form @submit.prevent="createGroup">
        <div>
          <label for="title">Title</label>
          <input id="title" v-model="newGroup.title" type="text" required />
        </div>

        <div>
          <label for="description">Description</label>
          <textarea id="description" v-model="newGroup.description"></textarea>
        </div>

        <button type="submit">Create Group</button>
      </form>

      <p v-if="createError">{{ createError }}</p>
    </section>

    <hr />

    <section>
      <h2>My Group Invitations</h2>

      <p v-if="loadingInvitations">Loading invitations...</p>
      <p v-if="invitationsError">{{ invitationsError }}</p>

      <p v-if="myInvitations.length === 0">
        No pending group invitations.
      </p>

      <article v-for="invitation in myInvitations" :key="invitation.id">
        <p>
          <strong>{{ invitation.inviter_name }}</strong>
          invited you to join
          <strong>{{ invitation.group_title }}</strong>.
        </p>

        <button @click="acceptInvitation(invitation.id)">
          Accept
        </button>

        <button @click="declineInvitation(invitation.id)">
          Decline
        </button>

        <hr />
      </article>
    </section>

    <hr />

    <section>
      <h2>All Groups</h2>

      <p v-if="loadingGroups">Loading groups...</p>
      <p v-if="groupsError">{{ groupsError }}</p>

      <p v-if="groups.length === 0">
        No groups yet.
      </p>

      <article v-for="group in groups" :key="group.id">
        <h3>
          <router-link :to="`/groups/${group.id}`">
            {{ group.title }}
          </router-link>
        </h3>

        <p>{{ group.description }}</p>

        <p>
          Created by:
          <strong>{{ group.creator_name }}</strong>
        </p>

        <p>
          Members:
          <strong>{{ group.member_count }}</strong>
        </p>

        <p>
          Your status:
          <strong>{{ group.membership_status }}</strong>
        </p>

        <button v-if="group.membership_status === 'none'" @click="requestJoinGroup(group.id)">
          Request to Join
        </button>

        <button v-else-if="group.membership_status === 'pending'" disabled>
          Join Request Pending
        </button>

        <button v-else-if="group.membership_status === 'invited'" disabled>
          You Are Invited
        </button>

        <router-link v-else :to="`/groups/${group.id}`">
          Open Group
        </router-link>

        <hr />
      </article>
    </section>
  </main>
</template>

<script setup>
import { onMounted, ref } from "vue";
import { apiRequest } from "../services/api";

const groups = ref([]);
const myInvitations = ref([]);

const loadingGroups = ref(false);
const loadingInvitations = ref(false);

const groupsError = ref("");
const createError = ref("");
const invitationsError = ref("");

const newGroup = ref({
  title: "",
  description: "",
});

async function loadGroups() {
  try {
    loadingGroups.value = true;
    groupsError.value = "";

    groups.value = await apiRequest("/groups");
  } catch (err) {
    groupsError.value = err.message;
  } finally {
    loadingGroups.value = false;
  }
}

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

async function createGroup() {
  try {
    createError.value = "";

    await apiRequest("/groups", {
      method: "POST",
      body: JSON.stringify({
        title: newGroup.value.title,
        description: newGroup.value.description,
      }),
    });

    newGroup.value.title = "";
    newGroup.value.description = "";

    await loadGroups();
  } catch (err) {
    createError.value = err.message;
  }
}

async function requestJoinGroup(groupId) {
  try {
    groupsError.value = "";

    await apiRequest(`/groups/${groupId}/join-request`, {
      method: "POST",
    });

    await loadGroups();
  } catch (err) {
    groupsError.value = err.message;
  }
}

async function acceptInvitation(invitationId) {
  try {
    invitationsError.value = "";

    await apiRequest(`/group-invitations/${invitationId}/accept`, {
      method: "POST",
    });

    await loadMyInvitations();
    await loadGroups();
  } catch (err) {
    invitationsError.value = err.message;
  }
}

async function declineInvitation(invitationId) {
  try {
    invitationsError.value = "";

    await apiRequest(`/group-invitations/${invitationId}/decline`, {
      method: "POST",
    });

    await loadMyInvitations();
    await loadGroups();
  } catch (err) {
    invitationsError.value = err.message;
  }
}

onMounted(async () => {
  await loadGroups();
  await loadMyInvitations();
});
</script>