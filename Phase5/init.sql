CREATE TABLE IF NOT EXISTS successful_transactions (
    transaction_id INT PRIMARY KEY,
    amount INT NOT NULL,
    status VARCHAR(20) NOT NULL
);