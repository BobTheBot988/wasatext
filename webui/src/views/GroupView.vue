<script>

import Group from '../components/Group.vue';

export default {
    components: {
        Group
    },
    props: ['group']
}
</script>
<template>
  <div id="container-group-info">
    <Group :group="group" />
  </div>
</template>