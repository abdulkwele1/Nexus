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
        this.loggedIn = false
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
  async setSensorTemperatureData(sensorId: number, date: string, soilTemperature: number): Promise<any> {
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

  async setSensorMoistureData(sensorId: number, date: string, soilMoisture: number): Promise<any> {
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


    async getSensorMoistureData(sensorId: number): Promise<any[]> {
      const url = `${VITE_NEXUS_API_URL}/sensors/${sensorId}/moisture_data`;

      try {
        const response = await fetch(url, {
          credentials: 'include',
        });

        if (!response.ok) {
          console.error(`Error fetching moisture data for sensor ${sensorId}: ${response.status} ${response.statusText}`);
          const errorBody = await response.text();
          console.error("Error body:", errorBody);
          return []; // Return empty array on error
        }

        const jsonData = await response.json();

        // Now, determine where the array of data points is in jsonData.
        // Common patterns:
        // 1. jsonData itself is the array: return jsonData;
        // 2. jsonData is an object like { "sensor_moisture_data": [...] }: return jsonData.sensor_moisture_data;
        // 3. jsonData is an object like { "data": [...] }: return jsonData.data;

        // The console.log in soilMoistureGraph.vue for `rawDataPoints` will help verify this structure.
        // For now, let's try a few common possibilities.
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
        return []; // Return empty if format is not recognized as an array
      } catch (error) {
        console.error(`Exception fetching/parsing moisture data for sensor ${sensorId}:`, error);
        return [];
      }
    }

    async getSensorTemperatureData(sensorId: number): Promise<any[]> {
      const url = `${VITE_NEXUS_API_URL}/sensors/${sensorId}/temperature_data`;
      
      try {
        const response = await fetch(url, {
          credentials: 'include',
        });

        if (!response.ok) {
          console.error(`Error fetching temperature data for sensor ${sensorId}: ${response.status} ${response.statusText}`);
          const errorBody = await response.text();
          console.error("Error body:", errorBody);
          return [];
        }

        const jsonData = await response.json();
        // Apply similar logic as above to extract the array
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
        console.error(`Exception fetching/parsing temperature data for sensor ${sensorId}:`, error);
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



    // Function to get a cookie by name
    getCookie(): any {
        const value = `; ${document.cookie}`;
        const parts = value.split(`; session_id=`);
        if (parts.length === 2) {
            return parts?.pop()?.split(';')?.shift();
        }
        return false
  }
}
