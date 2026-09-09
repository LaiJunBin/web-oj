
-- remove all database data "testdb"
DROP DATABASE IF EXISTS testdb;

-- create database "testdb" if it does not exist
CREATE DATABASE testdb;

-- Person table:
-- +----------+----------+-----------+
-- | personId | lastName | firstName |
-- +----------+----------+-----------+
-- | 1        | White    | Emma      |
-- | 2        | Black    | Noah      |
-- +----------+----------+-----------+
-- Address table:
-- +-----------+----------+---------------+------------+
-- | addressId | personId | city          | state      |
-- +-----------+----------+---------------+------------+
-- | 1         | 1        | Austin        | Texas      |
-- | 2         | NULL     | NULL          | NULL       |
-- +-----------+----------+---------------+------------+

use testdb;

-- create table "person"
CREATE TABLE person (
    personId INT PRIMARY KEY,
    lastName VARCHAR(255),
    firstName VARCHAR(255)
);
-- create table "address"
CREATE TABLE address (
    addressId INT PRIMARY KEY,
    personId INT,
    city VARCHAR(255),
    state VARCHAR(255)
);

-- insert data into "person"
INSERT INTO person (personId, lastName, firstName) VALUES
(1, 'White', 'Emma'),
(2, 'Black', 'Noah');
-- insert data into "address"
INSERT INTO address (addressId, personId, city, state) VALUES
(1, 1, 'Austin', 'Texas'),
(2, NULL, NULL, NULL);