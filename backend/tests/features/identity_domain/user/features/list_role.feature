Feature: List Role API
  As a user
  I want to list all roles
  So that I can see all roles if I have permission

  Background:
    Given I have users:
      | id  | name               | username | password | roles              |
      | 11  | AdminUser          | user01   | 123456   | admin              |
      | 12  | APIReader          | user02   | 123456   | api_reader         |
      | 13  | Contestant         | user03   | 123456   | contestant         |
      | 14  | AdminAndContestant | user04   | 123456   | admin,contestant   |

  Scenario: Successful list roles with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/roles":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
            {
              "id": 1,
              "name": "admin"
            },
            {
              "id": 2,
              "name": "contestant"
            },
            {
              "id": 3,
              "name": "api_reader"
            }
          ]
      }
      """
  
  Scenario: Successful list roles with "APIReader" user
    Given I am logged in as "user02"
    When I send a "GET" request to "/api/v1/roles":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
            {
              "id": 1,
              "name": "admin"
            },
            {
              "id": 2,
              "name": "contestant"
            },
            {
              "id": 3,
              "name": "api_reader"
            }
          ]
      }
      """
      
  Scenario: Successful list roles with "AdminAndContestant" user
    Given I am logged in as "user04"
    When I send a "GET" request to "/api/v1/roles":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
            {
              "id": 1,
              "name": "admin"
            },
            {
              "id": 2,
              "name": "contestant"
            },
            {
              "id": 3,
              "name": "api_reader"
            }
          ]
      }
      """

  Scenario: Unauthorized list roles
    When I send a "GET" request to "/api/v1/roles":
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """

  Scenario: Forbidden list roles with "Contestant" user
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/roles":
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """


