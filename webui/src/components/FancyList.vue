<script>
export default {
  props: ['items'],
  mounted() {
  }
}
</script>
<template>
  <ul>
    <li v-for="item in items" :key="item">
      <slot name="item" v-bind="item" />
    </li>
  </ul>
</template>