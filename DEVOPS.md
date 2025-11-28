# Developer Operations

## Goals

1. Hosting - When user's go to <https://nexus.eternalrelayrace.com>, they are able to access the latest version of the app
2. Deployment - Developers should be able to deploy new versions of the app after code is reviewed and merged into the main branch

### Hosting Steps

1.) Acquire domain name (e.g. from <https://www.hover.com>), this is the "address of your business" - e.g. the Nexus website, on "the internet". When people want to "see you", they go to your address - <https://nexus.eternalrelayrace.com>
2.) Acquire a public IP (Internet Protocol) address - this is the address of the computer (host) that your website runs on. Example IPv4 address - 35.94.111.25
3.) Add DNS record to your domain name to map it to the ip address of the host your website and api runs on
4.) Turn on your computer / launch a cloud instance - for example AWS EC2
5.) Associate your IP address with your computer

Now whenever a user goes to <https://nexus.eternalrelayrace.com>, their browser will look up the DNS entry for that website, receive an answer of 35.94.111.25, and send the users traffic to the computer with that address on the internet.

### First Time Host Setup

This section covers the complete process of setting up a new EC2 instance from scratch, including instance creation, security group configuration, tool installation, and initial deployment.

#### 1. Create EC2 Instance

1. **Navigate to EC2 Console**
   - Log into AWS Console (<https://console.aws.amazon.com>)
   - Select your region (e.g., `us-west-2` - Oregon)
   - Navigate to EC2 service

2. **Launch Instance**
   - Click "Launch Instance" button
   - Configure instance details:
     - **Name**: `nexus-demo-server-v2` (or your preferred name)
     - **AMI**: Select "Ubuntu Server" (latest LTS version, e.g., Ubuntu 24.04 LTS)
     - **Instance type**: `t3.micro` (suitable for development/testing)
     - **Key pair**: Create new or select existing key pair for SSH access
       - Download the `.pem` file and store securely
       - Set permissions: `chmod 400 your-key.pem`
     - **Network settings**:
       - Select or create a VPC
       - Select a public subnet
       - Enable "Auto-assign public IP"
     - **Storage**: Default 8GB gp3 volume is sufficient
   - Click "Launch Instance"

3. **Note Instance Details**
   - After launch, note the following from the instance details page:
     - **Instance ID**: e.g., `i-00ad9a252e5b94681`
     - **Public IPv4 address**: e.g., `34.217.174.48`
     - **Private IPv4 address**: e.g., `172.31.31.184`
     - **VPC ID**: e.g., `vpc-0738956fb4e6605b2`
     - **Subnet ID**: e.g., `subnet-0da9ce6da3f80c92e`

#### 2. Configure Security Group

The security group controls inbound and outbound traffic to your EC2 instance. Configure the following inbound rules:

1. **Navigate to Security Groups**
   - In EC2 Console, go to "Security Groups" in the left sidebar
   - Find the security group associated with your instance (or create a new one)
   - Click "Edit inbound rules"

2. **Add Required Inbound Rules**
   Click "Add rule" for each of the following:

   | Type | Protocol | Port Range | Source | Description |
   |------|----------|------------|--------|-------------|
   | SSH | TCP | 22 | My IP (or 0.0.0.0/0 for development) | SSH access for administration |
   | HTTP | TCP | 80 | 0.0.0.0/0 | HTTP web traffic |
   | HTTPS | TCP | 443 | 0.0.0.0/0 | HTTPS web traffic |
   | Custom TCP | TCP | 8080 | 0.0.0.0/0 | Nexus API service |
   | Custom TCP | TCP | 5173 | 0.0.0.0/0 | Nexus UI development server |

   **Note**: For production, restrict SSH (port 22) to specific IP addresses instead of 0.0.0.0/0

3. **Save Rules**
   - Click "Save rules" to apply the configuration

#### 3. Connect to EC2 Instance

1. **SSH into the instance**

   ```bash
   ssh -i your-key.pem ubuntu@<PUBLIC_IP_ADDRESS>
   ```

   Example:

   ```bash
   ssh -i nexus-demo-server.pem ubuntu@34.217.174.48
   ```

   - Type `yes` when prompted to trust the server
   - You should now be logged into the Ubuntu instance

#### 4. Install Required Tools

Run the following commands on the EC2 instance to install necessary tools:

1. **Update package manager**

   ```bash
   sudo apt update
   ```

2. **Install Docker**

   ```bash
   sudo snap install docker
   ```

   Or using apt (alternative method):

   ```bash
   sudo apt-get update
   sudo apt-get install -y docker.io
   sudo systemctl start docker
   sudo systemctl enable docker
   sudo usermod -aG docker ubuntu
   ```

   Note: If using snap, you may need to add docker group permissions separately

3. **Install Docker Compose**

   ```bash
   sudo apt-get install -y docker-compose
   ```

   Or install latest version:

   ```bash
   sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
   sudo chmod +x /usr/local/bin/docker-compose
   ```

4. **Install Make**

   ```bash
   sudo apt install -y make
   ```

5. **Install PostgreSQL Client (psql)**

   ```bash
   sudo apt-get update
   sudo apt-get install -y postgresql-client
   ```

6. **Verify installations**

   ```bash
   docker --version
   docker-compose --version
   make --version
   psql --version
   ```

#### 5. Clone and Configure Repository

1. **Clone the Nexus repository**

   ```bash
   git clone https://github.com/abdulkwele1/Nexus.git
   cd Nexus/
   ```

2. **Create environment file**

   ```bash
   touch .env
   ```

3. **Configure environment variables**

   ```bash
   vim .env
   ```

   Or use `nano` if you prefer:

   ```bash
   nano .env
   ```

4. **Set the API URL**
   In the `.env` file, set `VITE_NEXUS_API_URL` to match your EC2 instance's public IP:

   ```bash
   VITE_NEXUS_API_URL=http://34.217.174.48:8080
   ```

   Replace `34.217.174.48` with your actual public IP address.

   The `.env` file should include all necessary configuration. See the repository's `.env.example` or existing configuration for other required values.

#### 6. Initial Deployment

1. **Start all services**

   ```bash
   sudo make up
   ```

   This will:
   - Build Docker images for the API and UI
   - Start PostgreSQL database
   - Start Nexus API service
   - Start Nexus UI service

2. **Seed the database**

   ```bash
   sudo make seed-database
   ```

   This populates the database with initial data for development/testing.

3. **Verify services are running**

   ```bash
   sudo docker ps
   ```

   You should see containers for:
   - `nexus-nexus-db-1` (PostgreSQL)
   - `nexus-nexus-api-1` (Nexus API)
   - `nexus-nexus-ui-1` (Nexus UI)
   - `nexus-docker-host-1` (Docker host helper)

4. **Check logs (optional)**

   ```bash
   sudo make logs
   ```

   Press `Ctrl+C` to exit logs view.

#### 7. Verify Deployment

1. **Test API endpoint**
   Open in browser or use curl:

   ```bash
   curl http://34.217.174.48:8080/health
   ```

   Replace with your public IP address.

2. **Test UI**
   Open in browser:

   ```text
   http://34.217.174.48:5173
   ```

3. **Verify database connection**

   ```bash
   sudo make debug-database
   ```

Your Nexus application should now be running and accessible via the public IP address!

### Deployment Steps

1.) Download the private ssh key to the server (ask Abdul or Levi)
2.) Set the required security permissions for the key `chmod 400 nexus-demo-server.pem`
3.) SSH (remote login) to the server `ssh -i nexus-demo-server.pem ubuntu@35.94.111.25` (type yes when it asks you if you want to trust this server)
4.) `cd Nexus` to go to the Nexus code repo
5.) Get the latest merged code `git pull -p`
6.) If you need to deploy updates to the api run `sudo make refresh`
7.) If you need to seed the database run `sudo make seed-database`
8.) If you need to deploy updates to the frontend run `sudo docker compose up -d  nexus-ui --build --force-recreate`

#### Operations

If you need to debug and view the logs run `sudo make logs`

If you need to debug the state of the database `sudo make debug-database`
