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
/* Importing Google Fonts */
@import url('https://fonts.googleapis.com/css2?family=Roboto:wght@500&display=swap');

/* =======================
   Nav Bar Styles
   ======================= */
.navbar {
  position: relative; /* or fixed if you want it always on top */
  width: 100%;
  height: 60px;
  background-color: #f8f8f8; 
  display: flex;
  align-items: center;
  /* If you want the navbar to stay at the top:
     position: fixed;
     top: 0;
     left: 0; 
     z-index: 999;
  */
}

.navbar-content {
  display: flex;
  gap: 20px;
  margin-left: 20px; /* spacing from the left edge */
}

.navbar-content a {
  text-decoration: none;
  color: #333;
  font-weight: 500;
}

/* =======================
   Main Container
   ======================= */
.container {
  /* If navbar is fixed, add top padding. For example:
  padding-top: 60px; 
  */
  display: flex;
  flex-wrap: wrap; /* allow boxes to wrap to new line on smaller screens */
  justify-content: center;
  align-items: center;
  min-height: calc(80vh - 60px); /* space for content */
  gap: 50px; /* spacing between boxes */
  margin: 0 auto;
}

/* =======================
   Box Styles
   ======================= */
.box-link {
  text-decoration: none; /* remove default link underline */
}

.box {
  width: 200px;       /* fixed width on desktop */
  height: 300px;      /* fixed height on desktop */
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  border: 2px solid transparent;
  border-radius: 10px;
  cursor: pointer;
  overflow: hidden;
  transition: transform 0.3s ease, border-color 0.3s ease, box-shadow 0.3s ease;
}

/* Hover effect for boxes */
.box:hover {
  transform: scale(1.05);
  border-color: #ccc;
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
}

/* Style for the image inside the box */
.box-image {
  width: 100%;
  height: auto; /* maintain aspect ratio */
}

/* Text inside the box */
.box-text {
  margin-top: 10px;
  font-size: 18px;
  text-align: center;
  font-family: 'Roboto', sans-serif;
  font-weight: 500;
  letter-spacing: 0.5px;
  text-transform: uppercase;
  color: #333;
}

/* =======================
   Responsive Media Queries
   ======================= */

/* For tablets and below (max-width: 768px) */
@media screen and (max-width: 768px) {
  .container {
    flex-direction: column;   /* stack boxes vertically */
    gap: 20px;               /* reduce gap */
    min-height: auto;        /* let content grow */
  }

  .box {
    width: 60%;  /* narrower than full width, but bigger than 200px */
    max-width: 300px; /* so it doesn’t grow too big on mid-sized screens */
    height: auto;     /* auto-height so images are not squashed */
  }
}

/* For phones (max-width: 480px) */
@media screen and (max-width: 480px) {
  .box {
    width: 80%; /* can be even more narrow on phones */
    margin: 0 auto; /* center the box horizontally */
  }

  /* Font scaling for smaller screens, optional */
  .box-text {
    font-size: 16px;
  }
}
</style>
