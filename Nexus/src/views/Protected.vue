<template>
  <p>{{ message }}</p>
</template>

<script>

export default {
  data () {
    return {
      message: ''
    }
  },
  mounted () {
    this.message = "you are logged in"
     
  }
}
</script>

<style>
</style>

