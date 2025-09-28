<script>
</script>
<template>
  <div id="container-conv-list">
    <div id="search-convs">
      <slot name="search-convs">
        search-bar
      </slot>
    </div>
    <div id="list">
      <slot>
        search-list
      </slot>
    </div>
  </div>
</template>