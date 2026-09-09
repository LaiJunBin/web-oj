Feature: Logout API
  As a logged in user
  I want to logout
  So that I can invalidate my refresh token

  Background:
    Given I have refresh tokens for the following users:
      |  refresh_token     | user_id | ttl  |
      |  token1            |   1   |  30  |
      |  token2            |   2   |  -1  |
      |  token3            |   3   |  30  |

  Scenario: Successful logout
    When I send a "POST" request to "/api/v1/auth/logout" with the following cookies:
      | name          | value  |
      | refresh_token  | token1 |
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true
      }
      """

  Scenario: Unsuccessful logout with invalid refresh token
    When I send a "POST" request to "/api/v1/auth/logout" with the following cookies:
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

  
    
