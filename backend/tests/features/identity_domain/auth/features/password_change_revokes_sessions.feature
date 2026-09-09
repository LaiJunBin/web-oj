Feature: Password change revokes existing sessions
  As an administrator resetting a compromised account's password
  I want that account's existing sessions to stop working
  So that whoever holds the old credentials loses access

  # A refresh token outlives the access token by a week, so without this the
  # password reset does not actually cut anybody off — the attacker just keeps
  # renewing. Revocation has to be scoped to the account whose password moved:
  # revoking too widely logs out the whole contest.

  Background:
    Given I have users:
      |  id  | name      | username | password |     roles     |
      |  11  | AdminUser | user01   | 123456   | admin         |
      |  12  | Player    | user02   | 123456   | contestant    |
      |  13  | Player2   | user03   | 123456   | contestant    |
    And I have refresh tokens for the following users:
      |  refresh_token   | user_id | ttl  |
      |  player-session  |   12    | 300  |
      |  player-laptop   |   12    | 300  |
      |  other-session   |   13    | 300  |
    And I am logged in as "user01"

  Scenario: Changing a user's password revokes that user's refresh token
    When I send a "PATCH" request to "/api/v1/users/12" with the following data:
      """
      {
        "password": "a-new-password"
      }
      """
    Then the response code should be 200
    When I send a "POST" request to "/api/v1/auth/refresh-token" with the following cookies:
      | name          | value           |
      | refresh_token | player-session  |
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "invalid token",
        "details": null
      }
      """

  # One reset has to clear every device the account is signed in on, or the
  # attacker simply continues from the session that was missed.
  Scenario: Changing a user's password revokes all of that user's refresh tokens
    When I send a "PATCH" request to "/api/v1/users/12" with the following data:
      """
      {
        "password": "a-new-password"
      }
      """
    Then the response code should be 200
    When I send a "POST" request to "/api/v1/auth/refresh-token" with the following cookies:
      | name          | value          |
      | refresh_token | player-laptop  |
    Then the response code should be 401

  # The blast radius must stop at the account being reset.
  Scenario: Changing a user's password leaves other users signed in
    When I send a "PATCH" request to "/api/v1/users/12" with the following data:
      """
      {
        "password": "a-new-password"
      }
      """
    Then the response code should be 200
    When I send a "POST" request to "/api/v1/auth/refresh-token" with the following cookies:
      | name          | value          |
      | refresh_token | other-session  |
    Then the response code should be 200

  # Renaming somebody is not a security event; logging them out would make
  # routine admin edits disruptive during a contest.
  Scenario: Updating a user without changing the password keeps their session
    When I send a "PATCH" request to "/api/v1/users/12" with the following data:
      """
      {
        "name": "Player Renamed"
      }
      """
    Then the response code should be 200
    When I send a "POST" request to "/api/v1/auth/refresh-token" with the following cookies:
      | name          | value           |
      | refresh_token | player-session  |
    Then the response code should be 200
