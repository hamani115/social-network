<template>
  <main>
    <h1>Register</h1>

    <form @submit.prevent="submitRegister">
      <div>
        <label>Email</label>
        <input v-model="form.email" type="email" />
      </div>

      <div>
        <label>Password</label>
        <input v-model="form.password" type="password" />
      </div>

      <div>
        <label>First Name</label>
        <input v-model="form.first_name" type="text" />
      </div>

      <div>
        <label>Last Name</label>
        <input v-model="form.last_name" type="text" />
      </div>

      <div>
        <label>Date of Birth</label>
        <input v-model="form.date_of_birth" type="date" />
      </div>

      <div>
        <label>Nickname Optional</label>
        <input v-model="form.nickname" type="text" />
      </div>

      <div>
        <label>About Me Optional</label>
        <textarea v-model="form.about_me"></textarea>
      </div>

      <button type="submit">Register</button>
    </form>

    <p v-if="message">{{ message }}</p>
    <p v-if="error">{{ error }}</p>
  </main>
</template>

<script setup>
import { reactive, ref } from "vue";
import { useRouter } from "vue-router";
import { useAuthStore } from "../stores/auth";

const router = useRouter();
const auth = useAuthStore();

const message = ref("");
const error = ref("");

const form = reactive({
  email: "",
  password: "",
  first_name: "",
  last_name: "",
  date_of_birth: "",
  nickname: "",
  about_me: "",
});

async function submitRegister() {
  try {
    error.value = "";
    message.value = "";

    await auth.register(form);

    message.value = "Registered successfully. You can now login.";

    router.push("/login");
  } catch (err) {
    error.value = err.message;
  }
}

</script>
