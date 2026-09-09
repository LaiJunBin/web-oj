Feature: Import Users API
  As a user
  I want to import users with a json file
  So that I can import users with a json file if I have permission

  Background:
    Given I have users:
      | id  | name               | username | password | roles              |
      | 11  | AdminUser          | user01   | 123456   | admin              |
      | 12  | APIReader          | user02   | 123456   | api_reader         |
      | 13  | Contestant         | user03   | 123456   | contestant         |
      | 14  | AdminAndContestant | user04   | 123456   | admin,contestant   |

  Scenario: Successful import users with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/users/import" with the following formdata:
      | key            | value                |
      | users          | {userdata_01.json}   |
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": 3
      }
      """
  
  Scenario: Failed import users with "AdminAndContestant" user
    Given I am logged in as "user04"
    When I send a "POST" request to "/api/v1/users/import" with the following formdata:
      | key            | value                |
      | users          | {userdata_03.json}   |
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "importing users error",
        "details": ["failed in row: {\"name\":\"選手07\",\"username\":\"user07\",\"password\":\"1234\",\"roles\":[\"unknown\"]}"]
      }
      """

  Scenario: Failed import users
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/users/import" with the following formdata:
      | key            | value                |
      | users          | {userdata_04.json}   |
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "invalid json",
        "details": null
      }
      """

  Scenario: Failed import users with existing username
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/users/import" with the following formdata:
      | key            | value                |
      | users          | {userdata_05.json}   |
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "importing users error",
        "details": ["user01 already exists", "failed in row: {\"name\":\"NewUser01\",\"username\":\"user01\",\"password\":\"1234\",\"roles\":[\"contestant\"]}"]
      }
      """
 
  Scenario: Unauthorized import users
    When I send a "POST" request to "/api/v1/users/import" with the following formdata:
      | key            | value                |
      | users          | {userdata_01.json}   |
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """
    
  Scenario: Forbidden import users with "APIReader" user
    Given I am logged in as "user02"
    When I send a "POST" request to "/api/v1/users/import" with the following formdata:
      | key            | value                |
      | users          | {userdata_01.json}   |
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: Forbidden import users with "Contestant" user
    Given I am logged in as "user03"
    When I send a "POST" request to "/api/v1/users/import" with the following formdata:
      | key            | value                |
      | users          | {userdata_01.json}   |
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """



