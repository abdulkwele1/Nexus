<template>
  <div>
    <div class="container">
      <RouterLink to="/sensors" class="box-link">
        <div class="box">
          <img src="../assets/sensorIcon.jpg" alt="Sensor Icon" class="box-image" />
          <div class="box-text">Sensors</div>
        </div>
      </RouterLink>
      <RouterLink to="/solar" class="box-link">
        <div class="box">
          <img src="../assets/solarIcon.jpg" alt="Solar Icon" class="box-image" />
          <div class="box-text">Solar Panels</div>
        </div>
      </RouterLink>
      <!-- New Drone Box -->
      <RouterLink to="/drone" class="box-link">
        <div class="box">
          <img src="../assets/droneIcon.png" alt="Drone Icon" class="box-image" />
          <div class="box-text">Drone</div>
        </div>
      </RouterLink>
    </div>
  </div>
</template>


<style scoped>
/* Container style to center the boxes */
.container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: calc(80vh - 60px); /* Adjust height for navbar */
  gap: 50px; /* Spacing between boxes */
  transform: translateX(5%);
}

/* Style for each box */
.box {
  width: 200px;
  height: 300px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  border: 2px solid transparent; /* Initially hide the border */
  border-radius: 10px;
  cursor: pointer;
  overflow: hidden;
  transition: transform 0.3s ease, border-color 0.3s ease, box-shadow 0.3s ease;
}

/* Hover effect for boxes */
.box:hover {
  transform: scale(1.05);
  border-color: #ccc; /* Show the border on hover */
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
}

/* Style for the image inside the box */
.box-image {
  width: 100%;
  height: auto;
}

/* Style for the text inside the box */
.box-text {
  margin-top: 10px;
  font-size: 16px;
  text-align: center;
  font-family: 'MyCustomFont', sans-serif; 
}
</style>

<script setup>
import { onMounted } from 'vue'

// run every time page (component) loads
// https://learnvue.co/articles/vue-lifecycle-hooks-guide
onMounted(() => {
    console.log('hommie')
    // If user is on home page and document.cookie == “” is true, redirect the user to the login page
})
</script>
