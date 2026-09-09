Feature: Access token renewal end to end with a short lifetime
  As an operator
  I want a real login's token to expire and be renewed automatically
  So that the configured lifetime, the login path and the renewal path are proven to work together

  # Companion to renew_access_token.feature, which mints an already-expired
  # token to stay deterministic. This one proves the whole chain instead:
  # config supplies the lifetime, a real login issues a token with it, the
  # token genuinely expires, and the middleware renews it.
  #
  # It only runs when the access lifetime is actually short — see
  # TestAuthShortTTL, which skips otherwise rather than let this pass
  # vacuously against the normal hour-long lifetime. Run it with
  # `make test:feature:short-ttl`, on its own: a two-second access token
  # would make every other scenario in the suite flaky.

  Background:
    Given I have users:
      |  id |   name  | username | password |              roles               |
      |  1  |  User01 | user01   | 123456   |  admin, contestant   |

  Scenario: A real login's access token expires and is renewed automatically
    Given I log in via the API as "user01" with password "123456"
    When I send a "GET" request to "/api/v1/auth/profile":
    Then the response code should be 200
    # Longer than the configured lifetime, so the access token issued above
    # is genuinely past its exp by the next request.
    Given I wait for 4 seconds
    When I send a "GET" request to "/api/v1/auth/profile":
    Then the response code should be 200
    And the response should set cookie "access_token"
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
    # The refresh token issued at login is still the one the browser holds,
    # so renewal keeps working for as long as it lives.
    Given I wait for 4 seconds
    When I send a "GET" request to "/api/v1/auth/profile":
    Then the response code should be 200
