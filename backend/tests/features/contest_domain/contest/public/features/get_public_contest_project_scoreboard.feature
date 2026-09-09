Feature: Get Public Contest Project Scoreboard API
  As a public user
  I want to view the scoreboard for a public contest project
  So that I can see the rankings without logging in

  Background:
    Given I have users:
      | id | name       | username | password | roles      |
      | 13 | Contester  | user03   | 123456   | contestant |
      | 14 | Contester2 | user04   | 123456   | contestant |
    And I have images:
      | id | tag               | mount_path    | description | is_ready | createdAt           | updatedAt           |
      | 1  | nginx:latest      | /var/www/html | desc1       | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 2  | test-nginx:latest | /var/www/html | desc2       | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And I have projects:
      | id | name     | description | imageTag     | env | testImageTag      | testEnv | resultParser | createdAt           | updatedAt           |
      | 1  | Project1 | desc        | nginx:latest |     | test-nginx:latest |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have grading-items:
      | id | projectId | name | description | score | order | scoringStrategy | outputFormat | resultParser | label | createdAt | updatedAt |
      | 1 | 1 | Task 1 | desc | 100 | 1 | increase | text |  |  | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And I have contests:
      | id | name    | description | isPublic | activateTime        | deactivateTime      | createdAt           | updatedAt           |
      | 1  | Contest | desc        | 1        | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have contest project associations:
      | id | contestId | projectId | name     | isPublic | judge priority strategy | activateTime        | deactivateTime      | startTime           | endTime             |
      | 1  | 1         | 1         | Project1 | 1        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 |

  Scenario: Successfully get public scoreboard with direct participant
    Given I have contest participants:
      | contestId | userId | groupId |
      | 1         | 13     |         |
    When I send a "GET" request to "/api/v1/public/contest-projects/1/scoreboard":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "compact": false,
          "heads": [
            {"id": 1, "name": "Task 1"}
          ],
          "rows": [
            {"name": "Contester", "username": "user03", "statuses": ["NotSubmitted"]}
          ]
        }
      }
      """

  Scenario: Scoreboard includes group participants
    Given I have groups:
      | id | name    | description |
      | 1  | Group-1 | desc        |
    And I have user groups:
      | userId | groupId |
      | 13     | 1       |
      | 14     | 1       |
    And I have contest participants:
      | contestId | userId | groupId |
      | 1         |        | 1       |
    When I send a "GET" request to "/api/v1/public/contest-projects/1/scoreboard":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "compact": false,
          "heads": [
            {"id": 1, "name": "Task 1"}
          ],
          "rows": [
            {"name": "Contester", "username": "user03", "statuses": ["NotSubmitted"]},
            {"name": "Contester2", "username": "user04", "statuses": ["NotSubmitted"]}
          ]
        }
      }
      """

  Scenario: Scoreboard is empty when no participants are configured
    When I send a "GET" request to "/api/v1/public/contest-projects/1/scoreboard":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "compact": false,
          "heads": [
            {"id": 1, "name": "Task 1"}
          ],
          "rows": []
        }
      }
      """

  Scenario: Not found for non-existent contest project
    When I send a "GET" request to "/api/v1/public/contest-projects/999/scoreboard":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "project not found",
        "details": null
      }
      """

  Scenario: Not found for private contest project
    And I have contest project associations:
      | id | contestId | projectId | name            | isPublic | judge priority strategy | activateTime        | deactivateTime      | startTime           | endTime             |
      | 2  | 1         | 1         | PrivateProject1 | 0        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 |
    When I send a "GET" request to "/api/v1/public/contest-projects/2/scoreboard":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "Project not found",
        "details": null
      }
      """
