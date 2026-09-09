Feature: Get Submission API
  As an admin
  I want to get details of a specific submission
  So that I can review the full submission with all judgments

  Background:
    Given I have users:
      | id | name       | username | password | roles      |
      | 11 | AdminUser  | user01   | 123456   | admin      |
      | 13 | Contestant | user03   | 123456   | contestant |
    And I have images:
      | id | tag               | mount_path    | description | is_ready | createdAt           | updatedAt           |
      | 1  | nginx:latest      | /var/www/html | desc1       | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 2  | test-nginx:latest | /var/www/html | desc2       | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And I have projects:
      | id | name           | description                | imageTag     | env | testImageTag      | testEnv | resultParser | createdAt           | updatedAt           |
      | 11 | Test-Project-1 | Test-Project-1 Description | nginx:latest |     | test-nginx:latest |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have contests:
      | id | name           | description                | isPublic | activateTime        | deactivateTime      | createdAt           | updatedAt           |
      | 11 | Test-Contest-1 | Test-Contest-1 Description | 0        | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have contest project associations:
      | id | contestId | projectId | name           | isPublic | judge priority strategy | activateTime        | deactivateTime      | startTime           | endTime             |
      | 1  | 11        | 11        | Test-Project-1 | 0        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 |
    And I have submissions:
      | id   | userID | contestProjectID |
      | sub1 | 11     | 1                |

  Scenario: Submission without judgment is treated as not found
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/submissions/sub1":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "submission not found",
        "details": null
      }
      """

  Scenario: Admin can get a submission detail with judgment timestamps
    Given I have judgments:
      | id   | submissionID | status  | judgeType | createdAt             | judgingStartedAt      | judgingFinishedAt     |
      | jdg1 | sub1         | Judged  | JUDGE     | 2024-01-01T00:00:00Z  | 2024-01-01T00:00:05Z  | 2024-01-01T00:00:12Z  |
    And I am logged in as "user01"
    When I send a "GET" request to "/api/v1/submissions/sub1":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": "sub1",
          "isPartial": false,
          "ignored": false,
          "createdAt": "{createdAt}",
          "contest": {"id": 11, "name": "Test-Contest-1"},
          "project": {"id": 11, "name": "Test-Project-1"},
          "user": {"id": 11, "username": "user01", "name": "AdminUser"},
          "gradingItems": [],
          "judgments": [
            {
              "id": "jdg1",
              "judgeType": "JUDGE",
              "status": "Judged",
              "score": 0,
              "totalScore": 0,
              "rejudgeId": null,
              "judgingStartedAt": "2024-01-01T00:00:05Z",
              "judgingFinishedAt": "2024-01-01T00:00:12Z",
              "errorMessage": null,
              "staffNote": null,
              "project": {
                "id": 0,
                "imageTag": "",
                "testImageTag": "",
                "env": "",
                "testEnv": "",
                "mountPath": "",
                "resultParser": ""
              },
              "createdAt": "2024-01-01T00:00:00Z",
              "gradingItems": [],
              "judgmentLifecycle": {
                "beforeAll": [],
                "afterAll": []
              }
            }
          ]
        }
      }
      """

  # Every grading item in a judgment response used to report label "" and no name,
  # because the fields existed in the schema but nothing ever populated them —
  # there was no coverage below `"gradingItems": []` to notice. The values come
  # from the judgment's own snapshot, so this also pins that the testcase `score`
  # is the earned score with the configured max alongside it in `totalScore`.
  Scenario: Judgment grading items report their label, name and both scores
    Given I have judgments:
      | id   | submissionID | status | judgeType | score | totalScore | createdAt            | gradingItems                                                                                                                                                                                                            |
      | jdg1 | sub1         | Judged | JUDGE     | 10    | 30         | 2024-01-01T00:00:00Z | [{"id":1,"label":"2","name":"Arithmetic","status":"Failed","score":10,"totalScore":30,"scoringStrategy":"increase","resultParser":"auto","testcases":[{"id":7,"name":"handles overflow","status":"Failed","score":0,"totalScore":20,"parsedScore":5,"command":"run overflow"}]}] |
    And I am logged in as "user01"
    When I send a "GET" request to "/api/v1/submissions/sub1":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": "sub1",
          "isPartial": false,
          "ignored": false,
          "createdAt": "{createdAt}",
          "contest": {"id": 11, "name": "Test-Contest-1"},
          "project": {"id": 11, "name": "Test-Project-1"},
          "user": {"id": 11, "username": "user01", "name": "AdminUser"},
          "gradingItems": [],
          "judgments": [
            {
              "id": "jdg1",
              "judgeType": "JUDGE",
              "status": "Judged",
              "score": 10,
              "totalScore": 30,
              "rejudgeId": null,
              "judgingStartedAt": null,
              "judgingFinishedAt": null,
              "errorMessage": null,
              "staffNote": null,
              "project": {
                "id": 0,
                "imageTag": "",
                "testImageTag": "",
                "env": "",
                "testEnv": "",
                "mountPath": "",
                "resultParser": ""
              },
              "createdAt": "2024-01-01T00:00:00Z",
              "gradingItems": [
                {
                  "id": 1,
                  "label": "2",
                  "name": "Arithmetic",
                  "status": "Failed",
                  "score": 10,
                  "totalScore": 30,
                  "scoringStrategy": "increase",
                  "resultParser": "auto",
                  "testcases": [
                    {
                      "id": 7,
                      "name": "handles overflow",
                      "status": "Failed",
                      "score": 0,
                      "totalScore": 20,
                      "parsedScore": 5,
                      "command": "run overflow",
                      "lifecycle": { "beforeEach": [], "afterEach": [] },
                      "result": { "exitCode": 0, "execTime": 0, "stdout": "", "stderr": "" }
                    }
                  ]
                }
              ],
              "judgmentLifecycle": {
                "beforeAll": [],
                "afterAll": []
              }
            }
          ]
        }
      }
      """

  Scenario: Submission from a deleted user is treated as not found
    Given I have submissions:
      | id   | userID | contestProjectID |
      | sub2 | 13     | 1                |
    And I have judgments:
      | id   | submissionID | status | judgeType | score | totalScore |
      | jdg2 | sub2         | Judged | NORMAL    | 75    | 100        |
    And I am logged in as "user01"
    When I send a "DELETE" request to "/api/v1/users/13":
    Then the response code should be 200
    When I send a "GET" request to "/api/v1/submissions/sub2":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "submission not found",
        "details": null
      }
      """

  Scenario: Submission not found
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/submissions/nonexistent":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "submission not found",
        "details": null
      }
      """

  Scenario: Unauthorized access
    When I send a "GET" request to "/api/v1/submissions/sub1":
    Then the response code should be 401

  Scenario: Forbidden for contestant role
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/submissions/sub1":
    Then the response code should be 403

  # The admin detail echoes the staff note so the edit form can show what
  # the contestant currently sees; it sits beside errorMessage, which stays
  # admin-only.
  Scenario: Admin detail echoes the judgment's staff note
    Given I have judgments:
      | id   | submissionID | status | judgeType | createdAt            | staffNote                    |
      | jdg1 | sub1         | Judged | JUDGE     | 2024-01-01T00:00:00Z | 第 3 題輸出多了一個換行，請對照範例 |
    And I am logged in as "user01"
    When I send a "GET" request to "/api/v1/submissions/sub1":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": "sub1",
          "isPartial": false,
          "ignored": false,
          "createdAt": "{createdAt}",
          "contest": {"id": 11, "name": "Test-Contest-1"},
          "project": {"id": 11, "name": "Test-Project-1"},
          "user": {"id": 11, "username": "user01", "name": "AdminUser"},
          "gradingItems": [],
          "judgments": [
            {
              "id": "jdg1",
              "judgeType": "JUDGE",
              "status": "Judged",
              "score": 0,
              "totalScore": 0,
              "rejudgeId": null,
              "judgingStartedAt": null,
              "judgingFinishedAt": null,
              "errorMessage": null,
              "staffNote": "第 3 題輸出多了一個換行，請對照範例",
              "project": {
                "id": 0,
                "imageTag": "",
                "testImageTag": "",
                "env": "",
                "testEnv": "",
                "mountPath": "",
                "resultParser": ""
              },
              "createdAt": "2024-01-01T00:00:00Z",
              "gradingItems": [],
              "judgmentLifecycle": {
                "beforeAll": [],
                "afterAll": []
              }
            }
          ]
        }
      }
      """
