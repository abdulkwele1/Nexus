import { defineStore } from 'pinia'
const {
    VITE_NEXUS_API_URL,
  } = import.meta.env;

export const useNexusStore = defineStore('nexus', {
  state: () => {
    return {
        user: new User(VITE_NEXUS_API_URL),
    }
  }
})



class User {
    baseURL: string;
    userName: string;
    loggedIn: boolean;

    constructor(baseURL: any) {
        this.baseURL = baseURL
        this.userName = ""
        // Initialize loggedIn state from cookie
        this.loggedIn = this.getCookie() ? true : false
    }

    async login(userName: any, password: any): Promise<any> {
        const loginAPIUrl = `${VITE_NEXUS_API_URL}/login`
        const data = {
            username: userName,
            password: password,
          };

          const response = await fetch(loginAPIUrl, {
            headers: {
              'Content-Type': 'application/json',
            },
            method: 'POST',
            body: JSON.stringify(data),
          });

          return response
    }

    async logout() {
        const response = await fetch(`${this.baseURL}/logout`, {
            method: 'POST',
            credentials: 'include',
            headers: {
              'Content-Type': 'application/json',
            }
          });

          if (!response.ok) {
            const errorResponse = await response.json();
            console.error('Error response:', errorResponse); // Log the error for better visibility

            return false;
          }

          this.loggedIn = false

          return true
    }

    async getPanelYieldData(panelId: number, startDate: string, endDate: string,): Promise<any> {
      const url = `${VITE_NEXUS_API_URL}/panels/${panelId}/yield_data`

      const response = await fetch(url, {
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        method: 'GET',
        });

        return response
  }

  async getPanelConsumptionData(panelId: number, startDate: string, endDate: string,): Promise<any> {
      const url = `${VITE_NEXUS_API_URL}/panels/${panelId}/consumption_data`

      const response = await fetch(url, {
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        method: 'GET',
        });

        return response
  }

  async setPanelYieldData(panelId: number, startDate: string, endDate: string, yieldData: object): Promise<any> {
      const url = `${VITE_NEXUS_API_URL}/panels/${panelId}/yield_data`

      let payload = {
        yield_data: yieldData

      }
      
      const response = await fetch(url, {
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        method: 'POST',
        body: JSON.stringify(payload),
        });

        return response
  }


  async setPanelConsumptionData(panelId: number, startDate: string, endDate: string, consumptionData: object): Promise<any> {
      const url = `${VITE_NEXUS_API_URL}/panels/${panelId}/consumption_data`
      
      let payload = {
        consumption_data: consumptionData

      }
    
      const response = await fetch(url, {
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        method: 'POST',
        body: JSON.stringify(payload),
        });

        return response
}
  async setSensorTemperatureData(sensorId: string, date: string, soilTemperature: number): Promise<any> {
    const url = `${VITE_NEXUS_API_URL}/sensors/${sensorId}/temperature_data`

    const response = await fetch(url, {
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      },
      method: 'POST',
      body: JSON.stringify({
        date: date,
        soil_temperature: soilTemperature,
      }),
    });

    return response;
  }

  async setSensorMoistureData(sensorId: string, date: string, soilMoisture: number): Promise<any> {
    const url = `${VITE_NEXUS_API_URL}/sensors/${sensorId}/moisture_data`

    const response = await fetch(url, {
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      },
      method: 'POST',
      body: JSON.stringify({
        date: date,
        soil_moisture: soilMoisture,
      }),
    });

    return response;
  }


    async getSensorTemperatureData(sensorId: string) {
      try {
        const response = await fetch(`${this.baseURL}/sensors/${sensorId}/temperature_data`, {
          credentials: 'include',
        });
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        const jsonData = await response.json();
        
        // Handle different possible response formats
        if (Array.isArray(jsonData)) {
          return jsonData;
        }
        if (jsonData && Array.isArray(jsonData.sensor_temperature_data)) {
          return jsonData.sensor_temperature_data;
        }
        if (jsonData && Array.isArray(jsonData.data)) {
          return jsonData.data;
        }
        
        console.warn(`Temperature data for sensor ${sensorId} received in unexpected format:`, jsonData);
        return [];
      } catch (error) {
        console.error('Error fetching sensor temperature data:', error);
        return [];
      }
    }

    async getSensorMoistureData(sensorId: string) {
      try {
        const response = await fetch(`${this.baseURL}/sensors/${sensorId}/moisture_data`, {
          credentials: 'include',
        });
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        const jsonData = await response.json();
        
        // Handle different possible response formats
        if (Array.isArray(jsonData)) {
          return jsonData;
        }
        if (jsonData && Array.isArray(jsonData.sensor_moisture_data)) {
          return jsonData.sensor_moisture_data;
        }
        if (jsonData && Array.isArray(jsonData.data)) {
          return jsonData.data;
        }
        
        console.warn(`Moisture data for sensor ${sensorId} received in unexpected format:`, jsonData);
        return [];
      } catch (error) {
        console.error('Error fetching sensor moisture data:', error);
        return [];
      }
    }

   async logPanelData(userName: any, password: any): Promise<any> {
    const loginAPIUrl = `${VITE_NEXUS_API_URL}/login`
    const data = {
        username: userName,
        password: password,
      };

      const response = await fetch(loginAPIUrl, {
        headers: {
          'Content-Type': 'application/json',
        },
        method: 'POST',
        body: JSON.stringify(data),
      });

      return response
}

  async getDroneImages(startDate?: Date, endDate?: Date): Promise<any> {
    try {
      let url = `${this.baseURL}/drone_images`;
      
      // Add date range parameters if provided
      if (startDate && endDate) {
        url += `?start_date=${startDate.toISOString()}&end_date=${endDate.toISOString()}`;
      }

      const response = await fetch(url, {
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        method: 'GET',
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const jsonData = await response.json();
      const images = Array.isArray(jsonData.images) ? jsonData.images : [];
      
      // Process each image to create object URLs
      const processedImages = await Promise.all(images.map(async (img: any) => {
        try {
          const blob = await this.getImageContent(img.id);
          return {
            ...img,
            url: URL.createObjectURL(blob)
          };
        } catch (error) {
          console.error(`Failed to load image ${img.id}:`, error);
          return img;
        }
      }));

      return processedImages;
    } catch (error) {
      console.error('Error fetching drone images:', error);
      return [];
    }
  }

  private async getImageContent(imageId: string): Promise<Blob> {
    const response = await fetch(`${this.baseURL}/drone_images/${imageId}/content`, {
      credentials: 'include',
    });
    if (!response.ok) {
      throw new Error(`Failed to fetch image content: ${response.status}`);
    }
    return await response.blob();
  }

  async uploadDroneImage(imageFile: File, metadata: { location: string, timestamp: string }): Promise<any> {
    try {
      const formData = new FormData();
      formData.append('images', imageFile);
      formData.append('description', metadata.location);
      formData.append('metadata', JSON.stringify({
        location: metadata.location,
        timestamp: metadata.timestamp,
        original_name: imageFile.name,
        mime_type: imageFile.type // Add MIME type
      }));

      const response = await fetch(`${this.baseURL}/drone_images`, {
        credentials: 'include',
        method: 'POST',
        body: formData
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const result = await response.json();
      
      // Return the result with a temporary URL for immediate display
      if (result.uploaded_images && result.uploaded_images.length > 0) {
        const image = result.uploaded_images[0];
        return {
          ...image,
          url: URL.createObjectURL(imageFile) // Create temporary URL for immediate display
        };
      }
      return result;
    } catch (error) {
      console.error('Error uploading drone image:', error);
      throw error;
    }
  }

  async deleteDroneImage(imageId: string): Promise<any> {
    try {
      const response = await fetch(`${this.baseURL}/drone_images/${imageId}`, {
        credentials: 'include',
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      return true;
    } catch (error) {
      console.error('Error deleting drone image:', error);
      throw error;
    }
  }

    // Function to get a cookie by name
    getCookie(): any {
        const value = `; ${document.cookie}`;
        const parts = value.split(`; session_id=`);
        if (parts.length === 2) {
            const sessionId = parts?.pop()?.split(';')?.shift();
            if (sessionId) {
                this.loggedIn = true;
                return sessionId;
            }
        }
        this.loggedIn = false;
        return false;
  }

  async getAllSensors() {
    try {
      const response = await fetch(`${this.baseURL}/sensors`, {
        credentials: 'include',
      });
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      return await response.json();
    } catch (error) {
      console.error('Error fetching all sensors:', error);
      return [];
    }
  }
}
