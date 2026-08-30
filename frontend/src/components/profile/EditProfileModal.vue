<template>
  <Teleport to="body">
    <div class="modal-overlay" @click.self="closeEditModal">
      <section
        class="edit-profile-modal"
        role="dialog"
        aria-modal="true"
        aria-labelledby="edit-profile-title"
      >
        <header class="modal-header">
          <div>
            <h2 id="edit-profile-title">Edit profile</h2>
          </div>

          <button
            type="button"
            class="modal-close-button"
            aria-label="Close edit profile"
            :disabled="updatingProfile"
            @click="closeModal"
          >
            <i class="fa-solid fa-xmark" aria-hidden="true"></i>
          </button>
        </header>

        <!-- FORM -->
        <form class="edit-profile-form" @submit.prevent="updateProfile">
          <div class="modal-body">
            <div class="edit-field">
              <label for="edit-avatar">Profile picture</label>

              <div class="edit-avatar-row">
                <img
                  v-if="profile.avatar_path"
                  :src="profile.avatar_path"
                  alt="Current profile picture"
                  class="edit-avatar-preview"
                />

                <div v-else class="edit-avatar-preview edit-avatar-placeholder">
                  {{ userInitials(profile) }}
                </div>

                <div class="edit-avatar-controls">
                  <input
                    id="edit-avatar"
                    ref="editAvatarInput"
                    type="file"
                    accept="image/png,image/jpeg,image/gif"
                    @change="handleEditAvatarChange"
                  />

                  <small v-if="editAvatar">
                    Selected: {{ editAvatar.name }}
                  </small>

                  <small v-else> JPEG, PNG or GIF </small>
                </div>
              </div>
            </div>

            <div class="edit-name-grid">
              <div class="edit-field">
                <label for="edit-first-name">First name</label>

                <input
                  id="edit-first-name"
                  v-model="editForm.first_name"
                  type="text"
                  required
                />
              </div>

              <div class="edit-field">
                <label for="edit-last-name">Last name</label>

                <input
                  id="edit-last-name"
                  v-model="editForm.last_name"
                  type="text"
                  required
                />
              </div>
            </div>

            <div class="edit-field">
              <label for="edit-email">Email</label>

              <input
                id="edit-email"
                v-model="editForm.email"
                type="email"
                required
              />
            </div>

            <div class="edit-field">
              <label for="edit-date-of-birth">Date of birth</label>

              <input
                id="edit-date-of-birth"
                v-model="editForm.date_of_birth"
                type="date"
                required
              />
            </div>

            <div class="edit-field">
              <label for="edit-nickname">Nickname</label>

              <input
                id="edit-nickname"
                v-model="editForm.nickname"
                type="text"
                placeholder="Add a nickname"
              />
            </div>

            <div class="edit-field">
              <div class="field-heading">
                <label for="edit-about">About</label>

                <span>
                  {{ editForm.about_me.length }}
                </span>
              </div>

              <textarea
                id="edit-about"
                v-model="editForm.about_me"
                rows="6"
                placeholder="Tell people about yourself..."
              ></textarea>
            </div>

            <!-- PRIVACY -->
            <div class="profile-visibility-setting">
              <div>
                <strong>Public profile</strong>

                <p>
                  {{
                    editForm.is_public
                      ? "Anyone can view your full profile"
                      : "Only your followers can view your full profile"
                  }}
                </p>
              </div>

              <label class="visibility-switch">
                <input v-model="editForm.is_public" type="checkbox" />

                <span class="switch-track">
                  <span class="switch-knob"></span>
                </span>
              </label>
            </div>

            <p v-if="updateError" class="modal-error">
              {{ updateError }}
            </p>
          </div>

          <footer class="modal-footer">
            <button
              type="button"
              class="modal-cancel-button"
              :disabled="updatingProfile"
              @click="closeModal"
            >
              Cancel
            </button>

            <button
              type="submit"
              class="modal-save-button"
              :disabled="updatingProfile"
            >
              {{ updatingProfile ? "Saving..." : "Save changes" }}
            </button>
          </footer>
        </form>
      </section>
    </div>
  </Teleport>
</template>

<script setup>
import { ref } from "vue";
import { apiRequest } from "../../services/api";
const props = defineProps({
  profile: {
    type: Object,
    required: true,
  },
});
const emit = defineEmits(["close", "updated"]);
const editForm = ref({
  email: props.profile.email || "",
  first_name: props.profile.first_name || "",
  last_name: props.profile.last_name || "",
  date_of_birth: props.profile.date_of_birth || "",
  nickname: props.profile.nickname || "",
  about_me: props.profile.about_me || "",
  is_public: props.profile.is_public,
});
const editAvatar = ref(null);
const editAvatarInput = ref(null);
const updatingProfile = ref(false);
const updateError = ref("");

function handleEditAvatarChange(event) {
  editAvatar.value = event.target.files?.[0] || null;
}

function userInitials(user) {
  const first = user.first_name?.charAt(0) || "";
  const last = user.last_name?.charAt(0) || "";
  return (first + last).toUpperCase();
}

function closeModal() {
  if (updatingProfile.value) {
    return;
  }
  emit("close");
}

async function updateProfile() {
  try {
    updatingProfile.value = true;
    updateError.value = "";
    const formData = new FormData();
    formData.append("email", editForm.value.email);
    formData.append("first_name", editForm.value.first_name);
    formData.append("last_name", editForm.value.last_name);
    formData.append("date_of_birth", editForm.value.date_of_birth);
    formData.append("nickname", editForm.value.nickname);
    formData.append("about_me", editForm.value.about_me);
    formData.append("is_public", String(editForm.value.is_public));
    if (editAvatar.value) {
      formData.append("avatar", editAvatar.value);
    }
    await apiRequest("/profile/me", {
      method: "PUT",
      body: formData,
    });
    emit("updated");
  } catch (err) {
    updateError.value = err.message;
  } finally {
    updatingProfile.value = false;
  }
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;

  z-index: 1000;

  display: flex;
  align-items: center;
  justify-content: center;

  padding: 24px;

  background: rgba(3, 7, 18, 0.76);

  backdrop-filter: blur(5px);
}

.edit-profile-modal {
  width: min(620px, 100%);

  max-height: calc(100vh - 48px);

  overflow-y: auto;

  border: 1px solid rgba(255, 255, 255, 0.09);

  border-radius: 16px;

  background: #151b26;

  box-shadow: 0 24px 70px rgba(0, 0, 0, 0.55);
}

.modal-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;

  gap: 20px;

  padding: 22px 26px;

  padding-bottom: 10px;

  border-bottom: 1px solid rgba(255, 255, 255, 0.07);
}

.modal-header h2 {
  margin: 0;

  color: #f8fafc;

  font-size: 22px;
}

.modal-close-button {
  display: flex;
  align-items: center;
  justify-content: center;

  flex-shrink: 0;

  width: 36px;
  height: 36px;

  padding: 0;

  border: none;
  border-radius: 50%;

  color: #94a3b8;

  background: transparent;

  font-size: 28px;
  line-height: 1;

  cursor: pointer;
}

.modal-close-button:hover {
  color: #f8fafc;

  background: rgba(255, 255, 255, 0.06);
}

.modal-close-button:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.modal-body {
  display: flex;
  flex-direction: column;

  gap: 24px;

  padding: 26px;
}

.edit-field {
  display: flex;
  flex-direction: column;

  gap: 8px;
}

.edit-field label {
  color: #e2e8f0;

  font-size: 14px;
  font-weight: 600;
}

.edit-field input,
.edit-field textarea {
  width: 100%;

  box-sizing: border-box;

  padding: 11px 13px;

  border: 1px solid rgba(255, 255, 255, 0.1);

  border-radius: 9px;

  outline: none;

  color: #e2e8f0;

  background: rgba(255, 255, 255, 0.035);

  font: inherit;

  transition:
    border-color 0.2s,
    box-shadow 0.2s;
}

.edit-field textarea {
  resize: vertical;

  min-height: 130px;

  line-height: 1.6;
}

.edit-field input:focus,
.edit-field textarea:focus {
  border-color: rgba(79, 156, 255, 0.8);

  box-shadow: 0 0 0 3px rgba(79, 156, 255, 0.1);
}

.edit-field small {
  color: #64748b;

  font-size: 12px;
}

.field-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;

  gap: 12px;
}

.field-heading span {
  color: #64748b;

  font-size: 12px;
}

.edit-avatar-row {
  display: flex;
  align-items: center;

  gap: 16px;
}

.edit-avatar-preview {
  width: 72px;
  height: 72px;

  flex-shrink: 0;

  border: 2px solid rgba(79, 156, 255, 0.25);

  border-radius: 50%;

  object-fit: cover;

  background: #1f2937;
}

.edit-avatar-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;

  color: white;

  background: linear-gradient(135deg, #4f9cff, #2867d6);

  font-weight: 700;
}

.edit-avatar-controls {
  flex: 1;

  min-width: 0;

  display: flex;
  flex-direction: column;

  gap: 7px;
}

.edit-name-grid {
  display: grid;

  grid-template-columns: repeat(2, minmax(0, 1fr));

  gap: 14px;
}

.profile-visibility-setting {
  display: flex;
  align-items: center;
  justify-content: space-between;

  gap: 24px;

  padding: 18px;

  border: 1px solid rgba(255, 255, 255, 0.07);

  border-radius: 12px;

  background: rgba(255, 255, 255, 0.02);
}

.profile-visibility-setting strong {
  color: #e2e8f0;

  font-size: 14px;
}

.profile-visibility-setting p {
  margin: 6px 0 0;

  max-width: 400px;

  color: #64748b;

  font-size: 12px;
  line-height: 1.5;
}

.visibility-switch {
  flex-shrink: 0;

  cursor: pointer;
}

.visibility-switch input {
  position: absolute;

  opacity: 0;

  pointer-events: none;
}

.switch-track {
  position: relative;

  display: block;

  width: 48px;
  height: 26px;

  border-radius: 999px;

  background: #334155;

  transition: background 0.2s;
}

.switch-knob {
  position: absolute;

  top: 3px;
  left: 3px;

  width: 20px;
  height: 20px;

  border-radius: 50%;

  background: white;

  box-shadow: 0 2px 5px rgba(0, 0, 0, 0.25);

  transition: transform 0.2s;
}

.visibility-switch input:checked + .switch-track {
  background: #4f9cff;
}

.visibility-switch input:checked + .switch-track .switch-knob {
  transform: translateX(22px);
}

.modal-error {
  margin: 0;

  padding: 11px 13px;

  border-radius: 8px;

  color: #fca5a5;

  background: rgba(248, 113, 113, 0.08);

  font-size: 13px;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;

  gap: 10px;

  padding: 18px 26px;

  border-top: 1px solid rgba(255, 255, 255, 0.07);
}

.modal-cancel-button,
.modal-save-button {
  padding: 9px 18px;

  border-radius: 999px;

  font-size: 14px;
  font-weight: 600;

  cursor: pointer;
}

.modal-cancel-button {
  border: 1px solid rgba(255, 255, 255, 0.12);

  color: #cbd5e1;

  background: transparent;
}

.modal-cancel-button:hover {
  background: rgba(255, 255, 255, 0.05);
}

.modal-save-button {
  border: 1px solid #4f9cff;

  color: #07111f;

  background: #4f9cff;
}

.modal-save-button:hover:not(:disabled) {
  background: #72afff;
}

.modal-save-button:disabled,
.modal-cancel-button:disabled {
  opacity: 0.55;

  cursor: not-allowed;
}

@media (max-width: 700px) {
  .edit-avatar-row {
    align-items: flex-start;
    flex-direction: column;
  }

  .edit-name-grid {
    grid-template-columns: 1fr;
  }
}
</style>
