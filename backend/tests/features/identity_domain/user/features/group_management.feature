Feature: Group Management API
  As an admin
  I want to manage user groups
  So that I can grant contest access to many users at once

  Background:
    Given I have users:
      | id | name       | username | password | roles      |
      | 11 | AdminUser  | user01   | 123456   | admin      |
      | 12 | APIReader  | user02   | 123456   | api_reader |
      | 13 | Contestant | user03   | 123456   | contestant |
      | 14 | Contestant2| user04   | 123456   | contestant |

  Scenario: Successfully create a group
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/groups" with the following data:
      """
      {"name": "Group-1", "description": "desc"}
      """
    Then the response code should be 200
    And the response should match json:
      """
      {"success": true, "data": 1}
      """

  Scenario: Successfully create a group without description
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/groups" with the following data:
      """
      {"name": "Group-1"}
      """
    Then the response code should be 200
    And the response should match json:
      """
      {"success": true, "data": 1}
      """

  Scenario: Successfully list groups
    Given I have groups:
      | id | name    | description |
      | 1  | Group-1 | desc-1      |
      | 2  | Group-2 | desc-2      |
    And I am logged in as "user01"
    When I send a "GET" request to "/api/v1/groups":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
          {"id": 1, "name": "Group-1", "description": "desc-1"},
          {"id": 2, "name": "Group-2", "description": "desc-2"}
        ]
      }
      """

  Scenario: Successfully get a group with members
    Given I have groups:
      | id | name    | description |
      | 1  | Group-1 | desc-1      |
    And I have user groups:
      | userId | groupId |
      | 13     | 1       |
      | 14     | 1       |
    And I am logged in as "user01"
    When I send a "GET" request to "/api/v1/groups/1":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {"id": 1, "name": "Group-1", "description": "desc-1", "userIds": [13, 14]}
      }
      """

  Scenario: Successfully update a group
    Given I have groups:
      | id | name    | description |
      | 1  | Group-1 | desc-1      |
    And I am logged in as "user01"
    When I send a "PATCH" request to "/api/v1/groups/1" with the following data:
      """
      {"name": "Group-1-updated"}
      """
    Then the response code should be 200
    And the response should match json:
      """
      {"success": true}
      """
    When I send a "GET" request to "/api/v1/groups/1":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {"id": 1, "name": "Group-1-updated", "description": "desc-1", "userIds": []}
      }
      """

  Scenario: Successfully delete a group
    Given I have groups:
      | id | name    | description |
      | 1  | Group-1 | desc-1      |
    And I am logged in as "user01"
    When I send a "DELETE" request to "/api/v1/groups/1":
    Then the response code should be 200
    And the response should match json:
      """
      {"success": true}
      """
    When I send a "GET" request to "/api/v1/groups/1":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "group not found",
        "details": null
      }
      """

  Scenario: Successfully add and remove group members
    Given I have groups:
      | id | name    | description |
      | 1  | Group-1 | desc-1      |
    And I am logged in as "user01"
    When I send a "POST" request to "/api/v1/groups/1/users" with the following data:
      """
      {"userIds": [13, 14]}
      """
    Then the response code should be 200
    And the response should match json:
      """
      {"success": true}
      """
    When I send a "DELETE" request to "/api/v1/groups/1/users" with the following data:
      """
      {"userIds": [14]}
      """
    Then the response code should be 200
    And the response should match json:
      """
      {"success": true}
      """
    When I send a "GET" request to "/api/v1/groups/1":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {"id": 1, "name": "Group-1", "description": "desc-1", "userIds": [13]}
      }
      """

  Scenario: Failed to get non-existent group
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/groups/99":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "group not found",
        "details": null
      }
      """

  Scenario: Unauthorized without login
    When I send a "GET" request to "/api/v1/groups":
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """

  Scenario: Forbidden for contestant to list groups
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/groups":
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: Forbidden for contestant to modify group members
    Given I have groups:
      | id | name    | description |
      | 1  | Group-1 | desc-1      |
    And I am logged in as "user03"
    When I send a "POST" request to "/api/v1/groups/1/users" with the following data:
      """
      {"userIds": [13]}
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

  Scenario: Forbidden for api_reader to create group
    Given I am logged in as "user02"
    When I send a "POST" request to "/api/v1/groups" with the following data:
      """
      {"name": "Group-1"}
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
