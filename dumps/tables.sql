CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  firstname VARCHAR NOT NULL,
  lastname VARCHAR NOT NULL,
  patronymic VARCHAR NOT NULL,
  email VARCHAR UNIQUE,
  phone VARCHAR UNIQUE NOT NULL,
  password VARCHAR NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE passports (
  id SERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  serial INT NOT NULL,
  number INT NOT NULL,
  issued_by VARCHAR NOT NULL,
  issued_date TIMESTAMP NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE card_states (
  id INT PRIMARY KEY UNIQUE,
  description TEXT NOT NULL
);

CREATE TABLE transaction_states (
  id INT PRIMARY KEY UNIQUE,
  description TEXT NOT NULL
);

CREATE TABLE cards (
  id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  user_id BIGINT NOT NULL,
  number VARCHAR NOT NULL,
  mask VARCHAR NOT NULL,
  cvc INT NOT NULL,
  balance NUMERIC DEFAULT 0,
  release_date TIMESTAMP DEFAULT NULL,
  until_date TIMESTAMP DEFAULT NULL,
  state_id INT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (state_id) REFERENCES card_states(id) ON DELETE CASCADE
);

CREATE TABLE transactions (
  id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  card_id UUID NOT NULL,
  summ NUMERIC DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  succeeded_at TIMESTAMP DEFAULT NULL,
  state_id INT NOT NULL,
  target_id UUID NOT NULL,
  FOREIGN KEY (card_id) REFERENCES cards(id) ON DELETE CASCADE,
  FOREIGN KEY (target_id) REFERENCES cards(id) ON DELETE CASCADE,
  FOREIGN KEY (state_id) REFERENCES transaction_states(id) ON DELETE CASCADE
);


INSERT INTO card_states (id, description) VALUES (0, 'Not Activated');
INSERT INTO card_states (id, description) VALUES (1, 'Activated');
INSERT INTO card_states (id, description) VALUES (2, 'Blocked');
INSERT INTO card_states (id, description) VALUES (3, 'Outdated');

INSERT INTO transaction_states (id, description) VALUES (0, 'Failed');
INSERT INTO transaction_states (id, description) VALUES (1, 'Success');
INSERT INTO transaction_states (id, description) VALUES (2, 'Not enough balance');
INSERT INTO transaction_states (id, description) VALUES (3, 'Refund');

INSERT INTO users (firstname, lastname, patronymic, email, phone, password) VALUES ('Joe', 'Brown', 'Jackman', 'joe@mail.com', '79876543210', '123456');
INSERT INTO passports (user_id, serial, number, issued_by, issued_date) VALUES(1, 1234, 567890, 'Mars, Mountain #2', CURRENT_TIMESTAMP);
INSERT INTO cards (user_id, number, mask, cvc, state_id) VALUES (1, '8765564332450001', '8765 **** **** 0001', 789, 0);
INSERT INTO cards (user_id, number, mask, cvc, state_id) VALUES (1, '8765564332450002', '8765 **** **** 0002', 256, 0);