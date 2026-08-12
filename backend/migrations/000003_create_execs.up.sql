CREATE TABLE
  IF NOT EXISTS execs (
    id INT PRIMARY KEY AUTO_INCREMENT,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    password_changed_at VARCHAR(255),
    user_created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    password_reset_code VARCHAR(255),
    password_code_expires TIMESTAMP,
    inactive_status BOOLEAN NOT NULL DEFAULT FALSE,
    role VARCHAR(50) NOT NULL,
    INDEX idx_email (email),
    INDEX idx_username (username)
  ) ENGINE = InnoDB;