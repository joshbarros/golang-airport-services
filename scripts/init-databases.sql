-- Initialize multiple databases for different services

-- Flight Service Database
CREATE DATABASE flight_db;
\c flight_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Booking Service Database
\c postgres;
CREATE DATABASE booking_db;
\c booking_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Passenger Service Database
\c postgres;
CREATE DATABASE passenger_db;
\c passenger_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Baggage Service Database
\c postgres;
CREATE DATABASE baggage_db;
\c baggage_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Payment Service Database
\c postgres;
CREATE DATABASE payment_db;
\c payment_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Loyalty Service Database
\c postgres;
CREATE DATABASE loyalty_db;
\c loyalty_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Security Service Database
\c postgres;
CREATE DATABASE security_db;
\c security_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Analytics Service Database
\c postgres;
CREATE DATABASE analytics_db;
\c analytics_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "timescaledb" CASCADE;

-- Auth Service Database
\c postgres;
CREATE DATABASE auth_db;
\c auth_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Grant privileges to airport user
\c postgres;
GRANT ALL PRIVILEGES ON DATABASE flight_db TO airport;
GRANT ALL PRIVILEGES ON DATABASE booking_db TO airport;
GRANT ALL PRIVILEGES ON DATABASE passenger_db TO airport;
GRANT ALL PRIVILEGES ON DATABASE baggage_db TO airport;
GRANT ALL PRIVILEGES ON DATABASE payment_db TO airport;
GRANT ALL PRIVILEGES ON DATABASE loyalty_db TO airport;
GRANT ALL PRIVILEGES ON DATABASE security_db TO airport;
GRANT ALL PRIVILEGES ON DATABASE analytics_db TO airport;
GRANT ALL PRIVILEGES ON DATABASE auth_db TO airport;
