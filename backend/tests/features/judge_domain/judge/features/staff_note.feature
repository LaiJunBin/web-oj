Feature: Judgment Staff Note
  As an admin
  I want to attach a note to a judgment that the contestant can read
  So that I can tell them what to fix without going through a clarification

  Background:
    Given I have users:
      | id | name      | username | password | roles      |
      | 11 | AdminUser | admin01  | 123456   | admin      |
      | 12 | Reader    | reader01 | 123456   | api_reader |
    And I have submissions:
      | id   | userID | contestProjectID |
      | sub1 | 11     | 1                |
    And I have judgments:
      | id   | submissionID | status | judgeType |
      | jdg1 | sub1         | Judged | JUDGE     |

  Scenario: Set a staff note on a finished judgment
    Given I am logged in as "admin01"
    When I send a "PUT" request to "/api/v1/judgments/jdg1/staff-note" with the following data:
      """
      {
        "staffNote": "第 3 題輸出多了一個換行，請對照範例"
      }
      """
    Then the response code should be 200
    And the judgment "jdg1" should have staff note "第 3 題輸出多了一個換行，請對照範例"
    And the judgment "jdg1" should have status "Judged"

  Scenario: A note may be written while the judgment is still queued
    Given I am logged in as "admin01"
    And I have judgments:
      | id   | submissionID | status | judgeType |
      | jdg2 | sub1         | Queued | JUDGE     |
    When I send a "PUT" request to "/api/v1/judgments/jdg2/staff-note" with the following data:
      """
      {
        "staffNote": "評測排隊中，請耐心等候"
      }
      """
    Then the response code should be 200
    And the judgment "jdg2" should have staff note "評測排隊中，請耐心等候"
    And the judgment "jdg2" should have status "Queued"

  Scenario: A later note overwrites the earlier one
    Given I am logged in as "admin01"
    When I send a "PUT" request to "/api/v1/judgments/jdg1/staff-note" with the following data:
      """
      {
        "staffNote": "第一版附註"
      }
      """
    Then the response code should be 200
    And the judgment "jdg1" should have staff note "第一版附註"
    When I send a "PUT" request to "/api/v1/judgments/jdg1/staff-note" with the following data:
      """
      {
        "staffNote": "第二版附註"
      }
      """
    Then the response code should be 200
    And the judgment "jdg1" should have staff note "第二版附註"

  Scenario: An empty string clears the note
    Given I am logged in as "admin01"
    When I send a "PUT" request to "/api/v1/judgments/jdg1/staff-note" with the following data:
      """
      {
        "staffNote": "待撤回的附註"
      }
      """
    Then the response code should be 200
    And the judgment "jdg1" should have staff note "待撤回的附註"
    When I send a "PUT" request to "/api/v1/judgments/jdg1/staff-note" with the following data:
      """
      {
        "staffNote": ""
      }
      """
    Then the response code should be 200
    And the judgment "jdg1" should have no staff note

  Scenario: The note coexists with a cancellation reason
    Given I am logged in as "admin01"
    And I have judgments:
      | id   | submissionID | status | judgeType |
      | jdg3 | sub1         | Queued | JUDGE     |
    When I send a "POST" request to "/api/v1/judgments/jdg3/cancel" with the following data:
      """
      {
        "reason": "測資有誤，將重新評測"
      }
      """
    Then the response code should be 200
    And the judgment "jdg3" should have status "Cancelled"
    When I send a "PUT" request to "/api/v1/judgments/jdg3/staff-note" with the following data:
      """
      {
        "staffNote": "重測完成前請勿再提交"
      }
      """
    Then the response code should be 200
    And the judgment "jdg3" should have staff note "重測完成前請勿再提交"
    And the judgment "jdg3" should have reason "測資有誤，將重新評測"
    And the judgment "jdg3" should have status "Cancelled"

  Scenario: Cannot note a non-existent judgment
    Given I am logged in as "admin01"
    When I send a "PUT" request to "/api/v1/judgments/nonexistent/staff-note" with the following data:
      """
      {
        "staffNote": "test"
      }
      """
    Then the response code should be 404

  Scenario: Forbidden for read-only user
    Given I am logged in as "reader01"
    When I send a "PUT" request to "/api/v1/judgments/jdg1/staff-note" with the following data:
      """
      {
        "staffNote": "test"
      }
      """
    Then the response code should be 403

  Scenario: Unauthorized without login
    When I send a "PUT" request to "/api/v1/judgments/jdg1/staff-note" with the following data:
      """
      {
        "staffNote": "test"
      }
      """
    Then the response code should be 401
