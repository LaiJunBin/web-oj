Feature: List Contest Participants API
  As an admin
  I want to list participants of a contest
  So that I can manage who is participating

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

  Scenario: Successfully list empty participants
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/contests/1/participants":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "userIds": [],
          "groupIds": []
        }
      }
      """

  Scenario: Successfully list participants with users and groups
    Given I have groups:
      | id | name    | description |
      | 1  | Group-1 | desc        |
    And I have contest participants:
      | contestId | userId | groupId |
      | 1         | 13     |         |
      | 1         | 14     |         |
      | 1         |        | 1       |
    And I am logged in as "user01"
    When I send a "GET" request to "/api/v1/contests/1/participants":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "userIds": [13, 14],
          "groupIds": [1]
        }
      }
      """

  Scenario: Unauthorized
    When I send a "GET" request to "/api/v1/contests/1/participants":
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """

  Scenario: Forbidden for contestant
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/contests/1/participants":
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """
