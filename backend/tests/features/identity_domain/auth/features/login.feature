Feature: Login API
  As a user
  I want to log in through the API
  So that I can access my account

  Background:
    Given I have users:
      |  id  |   name  | username | password |    roles       |
      |  11  |  User01 | user01   | 123456   |    admin       |
      |  12  |  User02 | user02   | 654321   |    contestant  |

  Scenario: Successful login
    When I send a "POST" request to "/api/v1/auth/login" with the following data:
    """
    {
      "username": "user01",
      "password": "123456"
    }
    """
    Then the response code should be 200
    # Tokens are delivered ONLY as HttpOnly cookies now, never in the body, so a
    # same-origin XSS cannot read them. The body keeps the fields but empty; the
    # cookies below are the real delivery and what the client relies on.
    And the response should match json:
      """
      {
        "success": true,
        "data": {
          "accessToken": "",
          "refreshToken": ""
        }
      }
      """
    And set cookie "access_token"
    And set cookie "refresh_token"

  Scenario: Unsuccessful login with wrong password
    When I send a "POST" request to "/api/v1/auth/login" with the following data:
    """
    {
      "username": "user01",
      "password": "wrongpassword"
    }
    """
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "invalid credentials",
        "details": null
      }
      """

  Scenario: Unsuccessful login with non-existent user
    When I send a "POST" request to "/api/v1/auth/login" with the following data:
    """
    {
      "username": "user03",
      "password": "123456"
    }
    """
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "invalid credentials",
        "details": null
      }
      """

  Scenario: Login attempt with empty credentials
    When I send a "POST" request to "/api/v1/auth/login" with the following data:
    """
    {
      "username": "",
      "password": ""
    }
    """
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "bad request",
        "details": null
      }
      """

  Scenario: Login attempt with missing credentials
    When I send a "POST" request to "/api/v1/auth/login" with the following data:
    """
    {
      "username": "user01"
    }
    """
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "bad request",
        "details": null
      }
      """
