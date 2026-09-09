Feature: Get Profile API
  As a user
  I want to get my profile
  So that I can see my profile information

  Background:
    Given I have users:
      |  id  | name   | username | password | roles                         |
      |  11  | User01 | user01   | 123456   | admin, contestant |
      |  12  | User02 | user02   | 654321   | api_reader                    |

  Scenario: Successful get user01 profile
    Given I am logged in as "user01"
    When I send a "GET" request to "/api/v1/auth/profile":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "name": "User01",
          "username": "user01",
          "roles": [
            "admin",
            "contestant"
          ]
        }
      }
      """

  Scenario: Successful get user02 profile
    Given I am logged in as "user02"
    When I send a "GET" request to "/api/v1/auth/profile":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "name": "User02",
          "username": "user02",
          "roles": [
            "api_reader"
          ]
        }
      }
      """

  Scenario: Unsuccessful login without token
    When I send a "GET" request to "/api/v1/auth/profile":
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """
