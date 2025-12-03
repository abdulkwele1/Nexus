import { defineStore } from 'pinia'
import { apiFetch } from '@/services/api';

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
    isAdmin: boolean;
    userRole: string;

    constructor(baseURL: any) {
        this.baseURL = baseURL;
        this.userName = '';
        this.loggedIn = false;
        this.isAdmin = false;
        this.userRole = 'user';
    }

    async login(userName: any, password: any): Promise<any> {
        const loginAPIUrl = `${VITE_NEXUS_API_URL}/login`
        const data = {
            username: userName,
            password: password,
        };

        return apiFetch(loginAPIUrl, {
            headers: {
                'Content-Type': 'application/json',
            },
            method: 'POST',
            body: JSON.stringify(data),
            skipAuthCheck: true, // Skip auth check for login endpoint
        });
    }

    async logout() {
        const logoutAPIUrl = `${VITE_NEXUS_API_URL}/logout`;
        return apiFetch(logoutAPIUrl, {
            method: 'POST',
            skipAuthCheck: true, // Skip auth check for logout endpoint
        });
    }

    async refreshSession() {
        const refreshAPIUrl = `${VITE_NEXUS_API_URL}/refresh-session`;
        return apiFetch(refreshAPIUrl, {
            method: 'POST',
        });
    }

    async getPanelYieldData(panelId: number, startDate: string, endDate: string,): Promise<any> {
        const url = `${VITE_NEXUS_API_URL}/panels/${panelId}/yield_data`
        return apiFetch(url, {
            credentials: 'include',
            headers: {
                'Content-Type': 'application/json',
            },
            method: 'GET',
        });
    }

    async getPanelConsumptionData(panelId: number, startDate: string, endDate: string,): Promise<any> {
        const url = `${VITE_NEXUS_API_URL}/panels/${panelId}/consumption_data`
        return apiFetch(url, {
            credentials: 'include',
            headers: {
                'Content-Type': 'application/json',
            },
            method: 'GET',
        });
    }

    async setPanelYieldData(panelId: number, startDate: string, endDate: string, yieldData: object): Promise<any> {
        const url = `${VITE_NEXUS_API_URL}/panels/${panelId}/yield_data`
        const payload = {
            start_date: startDate,
            end_date: endDate,
            yield_data: yieldData,
        }
        
        return apiFetch(url, {
            credentials: 'include',
            headers: {
                'Content-Type': 'application/json',
            },
            method: 'POST',
            body: JSON.stringify(payload),
        });
    }

    async setPanelConsumptionData(panelId: number, startDate: string, endDate: string, consumptionData: object): Promise<any> {
        const url = `${VITE_NEXUS_API_URL}/panels/${panelId}/consumption_data`
        const payload = {
            start_date: startDate,
            end_date: endDate,
            consumption_data: consumptionData,
        }
      
        return apiFetch(url, {
            credentials: 'include',
            headers: {
                'Content-Type': 'application/json',
            },
            method: 'POST',
            body: JSON.stringify(payload),
        });
    }

    async setSensorTemperatureData(sensorId: string, date: string, soilTemperature: number): Promise<any> {
        const url = `${VITE_NEXUS_API_URL}/sensors/${sensorId}/temperature_data`
        return apiFetch(url, {
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
    }

    async setSensorMoistureData(sensorId: string, date: string, soilMoisture: number): Promise<any> {
        const url = `${VITE_NEXUS_API_URL}/sensors/${sensorId}/moisture_data`
        return apiFetch(url, {
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
    }

    async getSensorTemperatureData(sensorId: string) {
        try {
            const response = await apiFetch(`${this.baseURL}/sensors/${sensorId}/temperature_data`, {
                credentials: 'include',
            });
            
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
            const response = await apiFetch(`${this.baseURL}/sensors/${sensorId}/moisture_data`, {
                credentials: 'include',
            });
            
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

    async getDroneImages(startDate?: Date, endDate?: Date): Promise<any> {
        try {
            let url = `${this.baseURL}/drone_images`;
            
            // Add date range parameters if provided
            if (startDate && endDate) {
                url += `?start_date=${startDate.toISOString()}&end_date=${endDate.toISOString()}`;
            }

            const response = await apiFetch(url, {
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json',
                },
                method: 'GET',
            });

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
        const response = await apiFetch(`${this.baseURL}/drone_images/${imageId}/content`, {
            credentials: 'include',
        });
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
                mime_type: imageFile.type
            }));

            const response = await apiFetch(`${this.baseURL}/drone_images`, {
                credentials: 'include',
                method: 'POST',
                body: formData
            });

            const result = await response.json();
            
            // Return the result with a temporary URL for immediate display
            if (result.uploaded_images && result.uploaded_images.length > 0) {
                const image = result.uploaded_images[0];
                return {
                    ...image,
                    url: URL.createObjectURL(imageFile)
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
            const response = await apiFetch(`${this.baseURL}/drone_images/${imageId}`, {
                credentials: 'include',
                method: 'DELETE',
                headers: {
                    'Content-Type': 'application/json',
                },
            });

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
            console.log('[NexusStore] Fetching all sensors from:', `${this.baseURL}/sensors`);
            const response = await apiFetch(`${this.baseURL}/sensors`, {
                credentials: 'include',
            });
            const data = await response.json();
            console.log('[NexusStore] Got sensors response:', data);
            // If data is an array, wrap it in the expected format
            if (Array.isArray(data)) {
                return { sensors: data };
            }
            // If data already has sensors property, return as is
            if (data && Array.isArray(data.sensors)) {
                return data;
            }
            console.warn('[NexusStore] Unexpected sensors response format:', data);
            return { sensors: [] };
        } catch (error) {
            console.error('[NexusStore] Error fetching all sensors:', error);
            return { sensors: [] };
        }
    }

    async addSensor(eui: string, name?: string, location?: string) {
        try {
            console.log('[NexusStore] Adding new sensor with EUI:', eui);
            const response = await apiFetch(`${this.baseURL}/sensors`, {
                method: 'POST',
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    eui: eui,
                    name: name || `Sensor ${eui}`,
                    location: location || 'Unknown Location'
                }),
            });
            
            if (response.ok) {
                const data = await response.json();
                console.log('[NexusStore] Sensor added successfully:', data);
                // Trigger a custom event to notify components that sensors have been updated
                window.dispatchEvent(new CustomEvent('sensorsUpdated'));
                return { success: true, message: data.message };
            } else {
                const errorData = await response.json();
                console.error('[NexusStore] Failed to add sensor:', errorData);
                return { success: false, error: errorData.error || 'Failed to add sensor' };
            }
        } catch (error) {
            console.error('[NexusStore] Error adding sensor:', error);
            return { success: false, error: 'Network error occurred' };
        }
    }

    async deleteSensor(sensorId: string) {
        try {
            console.log('[NexusStore] Deleting sensor with ID:', sensorId);
            const response = await apiFetch(`${this.baseURL}/sensors/${sensorId}`, {
                method: 'DELETE',
                credentials: 'include',
            });
            
            if (response.ok) {
                const data = await response.json();
                console.log('[NexusStore] Sensor deleted successfully:', data);
                // Trigger a custom event to notify components that sensors have been updated
                window.dispatchEvent(new CustomEvent('sensorsUpdated'));
                return { success: true, message: data.message };
            } else {
                const errorData = await response.json();
                console.error('[NexusStore] Failed to delete sensor:', errorData);
                return { success: false, error: errorData.error || 'Failed to delete sensor' };
            }
        } catch (error) {
            console.error('[NexusStore] Error deleting sensor:', error);
            return { success: false, error: 'Network error occurred' };
        }
    }

    async getSensorBatteryData(sensorId: string, startDate?: string, endDate?: string) {
        try {
            let url = `${this.baseURL}/sensors/${sensorId}/battery_data`;
            
            // Add date range parameters if provided
            if (startDate && endDate) {
                url += `?start_date=${startDate}&end_date=${endDate}`;
            }
            
            console.log(`[NexusStore] Fetching battery data for sensor ${sensorId} from:`, url);
            
            const response = await apiFetch(url, {
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json',
                },
                method: 'GET',
            });
            
            const jsonData = await response.json();
            console.log(`[NexusStore] Got battery data for sensor ${sensorId}:`, jsonData);
            
            if (jsonData && Array.isArray(jsonData.battery_level_data)) {
                return {
                    battery_level_data: jsonData.battery_level_data
                };
            }
            
            console.warn(`[NexusStore] Battery data for sensor ${sensorId} received in unexpected format:`, jsonData);
            return {
                battery_level_data: []
            };
        } catch (error) {
            console.error('[NexusStore] Error fetching sensor battery data:', error);
            return {
                battery_level_data: []
            };
        }
    }

    async setSensorBatteryData(sensorId: string, batteryData: Array<{
        date: string;
        battery_level: number;
    }>) {
        const url = `${this.baseURL}/sensors/${sensorId}/battery_data`;
        return apiFetch(url, {
            credentials: 'include',
            headers: {
                'Content-Type': 'application/json',
            },
            method: 'POST',
            body: JSON.stringify({
                battery_level_data: batteryData
            }),
        });
    }

    // Admin methods
    async getAllUsers() {
        try {
            console.log('[NexusStore] Fetching all users from:', `${this.baseURL}/admin/users`);
            const response = await apiFetch(`${this.baseURL}/admin/users`, {
                credentials: 'include',
            });
            const data = await response.json();
            console.log('[NexusStore] Got users response:', data);
            return data;
        } catch (error) {
            console.error('[NexusStore] Error fetching all users:', error);
            throw error;
        }
    }

    async updateUserRole(username: string, role: string) {
        try {
            console.log('[NexusStore] Updating user role:', username, 'to', role);
            const response = await apiFetch(`${this.baseURL}/admin/users/${username}`, {
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json',
                },
                method: 'PATCH',
                body: JSON.stringify({ role }),
            });
            const data = await response.json();
            console.log('[NexusStore] User role update response:', data);
            return data;
        } catch (error) {
            console.error('[NexusStore] Error updating user role:', error);
            throw error;
        }
    }

    async removeAdminPermissions(username: string) {
        try {
            console.log('[NexusStore] Removing admin permissions for:', username);
            const response = await apiFetch(`${this.baseURL}/admin/users/${username}/remove-admin`, {
                credentials: 'include',
                method: 'DELETE',
            });
            const data = await response.json();
            console.log('[NexusStore] Remove admin response:', data);
            return data;
        } catch (error) {
            console.error('[NexusStore] Error removing admin permissions:', error);
            throw error;
        }
    }

    async createUser(username: string, password: string, role: string = 'user') {
        try {
            console.log('[NexusStore] Creating new user:', username, 'with role:', role);
            const response = await apiFetch(`${this.baseURL}/admin/users`, {
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json',
                },
                method: 'POST',
                body: JSON.stringify({ username, password, role }),
            });
            const data = await response.json();
            console.log('[NexusStore] Create user response:', data);
            return data;
        } catch (error) {
            console.error('[NexusStore] Error creating user:', error);
            throw error;
        }
    }

    async deleteUser(username: string) {
        try {
            console.log('[NexusStore] Deleting user:', username);
            const response = await apiFetch(`${this.baseURL}/admin/users/${username}`, {
                credentials: 'include',
                method: 'DELETE',
            });
            const data = await response.json();
            console.log('[NexusStore] Delete user response:', data);
            return data;
        } catch (error) {
            console.error('[NexusStore] Error deleting user:', error);
            throw error;
        }
    }

    async checkUsernameAvailable(username: string): Promise<{ available: boolean; error?: string }> {
        try {
            console.log('[NexusStore] Checking username availability:', username);
            const response = await apiFetch(`${this.baseURL}/admin/users/check/${encodeURIComponent(username)}`, {
                credentials: 'include',
                method: 'GET',
            });
            const data = await response.json();
            
            // If status is OK and message is "Username is available", it's available
            if (response.ok && data.message === 'Username is available') {
                return { available: true };
            }
            
            // If status is OK but has an error field, username is taken
            if (response.ok && data.error) {
                return { available: false, error: data.error };
            }
            
            // Otherwise, username is taken or there's an error
            return { available: false, error: data.error || 'Username is not available' };
        } catch (error: any) {
            console.error('[NexusStore] Error checking username:', error);
            // Try to parse error response
            try {
                const errorData = await error.json?.() || {};
                if (errorData.error) {
                    return { available: false, error: errorData.error };
                }
            } catch (e) {
                // Ignore JSON parse errors
            }
            // For network errors, assume available (don't block user)
            return { available: true };
        }
    }

    async checkAdminStatus() {
        try {
            // Try to access admin endpoint to check if user is admin
            const response = await apiFetch(`${this.baseURL}/admin/users`, {
                credentials: 'include',
            });
            return response.ok;
        } catch (error) {
            console.log('[NexusStore] User is not an admin:', error);
            return false;
        }
    }

    async getUserSettings() {
        try {
            console.log('[NexusStore] Fetching user settings from:', `${this.baseURL}/settings`);
            const response = await apiFetch(`${this.baseURL}/settings`, {
                credentials: 'include',
            });
            const data = await response.json();
            console.log('[NexusStore] Got user settings response:', data);
            
            // Update user role and admin status
            if (data.role) {
                this.userRole = data.role;
                this.isAdmin = data.role === 'admin' || data.role === 'root_admin';
            }
            
            return data;
        } catch (error) {
            console.error('[NexusStore] Error fetching user settings:', error);
            throw error;
        }
    }
}