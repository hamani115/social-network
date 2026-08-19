<template>
    <main class="chat-page">
        <h1>Private Chat</h1>

        <p v-if="websocket.connected">
            WebSocket: Connected
        </p>

        <p v-else>
            WebSocket: Disconnected
        </p>

        <p v-if="websocket.error">
            {{ websocket.error }}
        </p>

        <div class="chat-layout">
            <!-- LEFT SIDE: USERS -->
            <aside class="chat-users">
                <h2>Chats</h2>

                <p v-if="loadingUsers">
                    Loading users...
                </p>

                <p v-if="usersError">
                    {{ usersError }}
                </p>

                <p v-if="
                    !loadingUsers &&
                    chatUsers.length === 0
                ">
                    No users available for private chat.
                </p>

                <button v-for="user in chatUsers" :key="user.id" class="chat-user-button" :class="{
                    active:
                        selectedUser &&
                        selectedUser.id === user.id
                }" @click="selectUser(user)">
                    {{ displayUserName(user) }}

                    <span v-if="websocket.privateUnreadForUser(user.id) > 0">
                        ({{ websocket.privateUnreadForUser(user.id) }})
                    </span>
                </button>
            </aside>

            <!-- RIGHT SIDE: CONVERSATION -->
            <section class="conversation">
                <div v-if="!selectedUser">
                    <h2>Select a user</h2>

                    <p>
                        Choose someone from the left to open
                        a private conversation.
                    </p>
                </div>

                <template v-else>
                    <h2>
                        {{ displayUserName(selectedUser) }}
                    </h2>

                    <p v-if="loadingMessages">
                        Loading messages...
                    </p>

                    <p v-if="messagesError">
                        {{ messagesError }}
                    </p>

                    <div ref="messagesContainer" class="messages">
                        <p v-if="
                            !loadingMessages &&
                            messages.length === 0
                        ">
                            No messages yet. Start the conversation.
                        </p>

                        <div v-for="message in messages" :key="message.id" class="message-row" :class="{
                            mine:
                                message.sender_id === auth.user?.id,
                            theirs:
                                message.sender_id !== auth.user?.id
                        }">
                            <div class="message-bubble">
                                <strong>
                                    {{
                                        message.sender_id === auth.user?.id
                                            ? "Me"
                                            : message.sender_name
                                    }}
                                </strong>

                                <p>
                                    {{ message.content }}
                                </p>

                                <small>
                                    {{ message.created_at }}
                                </small>
                            </div>
                        </div>
                    </div>

                    <form class="message-form" @submit.prevent="sendMessage">
                        <input v-model="messageInput" type="text" placeholder="Type a message..." autocomplete="off" />

                        <button type="submit" :disabled="!websocket.connected ||
                            !messageInput.trim()
                            ">
                            Send
                        </button>
                    </form>

                    <p v-if="sendError">
                        {{ sendError }}
                    </p>
                </template>
            </section>
        </div>
    </main>
</template>

<script setup>
import { nextTick, onMounted, onUnmounted, ref, watch } from "vue";

import { apiRequest } from "../services/api";
import { useAuthStore } from "../stores/auth";
import { useWebSocketStore } from "../stores/websocket";

const auth = useAuthStore();
const websocket = useWebSocketStore();

const chatUsers = ref([]);
const selectedUser = ref(null);
const messages = ref([]);

const loadingUsers = ref(false);
const loadingMessages = ref(false);

const usersError = ref("");
const messagesError = ref("");
const sendError = ref("");

const messageInput = ref("");

const messagesContainer = ref(null);

function displayUserName(user) {
    if (!user) {
        return "";
    }

    const fullName =
        `${user.first_name} ${user.last_name}`;

    if (user.nickname) {
        return `${fullName} (${user.nickname})`;
    }

    return fullName;
}

async function loadChatUsers() {
    try {
        loadingUsers.value = true;
        usersError.value = "";

        chatUsers.value =
            await apiRequest("/chat/users");
    } catch (err) {
        usersError.value = err.message;
    } finally {
        loadingUsers.value = false;
    }
}

async function selectUser(user) {
    selectedUser.value = user;

    websocket.setActivePrivateChat(user.id);

    messages.value = [];
    messagesError.value = "";
    sendError.value = "";

    await loadMessages(user.id);
}

async function loadMessages(userId) {
    try {
        loadingMessages.value = true;
        messagesError.value = "";

        messages.value =
            await apiRequest(
                `/chat/${userId}/messages`
            );

        await scrollToBottom();
    } catch (err) {
        messagesError.value = err.message;
    } finally {
        loadingMessages.value = false;
    }
}

async function scrollToBottom() {
    await nextTick();

    if (!messagesContainer.value) {
        return;
    }

    messagesContainer.value.scrollTop =
        messagesContainer.value.scrollHeight;
}

function handleWebSocketMessage(event) {
    let wsEvent;

    try {
        wsEvent = JSON.parse(event.data);
    } catch {
        socketError.value =
            "Received invalid WebSocket data";

        return;
    }

    if (wsEvent.type === "error") {
        sendError.value =
            wsEvent.error || "WebSocket error";

        return;
    }

    if (wsEvent.type === "private_message") {
        handleIncomingPrivateMessage(
            wsEvent.data
        );
    }
}

function handleIncomingPrivateMessage(message) {
    if (!message) {
        return;
    }

    const currentUserID = auth.user?.id;
    const selectedUserID =
        selectedUser.value?.id;

    if (!currentUserID || !selectedUserID) {
        return;
    }

    const belongsToOpenConversation =
        (
            message.sender_id === currentUserID &&
            message.receiver_id === selectedUserID
        )
        ||
        (
            message.sender_id === selectedUserID &&
            message.receiver_id === currentUserID
        );

    if (!belongsToOpenConversation) {
        return;
    }

    const alreadyExists =
        messages.value.some(
            existingMessage =>
                existingMessage.id === message.id
        );

    if (alreadyExists) {
        return;
    }

    messages.value.push(message);

    scrollToBottom();
}

function sendMessage() {
    sendError.value = "";

    if (!selectedUser.value) {
        sendError.value =
            "select a user first";

        return;
    }

    const content =
        messageInput.value.trim();

    if (!content) {
        return;
    }

    const sent = websocket.send({
        type: "private_message",
        receiver_id: selectedUser.value.id,
        content,
    });

    if (!sent) {
        sendError.value =
            websocket.error;

        return;
    }

    messageInput.value = "";
}

watch(
    () => websocket.eventVersion,
    () => {
        const event =
            websocket.lastEvent;

        if (!event) {
            return;
        }

        if (event.type === "private_message") {
            handleIncomingPrivateMessage(
                event.data
            );
        }
    }
);

onMounted(async () => {
    await loadChatUsers();
});

onUnmounted(() => {
    websocket.setActivePrivateChat(null);
});
</script>

<style scoped>
.chat-page {
    padding: 20px;
}

.chat-layout {
    display: grid;
    grid-template-columns: 250px 1fr;
    gap: 20px;
    min-height: 600px;
}

.chat-users {
    border-right: 1px solid #ccc;
    padding-right: 15px;
}

.chat-user-button {
    display: block;
    width: 100%;
    margin-bottom: 8px;
    padding: 10px;
    text-align: left;
}

.chat-user-button.active {
    font-weight: bold;
}

.conversation {
    display: flex;
    flex-direction: column;
}

.messages {
    height: 450px;
    overflow-y: auto;
    border: 1px solid #ccc;
    padding: 15px;
    margin-bottom: 15px;
}

.message-row {
    display: flex;
    margin-bottom: 12px;
}

.message-row.mine {
    justify-content: flex-end;
}

.message-row.theirs {
    justify-content: flex-start;
}

.message-bubble {
    max-width: 70%;
    border: 1px solid #ccc;
    border-radius: 10px;
    padding: 10px;
}

.message-bubble p {
    margin: 5px 0;
}

.message-form {
    display: flex;
    gap: 10px;
}

.message-form input {
    flex: 1;
    padding: 10px;
}
</style>