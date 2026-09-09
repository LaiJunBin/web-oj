Feature: Get Rejudge Submission Comparison API
  As an admin
  I want to see what a rejudge changed for one submission down to the testcase
  So that I can decide whether to approve it on more than the total score

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
      | 11 | Test-Project-1 | Test-Project-1 Description | nginx:latest |     | test-nginx:latest |         |              | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have contests:
      | id | name           | description                | isPublic | activateTime        | deactivateTime      | createdAt           | updatedAt           |
      | 11 | Test-Contest-1 | Test-Contest-1 Description | 0        | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have contest project associations:
      | id | contestId | projectId | name           | isPublic | judge priority strategy | activateTime        | deactivateTime      | startTime           | endTime             |
      | 1  | 11        | 11        | Test-Project-1 | 0        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 |
    And I have submissions:
      | id    | userID | contestProjectID |
      | sub-1 | 13     | 1                |

  Scenario: Admin sees the grading item and testcase level change
    Given I have judgments:
      | id      | submissionID | judgeType | status | score | totalScore | createdAt           | gradingItems                                                                                                                                                                                                                    |
      | judge-1 | sub-1        | JUDGE     | Judged | 10    | 30         | 2021-09-01 00:00:00 | [{"id":1,"label":"1","name":"Arithmetic","status":"Failed","score":10,"totalScore":30,"scoringStrategy":"increase","testcases":[{"id":1,"name":"adds","status":"Passed","score":10,"totalScore":10,"parsedScore":10},{"id":2,"name":"overflows","status":"Failed","score":0,"totalScore":20,"parsedScore":5}]}] |
    And I have judgments:
      | id         | submissionID | judgeType | status   | score | totalScore | rejudgeID | createdAt           | gradingItems                                                                                                                                                                                                                    |
      | rejudged-1 | sub-1        | REJUDGE   | Rejudged | 30    | 30         | rejudge1  | 2021-09-02 00:00:00 | [{"id":1,"label":"1","name":"Arithmetic","status":"Passed","score":30,"totalScore":30,"scoringStrategy":"increase","testcases":[{"id":1,"name":"adds","status":"Passed","score":10,"totalScore":10,"parsedScore":10},{"id":2,"name":"overflows","status":"Passed","score":20,"totalScore":20,"parsedScore":20}]}] |
    And I have rejudges:
      | id       | status   | reason               | judgments  |
      | rejudge1 | Finished | fixed overflow check | rejudged-1 |
    And I am logged in as "user01"
    When I send a "GET" request to "/api/v1/rejudges/rejudge1/submissions/sub-1/comparison":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": "sub-1",
          "before": {
            "judgmentId": "judge-1",
            "status": "Judged",
            "score": 10,
            "totalScore": 30,
            "gradingItems": [
              {
                "id": 1,
                "label": "1",
                "name": "Arithmetic",
                "status": "Failed",
                "score": 10,
                "totalScore": 30,
                "scoringStrategy": "increase",
                "testcases": [
                  { "id": 1, "name": "adds", "status": "Passed", "score": 10, "totalScore": 10, "parsedScore": 10 },
                  { "id": 2, "name": "overflows", "status": "Failed", "score": 0, "totalScore": 20, "parsedScore": 5 }
                ]
              }
            ]
          },
          "after": {
            "judgmentId": "rejudged-1",
            "status": "Rejudged",
            "score": 30,
            "totalScore": 30,
            "gradingItems": [
              {
                "id": 1,
                "label": "1",
                "name": "Arithmetic",
                "status": "Passed",
                "score": 30,
                "totalScore": 30,
                "scoringStrategy": "increase",
                "testcases": [
                  { "id": 1, "name": "adds", "status": "Passed", "score": 10, "totalScore": 10, "parsedScore": 10 },
                  { "id": 2, "name": "overflows", "status": "Passed", "score": 20, "totalScore": 20, "parsedScore": 20 }
                ]
              }
            ]
          }
        }
      }
      """

  # A rejudge is usually triggered *because* the grading configuration changed, so
  # an id existing on only one side is the normal case, not corruption. Each side
  # reports the items it was actually scored against and the client pairs by id;
  # nothing is padded or dropped to make the two sides line up.
  Scenario: Grading items added and removed between the two judge runs are both reported
    Given I have judgments:
      | id      | submissionID | judgeType | status | score | totalScore | createdAt           | gradingItems                                                                                                                                          |
      | judge-1 | sub-1        | JUDGE     | Judged | 5     | 20         | 2021-09-01 00:00:00 | [{"id":1,"label":"1","name":"Kept","status":"Failed","score":5,"totalScore":10,"scoringStrategy":"increase","testcases":[]},{"id":2,"label":"2","name":"Removed","status":"Failed","score":0,"totalScore":10,"scoringStrategy":"increase","testcases":[]}] |
    And I have judgments:
      | id         | submissionID | judgeType | status   | score | totalScore | rejudgeID | createdAt           | gradingItems                                                                                                                                          |
      | rejudged-1 | sub-1        | REJUDGE   | Rejudged | 18    | 20         | rejudge1  | 2021-09-02 00:00:00 | [{"id":1,"label":"1","name":"Kept","status":"Passed","score":10,"totalScore":10,"scoringStrategy":"increase","testcases":[]},{"id":3,"label":"3","name":"Added","status":"Passed","score":8,"totalScore":10,"scoringStrategy":"increase","testcases":[]}] |
    And I have rejudges:
      | id       | status   | reason              | judgments  |
      | rejudge1 | Finished | reworked the rubric | rejudged-1 |
    And I am logged in as "user01"
    When I send a "GET" request to "/api/v1/rejudges/rejudge1/submissions/sub-1/comparison":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": "sub-1",
          "before": {
            "judgmentId": "judge-1",
            "status": "Judged",
            "score": 5,
            "totalScore": 20,
            "gradingItems": [
              { "id": 1, "label": "1", "name": "Kept", "status": "Failed", "score": 5, "totalScore": 10, "scoringStrategy": "increase", "testcases": [] },
              { "id": 2, "label": "2", "name": "Removed", "status": "Failed", "score": 0, "totalScore": 10, "scoringStrategy": "increase", "testcases": [] }
            ]
          },
          "after": {
            "judgmentId": "rejudged-1",
            "status": "Rejudged",
            "score": 18,
            "totalScore": 20,
            "gradingItems": [
              { "id": 1, "label": "1", "name": "Kept", "status": "Passed", "score": 10, "totalScore": 10, "scoringStrategy": "increase", "testcases": [] },
              { "id": 3, "label": "3", "name": "Added", "status": "Passed", "score": 8, "totalScore": 10, "scoringStrategy": "increase", "testcases": [] }
            ]
          }
        }
      }
      """

  Scenario: A submission with no previously adopted judgment has no baseline
    Given I have judgments:
      | id      | submissionID | judgeType | status | score | totalScore | createdAt           |
      | judge-1 | sub-1        | JUDGE     | Error  | 0     | 30         | 2021-09-01 00:00:00 |
    And I have judgments:
      | id         | submissionID | judgeType | status   | score | totalScore | rejudgeID | createdAt           | gradingItems                                                                                                                       |
      | rejudged-1 | sub-1        | REJUDGE   | Rejudged | 30    | 30         | rejudge1  | 2021-09-02 00:00:00 | [{"id":1,"label":"1","name":"Arithmetic","status":"Passed","score":30,"totalScore":30,"scoringStrategy":"increase","testcases":[]}] |
    And I have rejudges:
      | id       | status   | reason              | judgments  |
      | rejudge1 | Finished | first judge errored | rejudged-1 |
    And I am logged in as "user01"
    When I send a "GET" request to "/api/v1/rejudges/rejudge1/submissions/sub-1/comparison":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": "sub-1",
          "before": null,
          "after": {
            "judgmentId": "rejudged-1",
            "status": "Rejudged",
            "score": 30,
            "totalScore": 30,
            "gradingItems": [
              { "id": 1, "label": "1", "name": "Arithmetic", "status": "Passed", "score": 30, "totalScore": 30, "scoringStrategy": "increase", "testcases": [] }
            ]
          }
        }
      }
      """

  Scenario: An approved earlier rejudge is the baseline
    Given I have judgments:
      | id          | submissionID | judgeType | status   | score | totalScore | rejudgeID | createdAt           | gradingItems                                                                                                                   |
      | judge-1     | sub-1        | JUDGE     | Judged   | 10    | 30         |           | 2021-09-01 00:00:00 | [{"id":1,"label":"1","name":"Arithmetic","status":"Failed","score":10,"totalScore":30,"scoringStrategy":"increase","testcases":[]}] |
      | rejudged-1  | sub-1        | REJUDGE   | Judged   | 20    | 30         | rejudge1  | 2021-09-02 00:00:00 | [{"id":1,"label":"1","name":"Arithmetic","status":"Failed","score":20,"totalScore":30,"scoringStrategy":"increase","testcases":[]}] |
      | rejudged-2  | sub-1        | REJUDGE   | Rejudged | 30    | 30         | rejudge2  | 2021-09-03 00:00:00 | [{"id":1,"label":"1","name":"Arithmetic","status":"Passed","score":30,"totalScore":30,"scoringStrategy":"increase","testcases":[]}] |
    And I have rejudges:
      | id       | status   | reason      | judgments  |
      | rejudge1 | Approved | first pass  | rejudged-1 |
      | rejudge2 | Finished | second pass | rejudged-2 |
    And I am logged in as "user01"
    When I send a "GET" request to "/api/v1/rejudges/rejudge2/submissions/sub-1/comparison":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "id": "sub-1",
          "before": {
            "judgmentId": "rejudged-1",
            "status": "Judged",
            "score": 20,
            "totalScore": 30,
            "gradingItems": [
              { "id": 1, "label": "1", "name": "Arithmetic", "status": "Failed", "score": 20, "totalScore": 30, "scoringStrategy": "increase", "testcases": [] }
            ]
          },
          "after": {
            "judgmentId": "rejudged-2",
            "status": "Rejudged",
            "score": 30,
            "totalScore": 30,
            "gradingItems": [
              { "id": 1, "label": "1", "name": "Arithmetic", "status": "Passed", "score": 30, "totalScore": 30, "scoringStrategy": "increase", "testcases": [] }
            ]
          }
        }
      }
      """

  Scenario: Rejudge not found
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/rejudges/nonexistent/submissions/sub-1/comparison":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "rejudge not found",
        "details": null
      }
      """

  Scenario: Submission was not affected by this rejudge
    Given I have judgments:
      | id         | submissionID | judgeType | status   | score | totalScore | rejudgeID | createdAt           |
      | rejudged-1 | sub-1        | REJUDGE   | Rejudged | 30    | 30         | rejudge1  | 2021-09-02 00:00:00 |
    And I have rejudges:
      | id       | status   | reason      | judgments  |
      | rejudge1 | Finished | some reason | rejudged-1 |
    And I am logged in as "user01"
    When I send a "GET" request to "/api/v1/rejudges/rejudge1/submissions/sub-other/comparison":
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
    When I send a "GET" request to "/api/v1/rejudges/rejudge1/submissions/sub-1/comparison":
    Then the response code should be 401

  Scenario: Forbidden for contestant role
    Given I am logged in as "user03"
    When I send a "GET" request to "/api/v1/rejudges/rejudge1/submissions/sub-1/comparison":
    Then the response code should be 403
