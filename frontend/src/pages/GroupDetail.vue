<template>
  <main>
    <p v-if="loadingGroup">Loading group...</p>
    <p v-if="groupError">{{ groupError }}</p>

    <section v-if="group">
      <h1>{{ group.title }}</h1>

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

      <button v-if="group.membership_status === 'none'" @click="requestJoinGroup">
        Request to Join
      </button>

      <button v-else-if="group.membership_status === 'pending'" disabled>
        Join Request Pending
      </button>

      <button v-else-if="group.membership_status === 'invited'" disabled>
        You have an invitation. Check the Groups page.
      </button>

      <section v-if="isOwner">
        <hr />

        <h2>Pending Join Requests</h2>

        <p v-if="loadingJoinRequests">Loading join requests...</p>
        <p v-if="joinRequestsError">{{ joinRequestsError }}</p>

        <p v-if="joinRequests.length === 0">
          No pending join requests.
        </p>

        <article v-for="request in joinRequests" :key="request.id">
          <p>
            <strong>{{ request.requester_name }}</strong>
            wants to join this group.
          </p>

          <p v-if="request.requester_nickname">
            Nickname: {{ request.requester_nickname }}
          </p>

          <button @click="acceptJoinRequest(request.id)">
            Accept
          </button>

          <button @click="declineJoinRequest(request.id)">
            Decline
          </button>

          <hr />
        </article>

        <h2>Invite User</h2>

        <p v-if="usersError">{{ usersError }}</p>
        <p v-if="inviteError">{{ inviteError }}</p>
        <p v-if="inviteMessage">{{ inviteMessage }}</p>

        <form @submit.prevent="sendInvitation">
          <select v-model.number="selectedInviteeID" required>
            <option value="" disabled>
              Select user
            </option>

            <option v-for="user in users" :key="user.id" :value="user.id">
              {{ user.first_name }} {{ user.last_name }}
              <span v-if="user.nickname">
                - {{ user.nickname }}
              </span>
            </option>
          </select>

          <button type="submit">
            Send Invitation
          </button>
        </form>

        <h2>Group Invitations Sent</h2>

        <p v-if="loadingInvitations">Loading invitations...</p>
        <p v-if="invitationsError">{{ invitationsError }}</p>

        <p v-if="groupInvitations.length === 0">
          No invitations found.
        </p>

        <article v-for="invitation in groupInvitations" :key="invitation.id">
          <p>
            Invited:
            <strong>{{ invitation.invitee_name }}</strong>
          </p>

          <p>
            Status:
            <strong>{{ invitation.status }}</strong>
          </p>

          <hr />
        </article>
      </section>

      <section v-if="isMemberOrOwner">
        <hr />

        <h2>Group Events</h2>

        <form @submit.prevent="createGroupEvent">
          <div>
            <label for="event-title">Event Title</label>
            <input id="event-title" v-model="newGroupEvent.title" type="text" required />
          </div>

          <div>
            <label for="event-description">Event Description</label>
            <textarea id="event-description" v-model="newGroupEvent.description"></textarea>
          </div>

          <div>
            <label for="event-time">Event Time</label>
            <input id="event-time" v-model="newGroupEvent.event_time" type="datetime-local" required />
          </div>

          <button type="submit">Create Event</button>
        </form>

        <p v-if="loadingGroupEvents">Loading group events...</p>
        <p v-if="groupEventsError">{{ groupEventsError }}</p>

        <p v-if="groupEvents.length === 0">
          No group events yet.
        </p>

        <article v-for="event in groupEvents" :key="event.id">
          <h3>{{ event.title }}</h3>

          <p v-if="event.description">
            {{ event.description }}
          </p>

          <p>
            Time:
            <strong>{{ event.event_time }}</strong>
          </p>

          <p>
            Created by:
            <strong>{{ event.creator_name }}</strong>
          </p>

          <p>
            Going:
            <strong>{{ event.going_count }}</strong>
            |
            Not going:
            <strong>{{ event.not_going_count }}</strong>
          </p>

          <p>
            My response:
            <strong>{{ event.my_response }}</strong>
          </p>

          <button :disabled="event.my_response === 'going'" @click="respondToEvent(event.id, 'going')">
            Going
          </button>

          <button :disabled="event.my_response === 'not_going'" @click="respondToEvent(event.id, 'not_going')">
            Not Going
          </button>

          <hr />
        </article>
        <hr />

        <h2>Group Posts</h2>

        <form @submit.prevent="createGroupPost">
          <div>
            <label for="group-post-content">New Group Post</label>
            <textarea id="group-post-content" v-model="newGroupPostContent" required></textarea>
          </div>

          <div>
            <label for="group-post-image">Image/GIF</label>
            <input id="group-post-image" ref="groupPostImageInput" type="file" accept="image/png,image/jpeg,image/gif"
              @change="handleGroupPostImageChange" />
          </div>

          <button type="submit">Post to Group</button>
        </form>

        <p v-if="loadingGroupPosts">Loading group posts...</p>
        <p v-if="groupPostsError">{{ groupPostsError }}</p>

        <p v-if="groupPosts.length === 0">
          No group posts yet.
        </p>

        <article v-for="post in groupPosts" :key="post.id">
          <h3>{{ post.author_name }}</h3>

          <p v-if="post.author_nickname">
            Nickname: {{ post.author_nickname }}
          </p>

          <p>{{ post.content }}</p>

          <img v-if="post.image_path" :src="imageUrl(post.image_path)" alt="Group post image"
            style="max-width: 300px" />

          <p>{{ post.created_at }}</p>

          <section>
            <h4>Comments</h4>

            <p v-if="groupCommentErrors[post.id]">
              {{ groupCommentErrors[post.id] }}
            </p>

            <article v-for="comment in groupCommentsByPost[post.id] || []" :key="comment.id">
              <p>
                <strong>{{ comment.author_name }}</strong>
                <span v-if="comment.author_nickname">
                  - {{ comment.author_nickname }}
                </span>
              </p>

              <p>{{ comment.content }}</p>

              <img v-if="comment.image_path" :src="imageUrl(comment.image_path)" alt="Group comment image"
                style="max-width: 200px" />

              <p>{{ comment.created_at }}</p>
            </article>

            <form @submit.prevent="createGroupComment(post.id)">
              <div>
                <label :for="`group-comment-${post.id}`">
                  Add Comment
                </label>

                <input :id="`group-comment-${post.id}`" v-model="newGroupComments[post.id]" type="text" required />
              </div>

              <div>
                <label :for="`group-comment-image-${post.id}`">
                  Image/GIF
                </label>

                <input :id="`group-comment-image-${post.id}`" :ref="(el) => {
                  if (el) groupCommentImageInputs[post.id] = el;
                }" type="file" accept="image/png,image/jpeg,image/gif"
                  @change="handleGroupCommentImageChange(post.id, $event)" />
              </div>

              <button type="submit">Comment</button>
            </form>
          </section>

          <hr />
        </article>
      </section>
    </section>
  </main>
</template>

<script setup>
import { computed, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiRequest } from "../services/api";

const route = useRoute();

const group = ref(null);
const joinRequests = ref([]);
const groupInvitations = ref([]);
const users = ref([]);

const selectedInviteeID = ref("");

const loadingGroup = ref(false);
const loadingJoinRequests = ref(false);
const loadingInvitations = ref(false);

const groupError = ref("");
const joinRequestsError = ref("");
const invitationsError = ref("");
const usersError = ref("");
const inviteError = ref("");
const inviteMessage = ref("");

// Group Feed
const groupPosts = ref([]);
const groupCommentsByPost = ref({});
const newGroupPostContent = ref("");
const newGroupPostImage = ref(null);
const groupPostImageInput = ref(null);

const newGroupComments = ref({});
const newGroupCommentImages = ref({});
const groupCommentImageInputs = ref({});

const loadingGroupPosts = ref(false);
const groupPostsError = ref("");
const groupCommentErrors = ref({});

// Group Events
const groupEvents = ref([]);

const newGroupEvent = ref({
  title: "",
  description: "",
  event_time: "",
});

const loadingGroupEvents = ref(false);
const groupEventsError = ref("");

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

async function loadGroup() {
  try {
    loadingGroup.value = true;
    groupError.value = "";

    group.value = await apiRequest(`/groups/${groupId.value}`);

    if (isOwner.value) {
      await loadJoinRequests();
      await loadGroupInvitations();
      await loadUsers();
    }

    if (isMemberOrOwner.value) {
      await loadGroupPosts();
      await loadGroupEvents();
    }
  } catch (err) {
    groupError.value = err.message;
  } finally {
    loadingGroup.value = false;
  }
}

async function requestJoinGroup() {
  try {
    groupError.value = "";

    await apiRequest(`/groups/${groupId.value}/join-request`, {
      method: "POST",
    });

    await loadGroup();
  } catch (err) {
    groupError.value = err.message;
  }
}

async function loadJoinRequests() {
  try {
    loadingJoinRequests.value = true;
    joinRequestsError.value = "";

    joinRequests.value = await apiRequest(
      `/groups/${groupId.value}/join-requests`
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
      `/groups/${groupId.value}/join-requests/${requestId}/accept`,
      {
        method: "POST",
      }
    );

    await loadGroup();
  } catch (err) {
    joinRequestsError.value = err.message;
  }
}

async function declineJoinRequest(requestId) {
  try {
    joinRequestsError.value = "";

    await apiRequest(
      `/groups/${groupId.value}/join-requests/${requestId}/decline`,
      {
        method: "POST",
      }
    );

    await loadGroup();
  } catch (err) {
    joinRequestsError.value = err.message;
  }
}

async function loadGroupInvitations() {
  try {
    loadingInvitations.value = true;
    invitationsError.value = "";

    groupInvitations.value = await apiRequest(
      `/groups/${groupId.value}/invitations`
    );
  } catch (err) {
    invitationsError.value = err.message;
  } finally {
    loadingInvitations.value = false;
  }
}

async function loadUsers() {
  try {
    usersError.value = "";

    users.value = await apiRequest("/users");
  } catch (err) {
    usersError.value = err.message;
  }
}

async function sendInvitation() {
  try {
    inviteError.value = "";
    inviteMessage.value = "";

    await apiRequest(`/groups/${groupId.value}/invitations`, {
      method: "POST",
      body: JSON.stringify({
        invitee_id: selectedInviteeID.value,
      }),
    });

    selectedInviteeID.value = "";
    inviteMessage.value = "Invitation sent successfully.";

    await loadGroupInvitations();
  } catch (err) {
    inviteError.value = err.message;
  }
}

function imageUrl(path) {
  return `http://localhost:8080${path}`;
}

function handleGroupPostImageChange(event) {
  const file = event.target.files[0];
  newGroupPostImage.value = file || null;
}

function handleGroupCommentImageChange(postId, event) {
  const file = event.target.files[0];
  newGroupCommentImages.value[postId] = file || null;
}

async function loadGroupPosts() {
  if (!isMemberOrOwner.value) {
    groupPosts.value = [];
    groupCommentsByPost.value = {};
    return;
  }

  try {
    loadingGroupPosts.value = true;
    groupPostsError.value = "";

    groupPosts.value = await apiRequest(`/groups/${groupId.value}/posts`);

    for (const post of groupPosts.value) {
      await loadGroupComments(post.id);
    }
  } catch (err) {
    groupPostsError.value = err.message;
  } finally {
    loadingGroupPosts.value = false;
  }
}

async function loadGroupComments(postId) {
  try {
    groupCommentErrors.value[postId] = "";

    groupCommentsByPost.value[postId] = await apiRequest(
      `/groups/${groupId.value}/posts/${postId}/comments`
    );
  } catch (err) {
    groupCommentErrors.value[postId] = err.message;
  }
}

async function createGroupPost() {
  try {
    groupPostsError.value = "";

    const formData = new FormData();

    formData.append("content", newGroupPostContent.value);

    if (newGroupPostImage.value) {
      formData.append("image", newGroupPostImage.value);
    }

    await apiRequest(`/groups/${groupId.value}/posts`, {
      method: "POST",
      body: formData,
    });

    newGroupPostContent.value = "";
    newGroupPostImage.value = null;

    if (groupPostImageInput.value) {
      groupPostImageInput.value.value = "";
    }

    await loadGroupPosts();
  } catch (err) {
    groupPostsError.value = err.message;
  }
}

async function createGroupComment(postId) {
  try {
    groupCommentErrors.value[postId] = "";

    const content = newGroupComments.value[postId] || "";

    const formData = new FormData();

    formData.append("content", content);

    if (newGroupCommentImages.value[postId]) {
      formData.append("image", newGroupCommentImages.value[postId]);
    }

    await apiRequest(`/groups/${groupId.value}/posts/${postId}/comments`, {
      method: "POST",
      body: formData,
    });

    newGroupComments.value[postId] = "";
    newGroupCommentImages.value[postId] = null;

    if (groupCommentImageInputs.value[postId]) {
      groupCommentImageInputs.value[postId].value = "";
    }

    await loadGroupComments(postId);
  } catch (err) {
    groupCommentErrors.value[postId] = err.message;
  }
}

function formatEventTimeForBackend(value) {
  if (!value) return "";

  return value.replace("T", " ") + ":00";
}

async function loadGroupEvents() {
  if (!isMemberOrOwner.value) {
    groupEvents.value = [];
    return;
  }

  try {
    loadingGroupEvents.value = true;
    groupEventsError.value = "";

    groupEvents.value = await apiRequest(`/groups/${groupId.value}/events`);
  } catch (err) {
    groupEventsError.value = err.message;
  } finally {
    loadingGroupEvents.value = false;
  }
}

async function createGroupEvent() {
  try {
    groupEventsError.value = "";

    await apiRequest(`/groups/${groupId.value}/events`, {
      method: "POST",
      body: JSON.stringify({
        title: newGroupEvent.value.title,
        description: newGroupEvent.value.description,
        event_time: formatEventTimeForBackend(newGroupEvent.value.event_time),
      }),
    });

    newGroupEvent.value.title = "";
    newGroupEvent.value.description = "";
    newGroupEvent.value.event_time = "";

    await loadGroupEvents();
  } catch (err) {
    groupEventsError.value = err.message;
  }
}

async function respondToEvent(eventId, response) {
  try {
    groupEventsError.value = "";

    const action = response === "going" ? "going" : "not-going";

    await apiRequest(`/groups/${groupId.value}/events/${eventId}/${action}`, {
      method: "POST",
    });

    await loadGroupEvents();
  } catch (err) {
    groupEventsError.value = err.message;
  }
}

onMounted(() => {
  loadGroup();
});

watch(
  () => route.fullPath,
  () => {
    group.value = null;
    joinRequests.value = [];
    groupInvitations.value = [];
    users.value = [];
    selectedInviteeID.value = "";

    groupPosts.value = [];
    groupCommentsByPost.value = {};
    newGroupPostContent.value = "";
    newGroupPostImage.value = null;
    newGroupComments.value = {};
    newGroupCommentImages.value = {};
    groupPostsError.value = "";
    groupCommentErrors.value = {};

    groupEvents.value = [];
    newGroupEvent.value = {
      title: "",
      description: "",
      event_time: "",
    };
    groupEventsError.value = "";

    loadGroup();
  }
);

</script>