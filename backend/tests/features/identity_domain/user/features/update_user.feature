Feature: Update User API
  As a user
  I want to update a user
  So that I can update user information if I have permission

  Background:
    Given I have users:
      | id  | name               | username | password | roles              |
      | 11  | AdminUser          | user01   | 123456   | admin              |
      | 12  | APIReader          | user02   | 123456   | api_reader         |
      | 13  | Contestant         | user03   | 123456   | contestant         |
      | 14  | AdminAndContestant | user04   | 123456   | admin,contestant   |

  Scenario: Successful update user with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "PATCH" request to "/api/v1/users/11" with the following data:
      """
      {
        "name": "new user01",
        "password": "12345678",
        "roles": [1, 2, 3]
      }
      """
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true
      }
      """
    And the user should be updated with the following data:
      | name       | username | roles                         |
      | new user01 | user01   | admin,contestant,api_reader   |
    And can login with the updated user:
      """
      {
        "username": "user01",
        "password": "12345678"
      }
      """
  
  Scenario: Successful update user with "AdminAndContestant" user
    Given I am logged in as "user04"
    When I send a "PATCH" request to "/api/v1/users/14" with the following data:
      """
      {
        "password": "12345678"
      }
      """
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true
      }
      """
    And the user should be updated with the following data:
      | name               | username | roles                       |
      | AdminAndContestant | user04   | admin,contestant            |
    And can login with the new user:
      """
      {
        "username": "user04",
        "password": "12345678"
      }
      """

  Scenario: Successful update user with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "PATCH" request to "/api/v1/users/13" with the following data:
      """
      {
        "roles": []
      }
      """
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true
      }
      """
    And the user should be updated with the following data:
      | name       | username | roles                         |
      | Contestant | user03   |                               |
    And can login with the updated user:
      """
      {
        "username": "user03",
        "password": "123456"
      }
      """

  Scenario: Failed update user with invalid roles
    Given I am logged in as "user01"
    When I send a "PATCH" request to "/api/v1/users/11" with the following data:
      """
      {
        "roles": [-1]
      }
      """
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "invalid role",
        "details": null
      }
      """

  Scenario: Unauthorized update user
    When I send a "PATCH" request to "/api/v1/users/11" with the following data:
      """
      {
        "name": "new user01",
        "password": "12345678",
        "roles": [1, 2, 3]
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
    
  Scenario: Forbidden update user with "APIReader" user
    Given I am logged in as "user02"
    When I send a "PATCH" request to "/api/v1/users/11" with the following data:
      """
      {
        "name": "new user01",
        "password": "12345678",
        "roles": [1, 2, 3]
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

  Scenario: Forbidden update user with "Contestant" user
    Given I am logged in as "user03"
    When I send a "PATCH" request to "/api/v1/users/11" with the following data:
      """
      {
        "name": "new user01",
        "password": "12345678",
        "roles": [1, 2, 3]
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



