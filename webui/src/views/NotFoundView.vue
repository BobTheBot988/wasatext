<script>
export default {
    name: 'NotFoundView',
    methods: {
        goBack() {
            this.$router.go(-1)
        }
    }
}
</script>

<template>
  <div class="not-found-container">
    <h1>404 - Page Not Found</h1>
    <p>The page you're looking for doesn't exist or you may not have access to it.:(</p>

    <div class="actions">
      <button class="btn-medium-importance" @click="goBack">Go Back</button>
      <router-link to="/conversations" class="btn-high-importance">Go to Conversations</router-link>
    </div>
  </div>
</template>
