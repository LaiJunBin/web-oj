Feature: Add Users to Contest Participants API
  As an admin
  I want to add users to a contest's participant list
  So that they are allowed to participate

  Background:
    Given I have users:
      | id | name      | username | password | roles      |
      | 11 | AdminUser | user01   | 123456   | admin      |
      | 12 | APIReader | user02   | 123456   | api_reader |
      | 13 | User1     | user03   | 123456   | contestant |
      | 14 | User2     | user04   | 123456   | contestant |
    And I have contests:
      | id | name      | description | isPublic | activateTime        | deactivateTime      | createdAt           | updatedAt           |
      | 1  | Contest-1 | desc        | 0        | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |

  Scenario: Successfully add users to contest
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/contests/1/participants/users" with the following data:
      """
      {"userIds": [13, 14]}
      """
    Then the response code should be 200
    And the response should match json:
      """
      {"success": true}
      """
    When I send a "GET" request to "/api/v1/contests/1/participants":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "userIds": [13, 14],
          "groupIds": []
        }
      }
      """

  Scenario: Adding duplicate users is idempotent
    Given I have contest participants:
      | contestId | userId | groupId |
      | 1         | 13     |         |
    And I am logged in as "user01"
    When I send a "POST" request to "/api/v1/contests/1/participants/users" with the following data:
      """
      {"userIds": [13]}
      """
    Then the response code should be 200
    And the response should match json:
      """
      {"success": true}
      """

  Scenario: Adding with empty userIds is a bad request
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/contests/1/participants/users" with the following data:
      """
      {"userIds": []}
      """
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "userIds is required",
        "details": null
      }
      """

  Scenario: Unauthorized
    When I send a "POST" request to "/api/v1/contests/1/participants/users" with the following data:
      """
      {"userIds": [13]}
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

  Scenario: Forbidden for api_reader
    Given I am logged in as "user02"
    When I send a "POST" request to "/api/v1/contests/1/participants/users" with the following data:
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

  Scenario: Forbidden for contestant
    Given I am logged in as "user03"
    When I send a "POST" request to "/api/v1/contests/1/participants/users" with the following data:
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
