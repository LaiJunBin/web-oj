Feature: Create Project Lifecycle API
  As a user
  I want to create a project lifecycle
  So that I can create a project lifecycle if I have permission

  Background:
    Given I have users:
      | id  | name               | username | password | roles              |
      | 11  | AdminUser          | user01   | 123456   | admin              |
      | 12  | APIReader          | user02   | 123456   | api_reader         |
      | 13  | Contestant         | user03   | 123456   | contestant         |
      | 14  | AdminAndContestant | user04   | 123456   | admin,contestant   |
    And I have images:
      | id | tag                | mount_path    | description  | is_ready | createdAt           | updatedAt           |
      | 1  | nginx:latest       | /var/www/html | description1 | 0        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 2  | mysql:8.0.26       | /var/www/html | description2 | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 3  | test-nginx:latest  | /var/www/html | description3 | 0        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 4  | test-mysql:8.0.26  | /var/www/html | description4 | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 5  | test-nginx:test-2  | /var/www/html | description5 | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And I have projects:
      | id | name           | description               | imageTag       | env | testImageTag      | testEnv | resultParser | createdAt           | updatedAt           |
      | 11 | Test-Project-1 | Test-Project-1 Description| nginx:latest   | a=b | test-nginx:latest |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 12 | Test-Project-2 | Test-Project-2 Description| mysql:8.0.26   | c=d | test-mysql:8.0.26 |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 13 | Test-Project-3 | Test-Project-3 Description| nginx:latest   |     | test-nginx:test-2 |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
  Scenario: Successful create lifecycle with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/projects/11/lifecycles" with the following data:
      """
        {
            "type": "before_all",
            "name": "lc-test",
            "command": "echo 123",
            "runsOn": "project"
        }
      """
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": 1
      }
      """
  
  Scenario: Successful create lifecycle with "AdminAndContestant" user
    Given I am logged in as "user04"
    When I send a "POST" request to "/api/v1/projects/11/lifecycles" with the following data:
      """
        {
            "type": "after_each",
            "name": "lc-test",
            "command": "echo 123",
            "runsOn": "test"
        }
      """
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": 1
      }
      """

  Scenario: Successful create lifecycle and "AdminUser" user
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/projects/11/lifecycles" with the following data:
      """
         {
            "type": "after_all",
            "name": "lc-test",
            "command": "echo 123",
            "runsOn": "test"
        }
      """
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": 1
      }
      """

  Scenario: Failed create lifecycle with invalid type
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/projects/11/lifecycles" with the following data:
      """
         {
            "type": "invalid_type",
            "name": "lc-test",
            "command": "echo 123",
            "runsOn": "test"
        }
      """
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "invalid type, must be one of before_all, before_each, after_each, after_all",
        "details": null
      }
      """

  Scenario: Failed create lifecycle with invalid execution environment
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/projects/11/lifecycles" with the following data:
      """
         {
            "type": "before_all",
            "name": "lc-test",
            "command": "echo 123",
            "runsOn": "invalid_env"
        }
      """
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "invalid execution environment, must be one of project, test",
        "details": null
      }
      """

  Scenario: Failed create lifecycle with missing name
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/projects/11/lifecycles" with the following data:
      """
         {
            "type": "before_all",
            "name": "",
            "command": "echo 123",
            "runsOn": "test"
        }
      """
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "name is required",
        "details": null
      }
      """

  Scenario: Unauthorized create lifecycle
    When I send a "POST" request to "/api/v1/projects/11/lifecycles" with the following data:
      """
         {
            "type": "after_all",
            "name": "lc-test",
            "command": "echo 123",
            "runsOn": "test"
        }
      """
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """
    
  Scenario: Forbidden create lifecycle with "APIReader" user
    Given I am logged in as "user02"
    When I send a "POST" request to "/api/v1/projects/11/lifecycles" with the following data:
      """
         {
            "type": "after_all",
            "name": "lc-test",
            "command": "echo 123",
            "runsOn": "test"
        }
      """
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: Forbidden create lifecycle with "Contestant" user
    Given I am logged in as "user03"
    When I send a "POST" request to "/api/v1/projects/11/lifecycles" with the following data:
      """
         {
            "type": "after_all",
            "name": "lc-test",
            "command": "echo 123",
            "runsOn": "test"
        }
      """
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: Failed create lifecycle with project not found
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/projects/111/lifecycles" with the following data:
      """
         {
            "type": "after_all",
            "name": "lc-test",
            "command": "echo 123",
            "runsOn": "test"
        }
      """
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "project not found",
        "details": null
      }
      """
