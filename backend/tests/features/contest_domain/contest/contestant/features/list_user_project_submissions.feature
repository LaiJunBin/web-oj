Feature: List User Project Submissions API
  As a contestant
  I want to list my submissions for a contest project
  So that I can track my submission history

  Background:
    Given I have users:
      | id | name               | username | password | roles            |
      | 11 | AdminUser          | user01   | 123456   | admin            |
      | 12 | APIReader          | user02   | 123456   | api_reader       |
      | 13 | Contestant         | user03   | 123456   | contestant       |
      | 14 | AdminAndContestant | user04   | 123456   | admin,contestant |
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
  Scenario: Successfully list submissions (empty) for a contestant
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contest-projects/1/submissions":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": []
      }
      """

  Scenario: Successfully list submissions with data for a contestant
    Given I am logged in as "user03"
    And I have submissions:
      | id   | userID | contestProjectID | ignored |
      | sub1 | 13     | 1                | false   |
    And I have judgments:
      | id   | submissionID | status | judgeType |
      | jdg1 | sub1         | Queued | JUDGE     |
    When I send a "GET" request to "/api/v1/user/contest-projects/1/submissions":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
          {
            "id": "sub1",
            "status": "Queued",
            "isPartial": false,
            "score": 0,
            "totalScore": 0,
            "createdAt": "{createdAt}"
          }
        ]
      }
      """

  Scenario: Submission without judgment is excluded from the list
    Given I am logged in as "user03"
    And I have submissions:
      | id   | userID | contestProjectID | ignored |
      | sub1 | 13     | 1                | false   |
      | sub2 | 13     | 1                | false   |
    And I have judgments:
      | id   | submissionID | status | judgeType |
      | jdg1 | sub1         | Queued | JUDGE     |
    When I send a "GET" request to "/api/v1/user/contest-projects/1/submissions":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
          {
            "id": "sub1",
            "status": "Queued",
            "isPartial": false,
            "score": 0,
            "totalScore": 0,
            "createdAt": "{createdAt}"
          }
        ]
      }
      """

  Scenario: Contestant cannot see another user's submissions
    Given I am logged in as "user03"
    And I have submissions:
      | id   | userID | contestProjectID |
      | sub1 | 11     | 1                |
    When I send a "GET" request to "/api/v1/user/contest-projects/1/submissions":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": []
      }
      """

  Scenario: Not found for non-existent contest project
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/contest-projects/999/submissions":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "contest project not found",
        "details": null
      }
      """

  Scenario: Unauthorized access without login
    When I send a "GET" request to "/api/v1/user/contest-projects/1/submissions":
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """

  Scenario: Forbidden access for api_reader
    Given I am logged in as "user02"
    When I send a "GET" request to "/api/v1/user/contest-projects/1/submissions":
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  # The list is where a contestant first sees an Error; the reason rides along
  # so they do not have to open the detail to learn there is nothing to fix on
  # their side. Submissions without a recorded reason simply omit the field.
  Scenario: An errored submission's list entry carries its reason
    Given I am logged in as "user03"
    And I have submissions:
      | id   | userID | contestProjectID | ignored |
      | sub1 | 13     | 1                | false   |
    And I have judgments:
      | id   | submissionID | status | judgeType | reason                                 |
      | jdg1 | sub1         | Error  | JUDGE     | 評測前置步驟失敗，請聯絡現場工作人員 |
    When I send a "GET" request to "/api/v1/user/contest-projects/1/submissions":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
          {
            "id": "sub1",
            "status": "Error",
            "isPartial": false,
            "score": 0,
            "totalScore": 0,
            "reason": "評測前置步驟失敗，請聯絡現場工作人員",
            "createdAt": "{createdAt}"
          }
        ]
      }
      """

  # A staff note reaches the list the same way reason does: the list is where
  # a contestant first looks, so a note they could only find by opening the
  # detail is a note they miss. Entries without a note omit the field, and
  # the note is independent of reason — a Judged entry may carry one alone.
  Scenario: List entries carry their staff note, and omit it when there is none
    Given I am logged in as "user03"
    And I have submissions:
      | id   | userID | contestProjectID | ignored | createdAt            |
      | sub1 | 13     | 1                | false   | 2024-01-01T00:00:00Z |
      | sub2 | 13     | 1                | false   | 2024-01-01T00:01:00Z |
    And I have judgments:
      | id   | submissionID | status | judgeType | score | totalScore | staffNote                  |
      | jdg1 | sub1         | Judged | JUDGE     | 100   | 100        | 做得好，注意第 2 題的邊界 |
    And I have judgments:
      | id   | submissionID | status | judgeType |
      | jdg2 | sub2         | Queued | JUDGE     |
    When I send a "GET" request to "/api/v1/user/contest-projects/1/submissions":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
          {
            "id": "sub2",
            "status": "Queued",
            "isPartial": false,
            "score": 0,
            "totalScore": 0,
            "createdAt": "{createdAt}"
          },
          {
            "id": "sub1",
            "status": "Judged",
            "isPartial": false,
            "score": 100,
            "totalScore": 100,
            "staffNote": "做得好，注意第 2 題的邊界",
            "createdAt": "{createdAt}"
          }
        ]
      }
      """

  Scenario: An errored list entry carries both its reason and the staff note
    Given I am logged in as "user03"
    And I have submissions:
      | id   | userID | contestProjectID | ignored |
      | sub1 | 13     | 1                | false   |
    And I have judgments:
      | id   | submissionID | status | judgeType | reason                                 | staffNote                    |
      | jdg1 | sub1         | Error  | JUDGE     | 評測前置步驟失敗，請聯絡現場工作人員 | 已通知助教處理，請稍候再提交 |
    When I send a "GET" request to "/api/v1/user/contest-projects/1/submissions":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": [
          {
            "id": "sub1",
            "status": "Error",
            "isPartial": false,
            "score": 0,
            "totalScore": 0,
            "reason": "評測前置步驟失敗，請聯絡現場工作人員",
            "staffNote": "已通知助教處理，請稍候再提交",
            "createdAt": "{createdAt}"
          }
        ]
      }
      """
