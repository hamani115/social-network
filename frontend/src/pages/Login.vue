<template>
  <main>
    <h1>Login</h1>

    <form @submit.prevent="submitLogin">
      <div>
        <label>Email</label>
        <input v-model="email" type="email" />
      </div>

      <div>
        <label>Password</label>
        <input v-model="password" type="password" />
      </div>

      <button type="submit">Login</button>
    </form>

    <p v-if="error">{{ error }}</p>
  </main>
</template>

<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import { useAuthStore } from "../stores/auth";

const router = useRouter();
const auth = useAuthStore();

const email = ref("");
const password = ref("");
const error = ref("");

async function submitLogin() {
  try {
    error.value = "";

    await auth.login(email.value, password.value);

    router.push("/");
  } catch (err) {
    error.value = err.message;
  }
}

</script>
