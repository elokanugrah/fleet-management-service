CREATE TABLE IF NOT EXISTS vehicle_locations (
    id          BIGSERIAL           PRIMARY KEY,
    vehicle_id  VARCHAR(50)         NOT NULL,
    latitude    DOUBLE PRECISION    NOT NULL,
    longitude   DOUBLE PRECISION    NOT NULL,
    timestamp   BIGINT              NOT NULL,
    created_at  TIMESTAMPTZ         NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_vehicle_locations_vehicle_id
    ON vehicle_locations(vehicle_id);

CREATE INDEX IF NOT EXISTS idx_vehicle_locations_timestamp
    ON vehicle_locations(timestamp);

CREATE INDEX IF NOT EXISTS idx_vehicle_locations_vehicle_timestamp
    ON vehicle_locations(vehicle_id, timestamp);
