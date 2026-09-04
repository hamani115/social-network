<template>
  <section class="profile-card">
    <div class="profile-cover"></div>

    <div class="profile-header-content">
      <div class="avatar-wrapper">
        <UserAvatar
          :avatar-path="profile.avatar_path"
          :name="`${profile.first_name} ${profile.last_name}`"
          class="profile-avatar"
          :class="{
            'profile-avatar-placeholder': !profile.avatar_path,
          }"
        />
      </div>

      <div class="profile-main-info">
        <div class="profile-title-row">
          <div>
            <h1 class="profile-name">
              {{ profile.first_name }}
              {{ profile.last_name }}
            </h1>

            <p
              v-if="profile.can_view_profile && profile.nickname"
              class="profile-nickname"
            >
              {{ profile.nickname }}
            </p>
          </div>

          <span class="privacy-badge" :class="{ private: !profile.is_public }">
            {{ profile.is_public ? "Public profile" : "Private profile" }}
          </span>
        </div>

        <p
          v-if="profile.can_view_profile && profile.about_me"
          class="profile-headline"
        >
          {{ profile.about_me }}
        </p>

        <div v-if="profile.can_view_profile" class="profile-details-inline">
          <span v-if="profile.email">
            <strong>Email:</strong>
            {{ profile.email }}
          </span>

          <span v-if="profile.date_of_birth">
            <strong>Date of birth:</strong>
            {{ formatDateOfBirth(profile.date_of_birth) }}
          </span>
        </div>

        <div v-if="profile.can_view_profile" class="profile-stats">
          <span>
            <strong>
              {{ profile.followers_count }}
            </strong>
            followers
          </span>

          <span class="stat-divider">-</span>

          <span>
            <strong>
              {{ profile.following_count }}
            </strong>
            following
          </span>
        </div>

        <!-- OTHER USER PROFILE -->
        <div v-if="!profile.is_owner" class="profile-actions">
          <button
            v-if="profile.follow_status === 'none'"
            class="primary-button"
            type="button"
            @click="$emit('follow')"
          >
            <i class="fa-solid fa-user-plus" aria-hidden="true"></i>
            Follow
          </button>

          <button
            v-else-if="profile.follow_status === 'following'"
            class="secondary-button following-button"
            type="button"
            @click="$emit('unfollow')"
          >
            <i class="fa-solid fa-user-check" aria-hidden="true"></i>
            Following
          </button>

          <button
            v-else-if="profile.follow_status === 'pending'"
            class="secondary-button pending-button"
            type="button"
            title="Click to cancel follow request"
            @click="$emit('unfollow')"
          >
            Request pending
          </button>
        </div>

        <div v-if="profile.is_owner" class="profile-actions">
          <button class="secondary-button" type="button" @click="$emit('edit')">
            Edit profile
          </button>
        </div>
      </div>
    </div>
  </section>

  <!-- PRIVATE PROFILE -->
  <section
    v-if="!profile.can_view_profile && !profile.is_owner"
    class="locked-profile-card"
  >
    <div class="locked-profile-icon">
      <i class="fa-solid fa-lock" aria-hidden="true"></i>
    </div>

    <div class="locked-profile-content">
      <h2>This profile is private</h2>

      <p v-if="profile.follow_status === 'none'">
        Follow {{ profile.first_name }} to see their profile information, connections, and activity
      </p>

      <p v-else-if="profile.follow_status === 'pending'">
        Your follow request has been sent
      </p>
    </div>

    <button
      v-if="profile.follow_status === 'none'"
      type="button"
      class="primary-button locked-follow-button"
      @click="$emit('follow')"
    >
      <i class="fa-solid fa-user-plus" aria-hidden="true"></i>
      Follow
    </button>

    <button
      v-else-if="profile.follow_status === 'pending'"
      type="button"
      class="secondary-button locked-follow-button"
      @click="$emit('unfollow')"
    >
      Cancel request
    </button>
  </section>
</template>

<script setup>
import UserAvatar from "../UserAvatar.vue";
import { formatDateOfBirth } from "../../utils/date";
defineProps({
  profile: {
    type: Object,
    required: true,
  },
});
defineEmits(["follow", "unfollow", "edit"]);
</script>

<style scoped>
.profile-card {
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 16px;
  background: rgba(20, 25, 35, 0.95);
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.22);
}

.profile-cover {
  height: 190px;
  background:
    radial-gradient(
      circle at 20% 20%,
      rgba(79, 156, 255, 0.45),
      transparent 35%
    ),
    linear-gradient(135deg, #17253f, #0d1524 55%, #111827);
}

.profile-header-content {
  position: relative;
  display: flex;
  gap: 28px;
  padding: 0 32px 30px;
}

.avatar-wrapper {
  flex-shrink: 0;
  margin-top: -72px;
}

.profile-avatar {
  width: 150px;
  height: 150px;
  border: 5px solid #111827;
  border-radius: 50%;
  object-fit: cover;
  background: #1f2937;
  box-shadow: 0 5px 18px rgba(0, 0, 0, 0.4);
  overflow: hidden;
}

.profile-avatar-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  background: linear-gradient(135deg, #4f9cff, #2867d6);
  font-size: 42px;
  font-weight: 700;
}

.profile-main-info {
  flex: 1;
  min-width: 0;
  padding-top: 22px;
}

.profile-title-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20px;
}

.profile-name {
  margin: 0;
  color: #f8fafc;
  font-size: 30px;
  line-height: 1.2;
}

.profile-nickname {
  margin: 6px 0 0;
  color: #94a3b8;
  font-size: 15px;
}

.profile-headline {
  max-width: 720px;
  margin: 16px 0;
  color: #cbd5e1;
  line-height: 1.6;
}

.profile-details-inline {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 22px;
  margin-top: 14px;
  color: #94a3b8;
  font-size: 13px;
}

.profile-details-inline strong {
  color: #cbd5e1;
  font-weight: 600;
}

.profile-stats {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 14px;
  color: #94a3b8;
}

.profile-stats strong {
  color: #4f9cff;
}

.stat-divider {
  color: #64748b;
}

.privacy-badge {
  padding: 7px 12px;
  border: 1px solid rgba(79, 156, 255, 0.35);
  border-radius: 999px;
  color: #80b7ff;
  background: rgba(79, 156, 255, 0.1);
  font-size: 13px;
  font-weight: 600;
  white-space: nowrap;
}

.privacy-badge.private {
  color: #cbd5e1;
  border-color: rgba(148, 163, 184, 0.25);
  background: rgba(148, 163, 184, 0.08);
}

.profile-actions {
  display: flex;
  gap: 10px;
  margin-top: 20px;
}

.primary-button,
.secondary-button {
  min-width: 110px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 7px;
  padding: 9px 18px;
  border-radius: 999px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition:
    background 0.2s,
    border-color 0.2s,
    transform 0.2s;
}

.primary-button {
  border: 1px solid #4f9cff;
  color: #07111f;
  background: #4f9cff;
}

.primary-button:hover {
  background: #72afff;
  transform: translateY(-1px);
}

.secondary-button {
  border: 1px solid rgba(79, 156, 255, 0.65);
  color: #73afff;
  background: transparent;
}

.secondary-button:hover {
  background: rgba(79, 156, 255, 0.1);
}

.locked-profile-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-top: 20px;
  padding: 42px 30px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 16px;
  background: rgba(20, 25, 35, 0.95);
  text-align: center;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.18);
}

.locked-profile-icon {
  width: 64px;
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 18px;
  border-radius: 50%;
  background: rgba(79, 156, 255, 0.08);
  font-size: 26px;
}

.locked-profile-content {
  max-width: 520px;
}

.locked-profile-content h2 {
  margin-bottom: 8px;
}

.locked-profile-content p {
  margin: 0;
  color: #9ca6b3;
  line-height: 1.6;
}

.locked-follow-button {
  margin-top: 22px;
}

.following-button {
  color: #36c98f;
  border-color: rgba(54, 201, 143, 0.35);
}

.following-button:hover {
  color: #ff7883;
  border-color: rgba(255, 95, 109, 0.4);
  background: rgba(255, 95, 109, 0.08);
}

.pending-button {
  color: #f4b942;
  border-color: rgba(244, 185, 66, 0.35);
  background: rgba(244, 185, 66, 0.06);
}

.pending-button:hover {
  color: #ff7883;
  border-color: rgba(255, 95, 109, 0.4);
  background: rgba(255, 95, 109, 0.08);
}

@media (max-width: 700px) {
  .profile-cover {
    height: 140px;
  }

  .profile-header-content {
    flex-direction: column;
    gap: 10px;
    padding: 0 20px 24px;
  }

  .avatar-wrapper {
    margin-top: -60px;
  }

  .profile-avatar {
    width: 120px;
    height: 120px;
  }

  .profile-main-info {
    padding-top: 0;
  }

  .profile-title-row {
    flex-direction: column;
    gap: 10px;
  }

  .profile-name {
    font-size: 25px;
  }

  .locked-profile-card {
    padding: 36px 20px;
  }

  .locked-profile-icon {
    width: 58px;
    height: 58px;
    font-size: 24px;
  }
}
</style>
