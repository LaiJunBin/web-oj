Feature: Create User API
  As a user
  I want to create a user with a roles int
  So that I can create a user with a roles int if I have permission

  Background:
    Given I have users:
      | id  | name               | username | password | roles              |
      | 11  | AdminUser          | user01   | 123456   | admin              |
      | 12  | APIReader          | user02   | 123456   | api_reader         |
      | 13  | Contestant         | user03   | 123456   | contestant         |
      | 14  | AdminAndContestant | user04   | 123456   | admin,contestant   |

  Scenario: Successful create user with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/users" with the following data:
      """
      {
        "name": "NewUser",
        "username": "newuser",
        "password": "123456",
        "roles": [1, 2]
      }
      """
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": 15
      }
      """
    And the user should be created with the following data:
      | name    | username | roles              |
      | NewUser | newuser  | admin,contestant   |
    And can login with the new user:
      """
      {
        "username": "newuser",
        "password": "123456"
      }
      """
  
  Scenario: Successful create user with "AdminAndContestant" user
    Given I am logged in as "user04"
    When I send a "POST" request to "/api/v1/users" with the following data:
      """
      {
        "name": "NewUser",
        "username": "newuser",
        "password": "123456",
        "roles": []
      }
      """
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": 15
      }
      """
    And the user should be created with the following data:
      | name    | username | roles              |
      | NewUser | newuser  |                    |
    And can login with the new user:
      """
      {
        "username": "newuser",
        "password": "123456"
      }
      """

  Scenario: Failed create user with duplicated username
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/users" with the following data:
      """
      {
        "name": "NewUser",
        "username": "user01",
        "password": "123456",
        "roles": [1, 2]
      }
      """
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "username already exists",
        "details": null
      }
      """

  Scenario: Failed create user with a reserved-marker username
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/users" with the following data:
      """
      {
        "name": "Evil",
        "username": "evil__deleted_1",
        "password": "123456",
        "roles": [1, 2]
      }
      """
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "username contains a reserved word",
        "details": null
      }
      """

  Scenario: Failed create user with invalid roles
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/users" with the following data:
      """
      {
        "name": "NewUser",
        "username": "newuser",
        "password": "123456",
        "roles": [-1]
      }
      """
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "invalid roles",
        "details": null
      }
      """

  Scenario: Unauthorized create user
    When I send a "POST" request to "/api/v1/users" with the following data:
      """
      {
        "name": "NewUser",
        "username": "newuser",
        "password": "123456",
        "roles": [1, 2]
      }
      """
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """
    
  Scenario: Forbidden create user with "APIReader" user
    Given I am logged in as "user02"
    When I send a "POST" request to "/api/v1/users" with the following data:
      """
      {
        "name": "NewUser",
        "username": "newuser",
        "password": "123456",
        "roles": [1, 2]
      }
      """
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: Forbidden create user with "Contestant" user
    Given I am logged in as "user03"
    When I send a "POST" request to "/api/v1/users" with the following data:
      """
      {
        "name": "NewUser",
        "username": "newuser",
        "password": "123456",
        "roles": [1, 2]
      }
      """
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """



