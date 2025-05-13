# Install
Create a .env file in the root directory. This file should contain the following variables:  
```env
LOG_VERBOSE=false
DEVICE_IP=MY_DEVICE_IP
REFRESH_SECONDS=15
PORT=9093
PROM_NAMESPACE=epever
```

Download the latest release of the exporter and mark it as executable:  
```bash
# Download the latest release of the exporter
ARCH=$(uname -m); OS=$(uname -s | tr '[:upper:]' '[:lower:]'); \
case $ARCH in x86_64) ARCH=amd64;; i386|i686) ARCH=386;; aarch64) ARCH=arm64;; esac; \
FILENAME="epever-prom-export-${OS}-${ARCH}"; \
wget -q "https://github.com/BolverBlitz/EPEver2MQTT-Exporter/releases/latest/download/${FILENAME}" -O ${FILENAME}
wget -q "https://raw.githubusercontent.com/BolverBlitz/EPEver2MQTT-Exporter/refs/heads/main/config.json" -O config.json

# Mark the file as executable
chmod +x epever-prom-export-*
```