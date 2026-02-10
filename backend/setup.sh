#!/bin/bash

# INTERNA Setup Script
# Run this script to set up your development environment

echo "========================================"
echo "  INTERNA - Internship Management System"
echo "  Setup Script"
echo "========================================"
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Step 1: Check Go installation
echo -e "${YELLOW}[1/7] Checking Go installation...${NC}"
if ! command -v go &> /dev/null; then
    echo "Go is not installed. Please install Go 1.21 or higher."
    exit 1
fi
echo -e "${GREEN}✓ Go is installed: $(go version)${NC}"
echo ""

# Step 2: Check MySQL installation
echo -e "${YELLOW}[2/7] Checking MySQL installation...${NC}"
if ! command -v mysql &> /dev/null; then
    echo "MySQL is not installed. Please install MySQL 8.0 or higher."
    exit 1
fi
echo -e "${GREEN}✓ MySQL is installed${NC}"
echo ""

# Step 3: Install Go dependencies
echo -e "${YELLOW}[3/7] Installing Go dependencies...${NC}"
go mod download
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Dependencies installed${NC}"
else
    echo "Failed to install dependencies"
    exit 1
fi
echo ""

# Step 4: Create .env file
echo -e "${YELLOW}[4/7] Creating .env file...${NC}"
if [ ! -f .env ]; then
    cp .env.example .env
    echo -e "${GREEN}✓ .env file created${NC}"
    echo -e "${YELLOW}⚠ Please edit .env file and update database credentials${NC}"
else
    echo ".env file already exists, skipping..."
fi
echo ""

# Step 5: Create upload directories
echo -e "${YELLOW}[5/7] Creating upload directories...${NC}"
mkdir -p uploads/tasks
mkdir -p uploads/leaves
echo -e "${GREEN}✓ Upload directories created${NC}"
echo ""

# Step 6: Setup database
echo -e "${YELLOW}[6/7] Setting up database...${NC}"
echo "Please enter your MySQL root password:"
read -s MYSQL_PASSWORD

echo "Creating database..."
mysql -u root -p"$MYSQL_PASSWORD" -e "CREATE DATABASE IF NOT EXISTS interna_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

if [ $? -eq 0 ]; then
    echo "Importing schema..."
    mysql -u root -p"$MYSQL_PASSWORD" interna_db < database/schema.sql
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Database setup complete${NC}"
    else
        echo "Failed to import schema"
        exit 1
    fi
else
    echo "Failed to create database"
    exit 1
fi
echo ""

# Step 7: Done
echo -e "${YELLOW}[7/7] Setup complete!${NC}"
echo ""
echo "========================================"
echo -e "${GREEN}✓ Setup completed successfully!${NC}"
echo "========================================"
echo ""
echo "Next steps:"
echo "1. Edit .env file with your database credentials"
echo "2. Run: go run cmd/server/main.go"
echo "3. Access API at: http://localhost:8080"
echo ""
echo "API Health Check: http://localhost:8080/api/health"
echo ""
echo "Happy coding! 🚀"
