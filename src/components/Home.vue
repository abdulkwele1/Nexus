<template>
  <div>
    <div class="container">
      <div class="box">
        <img src="../assets/locationIcon.jpg" alt="Location Icon" class="box-image" />
        <div class="box-text">Locations</div>
      </div>
      <div class="box">
        <img src="../assets/sensorIcon.jpg" alt="Sensor Icon" class="box-image" />
        <div class="box-text">Sensors</div>
      </div>
      <div class="box">
        <img src="../assets/profileIcon.jpg" alt="Profile Icon" class="box-image" />
        <div class="box-text">Settings</div>
      </div>
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
}
</style>
