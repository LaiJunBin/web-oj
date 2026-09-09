Feature: List User API
  As a user
  I want to list all users
  So that I can see all users if I have permission

  Background:
    Given I have users:
      | id  | name               | username | password | roles              |
      | 11  | AdminUser          | user01   | 123456   | admin              |
      | 12  | APIReader          | user02   | 123456   | api_reader         |
      | 13  | Contestant         | user03   | 123456   | contestant         |
      | 14  | AdminAndContestant | user04   | 123456   | admin,contestant   |

  Scenario: Successful list users with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/users":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
            {
              "id": 11,
              "name": "AdminUser",
              "username": "user01",
              "roles": ["admin"],
              "createdAt": "{createdAt}",
              "updatedAt": "{updatedAt}"
            },
            {
              "id": 12,
              "name": "APIReader",
              "username": "user02",
              "roles": ["api_reader"],
              "createdAt": "{createdAt}",
              "updatedAt": "{updatedAt}"
            },
            {
              "id": 13,
              "name": "Contestant",
              "username": "user03",
              "roles": ["contestant"],
              "createdAt": "{createdAt}",
              "updatedAt": "{updatedAt}"
            },
            {
              "id": 14,
              "name": "AdminAndContestant",
              "username": "user04",
              "roles": ["admin", "contestant"],
              "createdAt": "{createdAt}",
              "updatedAt": "{updatedAt}"
            }
          ]
      }
      """
  
  Scenario: Successful list users with "APIReader" user
    Given I am logged in as "user02"
    When I send a "GET" request to "/api/v1/users":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
            {
              "id": 11,
              "name": "AdminUser",
              "username": "user01",
              "roles": ["admin"],
              "createdAt": "{createdAt}",
              "updatedAt": "{updatedAt}"
            },
            {
              "id": 12,
              "name": "APIReader",
              "username": "user02",
              "roles": ["api_reader"],
              "createdAt": "{createdAt}",
              "updatedAt": "{updatedAt}"
            },
            {
              "id": 13,
              "name": "Contestant",
              "username": "user03",
              "roles": ["contestant"],
              "createdAt": "{createdAt}",
              "updatedAt": "{updatedAt}"
            },
            {
              "id": 14,
              "name": "AdminAndContestant",
              "username": "user04",
              "roles": ["admin", "contestant"],
              "createdAt": "{createdAt}",
              "updatedAt": "{updatedAt}"
            }
          ]
      }
      """
      
  Scenario: Successful list users with "AdminAndContestant" user
    Given I am logged in as "user04"
    When I send a "GET" request to "/api/v1/users":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
            {
              "id": 11,
              "name": "AdminUser",
              "username": "user01",
              "roles": ["admin"],
              "createdAt": "{createdAt}",
              "updatedAt": "{updatedAt}"
            },
            {
              "id": 12,
              "name": "APIReader",
              "username": "user02",
              "roles": ["api_reader"],
              "createdAt": "{createdAt}",
              "updatedAt": "{updatedAt}"
            },
            {
              "id": 13,
              "name": "Contestant",
              "username": "user03",
              "roles": ["contestant"],
              "createdAt": "{createdAt}",
              "updatedAt": "{updatedAt}"
            },
            {
              "id": 14,
              "name": "AdminAndContestant",
              "username": "user04",
              "roles": ["admin", "contestant"],
              "createdAt": "{createdAt}",
              "updatedAt": "{updatedAt}"
            }
          ]
      }
      """

  Scenario: Unauthorized list users
    When I send a "GET" request to "/api/v1/users":
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """

  Scenario: Forbidden list users with "Contestant" user
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/users":
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """


