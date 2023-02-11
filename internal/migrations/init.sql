CREATE DATABASE dutybot;
CREATE USER dutybot_rw WITH ENCRYPTED PASSWORD '<your password>';
-- connect to dutybot db
\c dutybot
GRANT ALL ON SCHEMA public TO dutybot_rw;
GRANT ALL PRIVILEGES ON DATABASE dutybot TO dutybot_rw;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
