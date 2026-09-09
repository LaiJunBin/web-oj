Feature: Refresh Token API
  As a logged in user
  I want to refresh my access token
  So that I can continue to access protected resources

  Background:
    Given I have users:
      |  id |   name  | username | password |              roles               |
      |  1  |  User01 | user01   | 123456   |  admin, contestant   |
    Given I have refresh tokens for the following users:
      |  refresh_token     | user_id | ttl  |
      |  token1            |   1   |  30  |
      |  token2            |   2   |  -1  |
      |  token3            |   3   |  30  |
    Given I am logged in as "user01"

  Scenario: Successful refresh token
    When I send a "POST" request to "/api/v1/auth/refresh-token" with the following cookies:
      | name          | value  |
      | refresh_token  | token1 |
    Then the response code should be 200
    # Rotated tokens are delivered as HttpOnly cookies, not in the body (empty
    # here); the cookies below are the real delivery.
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

  Scenario: Unsuccessful refresh token with invalid refresh token
    When I send a "POST" request to "/api/v1/auth/refresh-token" with the following cookies:
      | name          | value  |
      | refresh_token  | invalidtoken |
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "invalid token",
        "details": null
      }
      """

  Scenario: Unsuccessful refresh token with expired refresh token
    When I send a "POST" request to "/api/v1/auth/refresh-token" with the following cookies:
      | name          | value  |
      | refresh_token  | token2 |
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "invalid token",
        "details": null
      }
      """

  Scenario: Unsuccessful refresh token with non-existent user
    When I send a "POST" request to "/api/v1/auth/refresh-token" with the following cookies:
      | name          | value  |
      | refresh_token  | token3 |
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "invalid credentials",
        "details": null
      }
      """

  Scenario: Unsuccessful refresh token with non-existent refresh token
    When I send a "POST" request to "/api/v1/auth/refresh-token" with the following cookies:
      | name          | value  |
      | refresh_token  | token4 |
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "invalid token",
        "details": null
      }
      """
    
