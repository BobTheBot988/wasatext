<script>
export default {
    props: ['search', 'placeholder']
}
</script>
<template>
  <div id="search-form">
    <input
      id="form-text" type="text" :value="search" :placeholder="{ placeholder }"
      autofocus @input="$emit('update:search', $event.target.search)"
    >
  </div>
</template>