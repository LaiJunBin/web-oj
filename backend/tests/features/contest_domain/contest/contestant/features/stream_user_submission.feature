Feature: Stream User Submission API
  As a contestant
  I want live updates for my submission
  So that I can watch it being judged without refreshing

  # The stream must apply the same visibility rule as the submission detail:
  # only testcases marked as samples may expose their run output. The frame is
  # triggered deterministically by an admin cancelling the queued judgment —
  # that always publishes one judge-updated event carrying the judgment's full
  # grading-item snapshot, with none of the timing of a real judging run.
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
      | id | name     | description | imageTag     | env | testImageTag      | testEnv | resultParser | createdAt           | updatedAt           |
      | 1  | Project1 | desc        | nginx:latest |     | test-nginx:latest |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have grading-items:
      | id | projectId | name | description | score | order | scoringStrategy | outputFormat | resultParser | label | createdAt | updatedAt |
      | 1 | 1 | Task 1 | desc | 20 | 1 | increase | text |  |  | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And I have testcases:
      | id | gradingItemId | name        | command | score | isSample | createdAt           | updatedAt           |
      | 1  | 1             | sample-case | true    | 10    | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 2  | 1             | hidden-case | true    | 10    | 0        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And I have contests:
      | id | name    | description | isPublic | activateTime        | deactivateTime      | createdAt           | updatedAt           |
      | 1  | Contest | desc        | 1        | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have contest project associations:
      | id | contestId | projectId | name     | isPublic | judge priority strategy | activateTime        | deactivateTime      | startTime           | endTime             |
      | 1  | 1         | 1         | Project1 | 1        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 |
    And I have contest participants:
      | contestId | userId | groupId |
      | 1         | 13     |         |
    And I have submissions:
      | id   | userID | contestProjectID | ignored |
      | sub1 | 13     | 1                | false   |
    And I have judgments:
      | id   | submissionID | contestProjectID | status | judgeType | score | totalScore | createdAt            | gradingItems                                                                                                                                                                                                                                                                                             |
      | jdg1 | sub1         | 1                | Queued | JUDGE     | 10    | 20         | 2024-01-01T00:00:00Z | [{"id":1,"label":"1","name":"Task 1","status":"Judging","score":10,"totalScore":20,"scoringStrategy":"increase","testcases":[{"id":1,"name":"sample-case","status":"Passed","score":10,"totalScore":10,"stdout":"sample-visible-output"},{"id":2,"name":"hidden-case","status":"Failed","score":0,"totalScore":10,"stdout":"hidden-secret-output","stderr":"hidden-secret-stderr"}]}] |
      | jdg2 | sub1         | 1                | Queued | JUDGE     | 10    | 20         | 2024-01-01T00:10:00Z | [{"id":1,"label":"1","name":"Task 1","status":"Judging","score":10,"totalScore":20,"scoringStrategy":"increase","testcases":[{"id":1,"name":"sample-case","status":"Passed","score":10,"totalScore":10,"stdout":"stale-sample-output"},{"id":2,"name":"hidden-case","status":"Failed","score":0,"totalScore":10,"stdout":"now-visible-output"}]}] |

  Scenario: Hidden testcase output is absent from the contestant stream
    Given I am logged in as "user03"
    And when start streaming logs via request to "/api/v1/user/stream-submissions/sub1":
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/judgments/jdg1/cancel" with the following data:
      """
      {
        "reason": "flush the snapshot to the stream"
      }
      """
    Then the response code should be 200
    And streaming logs should have "sample-visible-output" message within 10 seconds
    And streaming logs should not have "hidden-secret-output" message
    And streaming logs should not have "hidden-secret-stderr" message

  # Visibility can change mid-contest; an open stream follows within the
  # resolver's 5-second TTL. Both directions in one pass: testcase 1 flips
  # sample -> hidden, testcase 2 flips hidden -> sample, and the two seeded
  # judgments carry DISTINCT output strings because the negative assertion
  # scans everything the stream delivered since it opened.
  Scenario: An open stream follows a mid-contest visibility change
    Given I am logged in as "user03"
    And when start streaming logs via request to "/api/v1/user/stream-submissions/sub1":
    Given I am logged in as "user01"
    When I send a "PATCH" request to "/api/v1/testcases/1" with the following data:
      """
      {
        "isSample": false
      }
      """
    Then the response code should be 200
    When I send a "PATCH" request to "/api/v1/testcases/2" with the following data:
      """
      {
        "isSample": true
      }
      """
    Then the response code should be 200
    Given I wait for 6 seconds
    When I send a "POST" request to "/api/v1/judgments/jdg2/cancel" with the following data:
      """
      {
        "reason": "flush after visibility change"
      }
      """
    Then the response code should be 200
    And streaming logs should have "now-visible-output" message within 10 seconds
    And streaming logs should not have "stale-sample-output" message
