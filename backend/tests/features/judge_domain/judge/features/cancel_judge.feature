Feature: Cancel Judge
  As an admin
  I want to cancel a queued judgment
  So that I can stop unnecessary judging

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
      | jdg1 | sub1         | Queued | JUDGE     |

  Scenario: Successfully cancel a queued judgment
    Given I am logged in as "admin01"
    When I send a "POST" request to "/api/v1/judgments/jdg1/cancel" with the following data:
      """
      {
        "reason": "Wrong test data"
      }
      """
    Then the response code should be 200
    And the judgment "jdg1" should have status "Cancelled"
    And the judgment "jdg1" should have reason "Wrong test data"

  Scenario: Cancel judgment with default reason
    Given I am logged in as "admin01"
    When I send a "POST" request to "/api/v1/judgments/jdg1/cancel" with the following data:
      """
      {
        "reason": ""
      }
      """
    Then the response code should be 200
    And the judgment "jdg1" should have status "Cancelled"

  Scenario: Cannot cancel an already judged judgment
    Given I am logged in as "admin01"
    And I have judgments:
      | id   | submissionID | status | judgeType |
      | jdg2 | sub1         | Judged | JUDGE     |
    When I send a "POST" request to "/api/v1/judgments/jdg2/cancel" with the following data:
      """
      {
        "reason": "Too late"
      }
      """
    Then the response code should be 400

  Scenario: Cannot cancel a non-existent judgment
    Given I am logged in as "admin01"
    When I send a "POST" request to "/api/v1/judgments/nonexistent/cancel" with the following data:
      """
      {
        "reason": "test"
      }
      """
    Then the response code should be 500

  Scenario: Forbidden cancel for read-only user
    Given I am logged in as "reader01"
    When I send a "POST" request to "/api/v1/judgments/jdg1/cancel" with the following data:
      """
      {
        "reason": "test"
      }
      """
    Then the response code should be 403

  Scenario: Unauthorized cancel without login
    When I send a "POST" request to "/api/v1/judgments/jdg1/cancel" with the following data:
      """
      {
        "reason": "test"
      }
      """
    Then the response code should be 401
