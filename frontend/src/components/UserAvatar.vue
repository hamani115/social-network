<template>
  <span class="user-avatar-component">
    <img
      v-if="avatarPath"
      :src="avatarPath"
      :alt="name ? `${name} avatar` : 'User avatar'"
    />

    <span v-else class="user-avatar-initials">
      {{ nameInitials(name) }}
    </span>
  </span>
</template>

<script setup>
import { nameInitials } from "../utils/avatar";
defineProps({
  avatarPath: {
    type: String,
    default: "",
  },

  name: {
    type: String,
    default: "",
  },
});
</script>

<style scoped>
.user-avatar-component {
  display: grid;
  place-items: center;

  overflow: hidden;
}

.user-avatar-component img {
  width: 100%;
  height: 100%;
  display: block;
  margin: 0;
  border-radius: 0;
  object-fit: cover;
}

.user-avatar-initials {
  width: 100%;
  height: 100%;

  display: grid;
  place-items: center;
}
</style>
