SELECT 'CREATE DATABASE kratos OWNER manoreck'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'kratos')\gexec
