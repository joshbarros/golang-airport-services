-- Drop trigger
DROP TRIGGER IF EXISTS update_flights_updated_at ON flights;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_flights_active;
DROP INDEX IF EXISTS idx_flights_flight_number_date;
DROP INDEX IF EXISTS idx_flights_origin_destination;
DROP INDEX IF EXISTS idx_flights_departure_time;
DROP INDEX IF EXISTS idx_flights_destination;
DROP INDEX IF EXISTS idx_flights_origin;
DROP INDEX IF EXISTS idx_flights_status;
DROP INDEX IF EXISTS idx_flights_flight_number;

-- Drop table
DROP TABLE IF EXISTS flights;
