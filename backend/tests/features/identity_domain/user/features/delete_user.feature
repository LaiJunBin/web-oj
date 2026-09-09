Feature: Delete User API
  As a user
  I want to delete a user
  So that I can delete user if I have permission

  Background:
    Given I have users:
      | id  | name               | username | password | roles              |
      | 11  | AdminUser          | user01   | 123456   | admin              |
      | 12  | APIReader          | user02   | 123456   | api_reader         |
      | 13  | Contestant         | user03   | 123456   | contestant         |
      | 14  | AdminAndContestant | user04   | 123456   | admin,contestant   |

  Scenario: Successful delete user with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "DELETE" request to "/api/v1/users/13":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true
      }
      """
    And the user "user03" should not exist
  
  Scenario: Successful delete user with "AdminAndContestant" user
    Given I am logged in as "user04"
    When I send a "DELETE" request to "/api/v1/users/12":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true
      }
      """
    And the user "user02" should not exist

  Scenario: Unauthorized delete user
    When I send a "DELETE" request to "/api/v1/users/11":
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """
    
  Scenario: Forbidden delete user with "APIReader" user
    Given I am logged in as "user02"
    When I send a "DELETE" request to "/api/v1/users/14":
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: Forbidden delete user with "Contestant" user
    Given I am logged in as "user03"
    When I send a "DELETE" request to "/api/v1/users/15":
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: Delete non-existent user
    Given I am logged in as "user01"
    When I send a "DELETE" request to "/api/v1/users/999":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "user not found",
        "details": null
      }
      """

  Scenario: Deleting a user frees its username for re-creation
    Given I am logged in as "user01"
    When I send a "DELETE" request to "/api/v1/users/13":
    Then the response code should be 200
    # Before rename-on-delete this POST failed with a bare 500: the soft-deleted
    # row still occupied the unique username index (which has no partial-index
    # WHERE on ANY database), so the INSERT hit a duplicate-key error that the
    # scoped pre-check could not foresee and nothing mapped to a friendly reply.
    When I send a "POST" request to "/api/v1/users" with the following data:
      """
      {
        "name": "RecreatedUser",
        "username": "user03",
        "password": "654321",
        "roles": [2]
      }
      """
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": 15
      }
      """
    And can login with the new user:
      """
      {
        "username": "user03",
        "password": "654321"
      }
      """
    # Delete the recreated account too: the SECOND dead row of the same base
    # username must get a distinct row-id suffix on the unique index.
    When I send a "DELETE" request to "/api/v1/users/15":
    Then the response code should be 200
    And the user "user03" should not exist
