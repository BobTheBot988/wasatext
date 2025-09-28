<script>
import User from "../components/User.vue"
export default {
    components: {
        User,
    },
    props: ["user"]
}
</script>
<template>
  <User :user="user" />
</template>