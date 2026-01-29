#!/bin/bash
set -e

echo "🔧 Setting up PostgreSQL for pkt..."

# Create user if not exists
echo "Creating user 'pkt_user'..."
sudo -u postgres psql -c "DO \$\$ BEGIN IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'pkt_user') THEN CREATE ROLE pkt_user WITH LOGIN PASSWORD 'yourpassword'; END IF; END \$\$;"

# Create database if not exists
echo "Creating database 'pkt_db'..."
sudo -u postgres psql -c "SELECT 'CREATE DATABASE pkt_db OWNER pkt_user' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'pkt_db')\gexec"

echo "✅ Database setup complete."
echo ""
echo "🚀 Running pkt start..."

# Run pkt start with the new credentials
export PKT_DB_USER=pkt_user
export PKT_DB_PASSWORD=yourpassword
export PKT_DB_NAME=pkt_db

./bin/pkt start
