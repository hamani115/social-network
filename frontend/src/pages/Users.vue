<template>
    <main>
        <h1>Users</h1>

        <section>
            <h2>People</h2>

            <p v-if="loadingUsers">Loading users...</p>
            <p v-if="usersError">{{ usersError }}</p>

            <article v-for="user in users" :key="user.id">
                <h3>
                    <router-link :to="`/profiles/${user.id}`">
                        {{ user.first_name }} {{ user.last_name }}
                    </router-link>
                </h3>

                <p v-if="user.nickname">
                    Nickname: {{ user.nickname }}
                </p>

                <p>
                    Profile:
                    <strong>{{ user.is_public ? "Public" : "Private" }}</strong>
                </p>

                <p>
                    Follow status:
                    <strong>{{ user.follow_status }}</strong>
                </p>

                <button v-if="user.follow_status === 'none'" @click="followUser(user.id)">
                    Follow
                </button>

                <button v-else-if="user.follow_status === 'following'" @click="unfollowUser(user.id)">
                    Unfollow
                </button>

                <button v-else-if="user.follow_status === 'pending'" @click="unfollowUser(user.id)">
                    Cancel Request
                </button>

                <hr />
            </article>
        </section>

        <section>
            <h2>Follow Requests</h2>

            <p v-if="loadingRequests">Loading follow requests...</p>
            <p v-if="requestsError">{{ requestsError }}</p>

            <p v-if="followRequests.length === 0">
                No pending follow requests.
            </p>

            <article v-for="request in followRequests" :key="request.id">
                <p>
                    <strong>{{ request.requester_name }}</strong>
                    wants to follow you.
                </p>

                <p v-if="request.requester_nickname">
                    Nickname: {{ request.requester_nickname }}
                </p>

                <button @click="acceptRequest(request.id)">
                    Accept
                </button>

                <button @click="declineRequest(request.id)">
                    Decline
                </button>

                <hr />
            </article>
        </section>
    </main>
</template>

<script setup>
import { onMounted, ref } from "vue";
import { apiRequest } from "../services/api";

const users = ref([]);
const followRequests = ref([]);

const loadingUsers = ref(false);
const loadingRequests = ref(false);

const usersError = ref("");
const requestsError = ref("");

async function loadUsers() {
    try {
        loadingUsers.value = true;
        usersError.value = "";

        users.value = await apiRequest("/users");
    } catch (err) {
        usersError.value = err.message;
    } finally {
        loadingUsers.value = false;
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
        usersError.value = "";

        await apiRequest(`/users/${userId}/follow`, {
            method: "POST",
        });

        await loadUsers();
        await loadFollowRequests();
    } catch (err) {
        usersError.value = err.message;
    }
}

async function unfollowUser(userId) {
    try {
        usersError.value = "";

        await apiRequest(`/users/${userId}/unfollow`, {
            method: "POST",
        });

        await loadUsers();
        await loadFollowRequests();
    } catch (err) {
        usersError.value = err.message;
    }
}

async function acceptRequest(requestId) {
    try {
        requestsError.value = "";

        await apiRequest(`/follow-requests/${requestId}/accept`, {
            method: "POST",
        });

        await loadUsers();
        await loadFollowRequests();
    } catch (err) {
        requestsError.value = err.message;
    }
}

async function declineRequest(requestId) {
    try {
        requestsError.value = "";

        await apiRequest(`/follow-requests/${requestId}/decline`, {
            method: "POST",
        });

        await loadUsers();
        await loadFollowRequests();
    } catch (err) {
        requestsError.value = err.message;
    }
}

onMounted(async () => {
    await loadUsers();
    await loadFollowRequests();
});
</script>