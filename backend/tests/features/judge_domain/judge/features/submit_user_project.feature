Feature: Submit User Project
  As a contestant
  I want to submit my project
  So that it gets judged automatically

  Background:
    Given I have users:
      | id | name       | username | password | roles      |
      | 11 | AdminUser  | admin01  | 123456   | admin      |
      | 12 | Reader     | reader01 | 123456   | api_reader |
      | 13 | Contestant | user01   | 123456   | contestant |
    And I have images:
      | id | tag               | mount_path    | description  | is_ready | createdAt           | updatedAt           |
      | 1  | nginx:latest      | /var/www/html | desc1        | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 2  | test-nginx:latest | /var/www/html | desc2        | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And I have projects:
      | id | name     | description | imageTag     | env | testImageTag      | testEnv | resultParser | createdAt           | updatedAt           |
      | 1  | Project1 | desc        | nginx:latest |     | test-nginx:latest |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have grading-items:
      | id | projectId | name | description | score | order | scoringStrategy | outputFormat | resultParser | label | createdAt | updatedAt |
      | 1 | 1 | Task 1 | desc | 50 | 1 | increase | text |  |  | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 2 | 1 | Task 2 | desc | 50 | 2 | increase | text |  |  | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And I have testcases:
      | id | gradingItemId | name    | command     | score | isSample | createdAt           | updatedAt           |
      | 1  | 1             | Test 01 | echo hello  | 25    | false    | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 2  | 1             | Test 02 | echo world  | 25    | false    | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 3  | 2             | Test 03 | echo foo    | 50    | false    | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And I have contests:
      | id | name    | description | isPublic | activateTime        | deactivateTime      | createdAt           | updatedAt           |
      | 1  | Contest | desc        | 1        | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I have contest project associations:
      | id | contestId | projectId | name     | isPublic | judge priority strategy | activateTime        | deactivateTime      | startTime           | endTime             |
      | 1  | 1         | 1         | Project1 | 1        | default                 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 | 2021-09-01 00:00:00 | 2099-09-30 23:59:59 |
    And I have contest participants:
      | contestId | userId | groupId |
      | 1         | 13     |         |
  Scenario: Successfully submit a project and trigger judging
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/user/contest-projects/1/submit" with the following formdata:
      | key     | value           |
      | project | {nginx project} |
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {}
      }
      """
    And a submission should exist for contest project 1
    And a judgment should exist for contest project 1
    And the judgment for contest project 1 should reach status "Judged"
    And the judgment score for contest project 1 should be 100

  Scenario: Unauthorized submit without login
    When I send a "POST" request to "/api/v1/user/contest-projects/1/submit" with the following formdata:
      | key     | value           |
      | project | {nginx project} |
    Then the response code should be 401

  Scenario: Forbidden submit for read-only user
    Given I am logged in as "reader01"
    When I send a "POST" request to "/api/v1/user/contest-projects/1/submit" with the following formdata:
      | key     | value           |
      | project | {nginx project} |
    Then the response code should be 403

  Scenario: Submit to non-existent contest project
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/user/contest-projects/999/submit" with the following formdata:
      | key     | value           |
      | project | {nginx project} |
    Then the response code should be 404

  # A before_* lifecycle that exits non-zero used to be logged as "finished"
  # and judging carried on against an environment that was never set up; the
  # contestant got an inexplicable failing score. It now aborts the judgment as
  # ERROR with a neutral, contestant-safe reason (the worker log carries the
  # exit code and stderr for the operator). The fake runtime treats the literal
  # command `exit 1` as a failing step.
  Scenario: A failing before_all lifecycle aborts the judgment with a neutral reason
    Given I have lifecycles:
      | id | projectId | name        | description           | type       | runsOn | command | createdAt           | updatedAt           |
      | 1  | 1         | setup-fails | simulated setup error | before_all | test   | exit 1  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    And I am logged in as "user01"
    When I send a "POST" request to "/api/v1/user/contest-projects/1/submit" with the following formdata:
      | key     | value           |
      | project | {nginx project} |
    Then the response code should be 200
    And a submission should exist for contest project 1
    And the judgment for contest project 1 should reach status "Error"
    And the judgment for contest project 1 should have reason "評測前置步驟失敗，請聯絡現場工作人員"
    And every testcase of the judgment for contest project 1 should have status "Cancelled"
