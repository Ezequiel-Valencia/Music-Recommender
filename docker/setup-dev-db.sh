
#!/bin/bash

# Ask the user if they are running Docker Compose
read -p "Are you using Docker Compose? (yes/no): " response

# Convert the response to lowercase for case-insensitive comparison
response=$(echo "$response" | tr '[:upper:]' '[:lower:]')

if [[ "$response" == "yes" ]]; then
    PGPASSWORD=passwd psql -h localhost -p 5432 -U postgres -d postgres -f setup.sql
else
    echo "Proceeding with managing a PostgreSQL container..."
fi

# Check if a container named thDB already exists
container_exists=$(docker ps -a --filter "name=^thDB$" --format '{{.Names}}')

if [[ "$container_exists" == "thDB" ]]; then
    echo "Container 'thDB' already exists. Deleting it..."
    docker rm -f thDB
fi

# Create a new PostgreSQL container named thDB
echo "Creating a new PostgreSQL container named 'thDB'..."
docker run --name thDB -e POSTGRES_PASSWORD=passwd -p 127.0.0.1:5432:5432 postgres

if [[ $? -eq 0 ]]; then
    echo "PostgreSQL container 'thDB' has been created successfully." 
    else
    echo "Failed to create the PostgreSQL container. Please check Docker and try again."
fi


csrf: MTczNTQ5NTcxOHxSd3dBUkU5dVpTMVdSRWczV1RkdGJDMTJXREF0Ym1JeVIxZHdjalF4U1ZKS1FraG5OVzFmU0cxS1dDMTROM2hZYjFaelVUbDJiVEpFYmtSMGJ6Qk1OM2xsVVRFMlpHTTl8RE600GNhO32kS99DMKl55yb02kPSj03dexUCwvvAnGw=


