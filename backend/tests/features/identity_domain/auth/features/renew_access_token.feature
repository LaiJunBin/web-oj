Feature: Automatic access token renewal
  As a logged in user whose access token has expired
  I want the middleware to renew it from my refresh token
  So that my session and my SSE streams survive past the access token's lifetime

  # This covers the authorization middleware's automatic renewal, which is a
  # different path from the explicit POST /auth/refresh-token endpoint in
  # refresh_token.feature and had no coverage at all. It is the path that
  # broke in production: contestants were logged out and their scoreboards
  # froze about an hour into a contest.

  Background:
    Given I have users:
      |  id |   name  | username | password |              roles               |
      |  1  |  User01 | user01   | 123456   |  admin, contestant   |
    Given I have refresh tokens for the following users:
      |  refresh_token     | user_id | ttl  |
      |  token1            |   1     |  30  |
      |  expiredtoken      |   1     |  -1  |

  Scenario: An expired access token is renewed from a valid refresh token
    Given I am logged in as "user01" with an expired access token
    When I send a "GET" request to "/api/v1/auth/profile" with the following cookies:
      | name           | value  |
      | refresh_token  | token1 |
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
    # Renewal is only useful if the browser is handed the new token. In
    # production the renewed cookie was issued on an SSE response behind the
    # BFF proxy and never made it back.
    And the response should set cookie "access_token"

  # The invariant the production fix rests on. A contest page holds three SSE
  # connections open at once, all carrying the same refresh cookie, so an
  # expired access token makes all three renew at the same instant. If
  # renewal rotated — as the explicit endpoint does — the first RENAME would
  # destroy the refresh token and every later use of it would 401. Reusing
  # the same refresh token is the sequential stand-in for that race: it fails
  # for exactly the same reason, without depending on timing.
  Scenario: The same refresh token can renew more than once
    Given I am logged in as "user01" with an expired access token
    When I send a "GET" request to "/api/v1/auth/profile" with the following cookies:
      | name           | value  |
      | refresh_token  | token1 |
    Then the response code should be 200
    And the response should set cookie "access_token"
    Given I am logged in as "user01" with an expired access token
    When I send a "GET" request to "/api/v1/auth/profile" with the following cookies:
      | name           | value  |
      | refresh_token  | token1 |
    Then the response code should be 200
    And the response should set cookie "access_token"
    Given I am logged in as "user01" with an expired access token
    When I send a "GET" request to "/api/v1/auth/profile" with the following cookies:
      | name           | value  |
      | refresh_token  | token1 |
    Then the response code should be 200

  Scenario: An expired access token with no refresh token is rejected
    Given I am logged in as "user01" with an expired access token
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

  Scenario: An expired access token with an unknown refresh token is rejected
    Given I am logged in as "user01" with an expired access token
    When I send a "GET" request to "/api/v1/auth/profile" with the following cookies:
      | name           | value        |
      | refresh_token  | notatoken    |
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """

  Scenario: An expired access token with an expired refresh token is rejected
    Given I am logged in as "user01" with an expired access token
    When I send a "GET" request to "/api/v1/auth/profile" with the following cookies:
      | name           | value         |
      | refresh_token  | expiredtoken  |
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """
