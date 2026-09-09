Feature: Rejudge Submission
  As an admin
  I want to rejudge a submission
  So that I can re-evaluate it with updated test configurations

  Background:
    Given I have users:
      | id | name      | username | password | roles      |
      | 11 | AdminUser | admin01  | 123456   | admin      |
      | 12 | Reader    | reader01 | 123456   | api_reader |
    And I have images:
      | id | tag          | mountPath | description | isReady | createdAt           | updatedAt           |
      | 1  | img:latest   | /app      | test image  | true    | 2024-01-01 00:00:00 | 2024-01-01 00:00:00 |
      | 2  | test:latest  | /app      | test runner | true    | 2024-01-01 00:00:00 | 2024-01-01 00:00:00 |
    And I have projects:
      | id | name     | description | imageTag    | env | testImageTag | testEnv | resultParser | createdAt           | updatedAt           |
      | 1  | Project1 | desc        | img:latest  |     | test:latest  |         |  | 2024-01-01 00:00:00 | 2024-01-01 00:00:00 |
    And I have grading-items:
      | id | projectID | name | description | score | order | scoringStrategy | outputFormat | resultParser | label | createdAt | updatedAt |
      | 1 | 1 | Task 1 | desc | 100 | 1 | increase | text |  |  | 2024-01-01 00:00:00 | 2024-01-01 00:00:00 |
    And I have testcases:
      | id | gradingItemID | name    | command     | score | isSample | createdAt           | updatedAt           |
      | 1  | 1             | Test 01 | echo hello  | 50    | false    | 2024-01-01 00:00:00 | 2024-01-01 00:00:00 |
      | 2  | 1             | Test 02 | echo world  | 50    | false    | 2024-01-01 00:00:00 | 2024-01-01 00:00:00 |
    And I have contests:
      | id | name    | description | isPublic | activateTime        | deactivateTime      | createdAt           | updatedAt           |
      | 1  | Contest | desc        | 1        | 2024-01-01 00:00:00 | 2099-12-31 23:59:59 | 2024-01-01 00:00:00 | 2024-01-01 00:00:00 |
    And I have contest project associations:
      | id | contestId | projectId | name     | isPublic | judge priority strategy | activateTime        | deactivateTime      | startTime           | endTime             |
      | 1  | 1         | 1         | Project1 | 1        | default                 | 2024-01-01 00:00:00 | 2099-12-31 23:59:59 | 2024-01-01 00:00:00 | 2099-12-31 23:59:59 |
    And I have submissions:
      | id   | userID | contestProjectID |
      | sub1 | 11     | 1                |

  Scenario: Successfully rejudge a submission
    Given I am logged in as "admin01"
    And submission "sub1" has stored judge files
    When I send a "POST" request to "/api/v1/submissions/sub1/rejudge" with the following data:
      """
      {
        "reason": "Updated test cases"
      }
      """
    Then the response code should be 200
    And a judgment should exist for submission "sub1"
    And the judgment for submission "sub1" should reach status "Rejudged"
    And a rejudge should exist with status "Finished"
    # A single-submission rejudge takes the project from the submission itself.
    # The admin console files rejudges under their contest project, so one that
    # arrives without it is filed under project 0.
    When I send a "GET" request to "/api/v1/rejudges":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
          {
            "id": "{id}",
            "contestProjectId": 1,
            "reason": "Updated test cases",
            "status": "Finished",
            "effectedJudgments": 1,
            "completedJudgments": 1,
            "createdAt": "{createdAt}"
          }
        ]
      }
      """

  Scenario: Rejudge requires a reason
    Given I am logged in as "admin01"
    When I send a "POST" request to "/api/v1/submissions/sub1/rejudge" with the following data:
      """
      {
        "reason": ""
      }
      """
    Then the response code should be 400

  Scenario: Cannot rejudge non-existent submission
    Given I am logged in as "admin01"
    When I send a "POST" request to "/api/v1/submissions/nonexistent/rejudge" with the following data:
      """
      {
        "reason": "test"
      }
      """
    Then the response code should be 500

  Scenario: Forbidden rejudge for read-only user
    Given I am logged in as "reader01"
    When I send a "POST" request to "/api/v1/submissions/sub1/rejudge" with the following data:
      """
      {
        "reason": "test"
      }
      """
    Then the response code should be 403

  Scenario: Unauthorized rejudge without login
    When I send a "POST" request to "/api/v1/submissions/sub1/rejudge" with the following data:
      """
      {
        "reason": "test"
      }
      """
    Then the response code should be 401
