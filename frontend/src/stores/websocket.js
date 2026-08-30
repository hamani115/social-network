import { defineStore } from "pinia";
let socket = null;
export const useWebSocketStore = defineStore("websocket", {
  state: () => ({
    connected: false,
    error: "",

    lastEvent: null,
    eventVersion: 0,

    currentUserID: null,
    onlineUserIDs: {},

    activePrivateChatUserID: null,
    privateUnreadByUser: {},

    activeGroupID: null,
    groupUnreadByGroup: {},
  }),

  getters: {
    isUserOnline: (state) => {
      return (userID) => {
        const id = Number(userID);
        if (!id) {
          return false;
        }
        return Boolean(state.onlineUserIDs[id]);
      };
    },
    privateUnreadTotal(state) {
      return Object.values(state.privateUnreadByUser).reduce(
        (total, count) => total + count,
        0,
      );
    },

    privateUnreadForUser: (state) => {
      return (userID) => state.privateUnreadByUser[userID] || 0;
    },

    groupUnreadTotal(state) {
      return Object.values(state.groupUnreadByGroup).reduce(
        (total, count) => total + count,
        0,
      );
    },

    groupUnreadForGroup: (state) => {
      return (groupID) => state.groupUnreadByGroup[groupID] || 0;
    },
  },

  actions: {
    setPresenceSnapshot(data) {
      const userIDs = data?.user_ids;
      if (!Array.isArray(userIDs)) {
        return;
      }
      const onlineUsers = {};
      for (const userID of userIDs) {
        const id = Number(userID);
        if (id > 0) {
          onlineUsers[id] = true;
        }
      }
      this.onlineUserIDs = onlineUsers;
    },

    updatePresence(data) {
      const userID = Number(data?.user_id);
      if (!userID) {
        return;
      }
      if (data.online) {
        this.onlineUserIDs[userID] = true;
        return;
      }
      delete this.onlineUserIDs[userID];
    },
    connect(userID) {
      this.currentUserID = userID;
      if (
        socket &&
        (socket.readyState === WebSocket.OPEN ||
          socket.readyState === WebSocket.CONNECTING)
      ) {
        return;
      }
      this.error = "";
      const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
      const url = `${protocol}//${window.location.host}/api/ws`;
      socket = new WebSocket(url);
      socket.onopen = () => {
        console.log("GLOBAL WS OPEN");
        this.connected = true;
        this.error = "";
      };
      socket.onclose = (event) => {
        console.log(
          "GLOBAL WS CLOSED:",
          event.code,
          event.reason,
          event.wasClean,
        );
        this.connected = false;
        this.onlineUserIDs = {};
        socket = null;
      };
      socket.onerror = (event) => {
        console.error("GLOBAL WS ERROR:", event);
        this.error = "WebSocket connection error";
      };
      socket.onmessage = (event) => {
        this.handleMessage(event);
      };
    },

    handleMessage(event) {
      let parsed;
      try {
        parsed = JSON.parse(event.data);
      } catch {
        this.error = "Received invalid WebSocket data";
        return;
      }
      if (parsed.type === "error") {
        this.error = parsed.error || "WebSocket error";
        return;
      }
      if (parsed.type === "presence_snapshot") {
        this.setPresenceSnapshot(parsed.data);
        return;
      }
      if (parsed.type === "presence") {
        this.updatePresence(parsed.data);
        return;
      }
      this.trackPrivateUnread(parsed);
      this.trackGroupUnread(parsed);
      this.lastEvent = parsed;
      this.eventVersion++;
    },

    send(data) {
      this.error = "";
      if (!socket || socket.readyState !== WebSocket.OPEN) {
        this.error = "WebSocket is not connected";
        return false;
      }
      socket.send(JSON.stringify(data));
      return true;
    },

    disconnect() {
      if (socket) {
        socket.close();
        socket = null;
      }
      this.connected = false;
      this.lastEvent = null;
      this.currentUserID = null;
      this.onlineUserIDs = {};
      this.activePrivateChatUserID = null;
      this.activeGroupID = null;
      this.privateUnreadByUser = {};
      this.groupUnreadByGroup = {};
    },

    trackPrivateUnread(event) {
      if (!event || event.type !== "private_message" || !event.data) {
        return;
      }
      const message = event.data;
      if (message.sender_id === this.currentUserID) {
        return;
      }
      if (message.sender_id === this.activePrivateChatUserID) {
        return;
      }
      const senderID = message.sender_id;
      const currentCount = this.privateUnreadByUser[senderID] || 0;
      this.privateUnreadByUser[senderID] = currentCount + 1;
    },

    setActivePrivateChat(userID) {
      this.activePrivateChatUserID = userID;
      if (userID) {
        this.clearPrivateUnread(userID);
      }
    },

    clearPrivateUnread(userID) {
      delete this.privateUnreadByUser[userID];
    },

    clearAllPrivateUnread() {
      this.privateUnreadByUser = {};
    },

    trackGroupUnread(event) {
      if (!event || event.type !== "group_message" || !event.data) {
        return;
      }
      const message = event.data;
      if (message.sender_id === this.currentUserID) {
        return;
      }
      if (message.group_id === this.activeGroupID) {
        return;
      }
      const groupID = message.group_id;
      const currentCount = this.groupUnreadByGroup[groupID] || 0;
      this.groupUnreadByGroup[groupID] = currentCount + 1;
    },

    setActiveGroup(groupID) {
      this.activeGroupID = groupID;
      if (groupID) {
        this.clearGroupUnread(groupID);
      }
    },

    clearGroupUnread(groupID) {
      delete this.groupUnreadByGroup[groupID];
    },

    clearAllGroupUnread() {
      this.groupUnreadByGroup = {};
    },
  },
});
