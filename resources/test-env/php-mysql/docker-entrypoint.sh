#!/bin/bash
set -euo pipefail

# If MariaDB datadir is empty, initialize
if [ ! -d /var/lib/mysql/mysql ]; then
  echo "Initializing MariaDB data directory..."
  chown -R mysql:mysql /var/lib/mysql || true
  mysqld --initialize-insecure --user=mysql --datadir=/var/lib/mysql
fi

# Start mariadb in background (use mysqld_safe if available)
echo "Starting MariaDB..."
if command -v mysqld_safe >/dev/null 2>&1; then
  mysqld_safe --datadir=/var/lib/mysql --bind-address=0.0.0.0 &
else
  mysqld --user=mysql --datadir=/var/lib/mysql --bind-address=0.0.0.0 &
fi

# Wait for database to be ready
echo "Waiting for MariaDB to be available..."
max_retries=60
until mysqladmin ping --silent; do
  sleep 1
  max_retries=$((max_retries-1))
  if [ "$max_retries" -le 0 ]; then
    echo "MariaDB did not become available in time"
    exit 1
  fi
done

# Secure root and create admin user if needed
echo "Configuring database users..."
mysql -u root <<-EOSQL
ALTER USER 'root'@'localhost' IDENTIFIED BY 'root';
CREATE USER IF NOT EXISTS 'admin'@'%' IDENTIFIED BY '1234';
GRANT ALL PRIVILEGES ON *.* TO 'admin'@'%' WITH GRANT OPTION;
FLUSH PRIVILEGES;
EOSQL

# Skipping automatic execution of init SQL files to keep environment clean.
# Users can mount SQL files to `/docker-entrypoint-initdb.d` or run imports manually.

# Finally, exec the container's main process (apache)
exec "$@"
