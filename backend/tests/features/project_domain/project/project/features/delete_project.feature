Feature: Delete Project API
  As a user
  I want to delete a project
  So that I can delete project if I have permission

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
      | id  | name            | description               | imageTag       | env    | testImageTag      | testEnv | resultParser | createdAt           | updatedAt           |
      | 11  | Test-Project-1  | Test-Project-1 Description| nginx:latest   | a=b    | test-nginx:latest |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 12  | Test-Project-2  | Test-Project-2 Description| mysql:8.0.26   | c=d    | test-mysql:8.0.26 |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | 13  | Test-Project-3  | Test-Project-3 Description| nginx:latest   |        | test-nginx:test-2 |         |  | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
  Scenario: Successful delete project with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "DELETE" request to "/api/v1/projects/13":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true
      }
      """
  
  Scenario: Successful delete project with "AdminAndContestant" user
    Given I am logged in as "user04"
    When I send a "DELETE" request to "/api/v1/projects/11":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true
      }
      """

  Scenario: Unauthorized delete project
    When I send a "DELETE" request to "/api/v1/projects/11":
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """
    
  Scenario: Forbidden delete project with "APIReader" user
    Given I am logged in as "user02"
    When I send a "DELETE" request to "/api/v1/projects/14":
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: Forbidden delete project with "Contestant" user
    Given I am logged in as "user03"
    When I send a "DELETE" request to "/api/v1/projects/15":
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: Delete non-existent project
    Given I am logged in as "user01"
    When I send a "DELETE" request to "/api/v1/projects/999":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "project not found",
        "details": null
      }
      """
