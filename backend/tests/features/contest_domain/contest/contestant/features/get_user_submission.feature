Feature: Get User Submission API
  As a contestant
  I want to get details of my submission
  So that I can review my submission result

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
  Scenario: Successfully get submission detail for a contestant
    Given I am logged in as "user03"
    And I have submissions:
      | id   | userID | contestProjectID | ignored |
      | sub1 | 13     | 1                | false   |
    And I have judgments:
      | id   | submissionID | status | judgeType |
      | jdg1 | sub1         | Queued | JUDGE     |
    When I send a "GET" request to "/api/v1/user/submissions/sub1":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": "sub1",
          "status": "Queued",
          "isPartial": false,
          "score": 0,
          "totalScore": 0,
          "gradingItems": [],
          "createdAt": "{createdAt}"
        }
      }
      """

  Scenario: Contestant cannot get another user's submission
    Given I am logged in as "user03"
    And I have submissions:
      | id   | userID | contestProjectID |
      | sub1 | 11     | 1                |
    When I send a "GET" request to "/api/v1/user/submissions/sub1":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "submission not found",
        "details": null
      }
      """

  Scenario: Submission without judgment is treated as not found
    Given I am logged in as "user03"
    And I have submissions:
      | id   | userID | contestProjectID |
      | sub2 | 13     | 1                |
    When I send a "GET" request to "/api/v1/user/submissions/sub2":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "submission not found",
        "details": null
      }
      """

  Scenario: Not found for non-existent submission
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/user/submissions/nonexistent":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "submission not found",
        "details": null
      }
      """

  Scenario: Unauthorized access without login
    When I send a "GET" request to "/api/v1/user/submissions/sub1":
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
    When I send a "GET" request to "/api/v1/user/submissions/sub1":
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  # The reason a judgment recorded (an Error's neutral explanation) is part of
  # the contestant's view: without it an Error is a dead end with nothing to
  # act on. errorMessage — the operator's detail — stays on the admin API.
  Scenario: An errored submission's detail carries its reason
    Given I am logged in as "user03"
    And I have submissions:
      | id   | userID | contestProjectID | ignored |
      | sub1 | 13     | 1                | false   |
    And I have judgments:
      | id   | submissionID | status | judgeType | reason                                 |
      | jdg1 | sub1         | Error  | JUDGE     | 評測前置步驟失敗，請聯絡現場工作人員 |
    When I send a "GET" request to "/api/v1/user/submissions/sub1":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": "sub1",
          "status": "Error",
          "isPartial": false,
          "score": 0,
          "totalScore": 0,
          "reason": "評測前置步驟失敗，請聯絡現場工作人員",
          "gradingItems": [],
          "createdAt": "{createdAt}"
        }
      }
      """

  # A staff note is the operator speaking to THIS contestant about THIS
  # judgment. It is independent of reason: a healthy Judged submission may
  # carry one, and an Error may carry both.
  Scenario: A submission's detail carries the staff note next to the reason
    Given I am logged in as "user03"
    And I have submissions:
      | id   | userID | contestProjectID | ignored |
      | sub1 | 13     | 1                | false   |
    And I have judgments:
      | id   | submissionID | status | judgeType | reason                                 | staffNote                    |
      | jdg1 | sub1         | Error  | JUDGE     | 評測前置步驟失敗，請聯絡現場工作人員 | 已通知助教處理，請稍候再提交 |
    When I send a "GET" request to "/api/v1/user/submissions/sub1":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": "sub1",
          "status": "Error",
          "isPartial": false,
          "score": 0,
          "totalScore": 0,
          "reason": "評測前置步驟失敗，請聯絡現場工作人員",
          "staffNote": "已通知助教處理，請稍候再提交",
          "gradingItems": [],
          "createdAt": "{createdAt}"
        }
      }
      """

  Scenario: A staff note on a judged submission stands alone, without a reason
    Given I am logged in as "user03"
    And I have submissions:
      | id   | userID | contestProjectID | ignored |
      | sub1 | 13     | 1                | false   |
    And I have judgments:
      | id   | submissionID | status | judgeType | score | totalScore | staffNote                  |
      | jdg1 | sub1         | Judged | JUDGE     | 100   | 100        | 做得好，注意第 2 題的邊界 |
    When I send a "GET" request to "/api/v1/user/submissions/sub1":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": "sub1",
          "status": "Judged",
          "isPartial": false,
          "score": 100,
          "totalScore": 100,
          "staffNote": "做得好，注意第 2 題的邊界",
          "gradingItems": [],
          "createdAt": "{createdAt}"
        }
      }
      """
