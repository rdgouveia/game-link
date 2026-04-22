-- +goose Up
CREATE TABLE user_availability (
    user_id     UUID     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    day_of_week SMALLINT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
    start_time  TIME     NOT NULL,
    end_time    TIME     NOT NULL,
    PRIMARY KEY (user_id, day_of_week, start_time),
    CHECK (end_time > start_time)
);

-- +goose Down
DROP TABLE user_availability;
