-- Create flights table
CREATE TABLE IF NOT EXISTS flights (
    id VARCHAR(36) PRIMARY KEY,
    flight_number VARCHAR(10) NOT NULL,
    origin VARCHAR(3) NOT NULL,
    destination VARCHAR(3) NOT NULL,
    departure_time TIMESTAMP WITH TIME ZONE NOT NULL,
    arrival_time TIMESTAMP WITH TIME ZONE NOT NULL,
    actual_departure_time TIMESTAMP WITH TIME ZONE,
    actual_arrival_time TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) NOT NULL,
    aircraft_type VARCHAR(50) NOT NULL,
    gate VARCHAR(10) NOT NULL,
    terminal VARCHAR(10),
    delay_reason TEXT,
    cancellation_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for common query patterns
CREATE INDEX idx_flights_flight_number ON flights(flight_number);
CREATE INDEX idx_flights_status ON flights(status);
CREATE INDEX idx_flights_origin ON flights(origin);
CREATE INDEX idx_flights_destination ON flights(destination);
CREATE INDEX idx_flights_departure_time ON flights(departure_time);
CREATE INDEX idx_flights_origin_destination ON flights(origin, destination);
CREATE INDEX idx_flights_flight_number_date ON flights(flight_number, DATE(departure_time));

-- Create composite index for active flights (frequently queried)
CREATE INDEX idx_flights_active ON flights(status, departure_time)
    WHERE status NOT IN ('arrived', 'cancelled');

-- Add constraint to ensure arrival is after departure
ALTER TABLE flights ADD CONSTRAINT chk_arrival_after_departure
    CHECK (arrival_time > departure_time);

-- Add constraint for valid status values
ALTER TABLE flights ADD CONSTRAINT chk_valid_status
    CHECK (status IN ('scheduled', 'boarding', 'departed', 'in_flight', 'landed', 'arrived', 'delayed', 'cancelled'));

-- Add constraint for valid airport codes (IATA 3-letter codes)
ALTER TABLE flights ADD CONSTRAINT chk_valid_origin
    CHECK (LENGTH(origin) = 3 AND origin = UPPER(origin));

ALTER TABLE flights ADD CONSTRAINT chk_valid_destination
    CHECK (LENGTH(destination) = 3 AND destination = UPPER(destination));

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to auto-update updated_at
CREATE TRIGGER update_flights_updated_at
    BEFORE UPDATE ON flights
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE flights IS 'Stores flight information for the airport services system';
COMMENT ON COLUMN flights.id IS 'Unique identifier for the flight (UUID)';
COMMENT ON COLUMN flights.flight_number IS 'Flight number (e.g., AA123, DL1234)';
COMMENT ON COLUMN flights.origin IS 'Origin airport IATA code (e.g., JFK)';
COMMENT ON COLUMN flights.destination IS 'Destination airport IATA code (e.g., LAX)';
COMMENT ON COLUMN flights.status IS 'Current flight status (scheduled, boarding, departed, etc.)';
COMMENT ON COLUMN flights.actual_departure_time IS 'Actual departure time (null if not yet departed)';
COMMENT ON COLUMN flights.actual_arrival_time IS 'Actual arrival time (null if not yet arrived)';
