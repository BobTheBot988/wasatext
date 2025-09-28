<script>
import HeaderDefault from "./HeaderDefault.vue";

export default {
  components: {
    HeaderDefault,
  },
  props: ["userName"],
};
</script>
<template>
  <HeaderDefault style="cursor: not-allowed">
    <div class="username">
      {{ userName }}
    </div>
  </HeaderDefault>
</template>
