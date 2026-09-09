Feature: Update Asset API
  As a user
  I want to update an asset
  So that I can update an asset if I have permission

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
    And I have assets: 
      | id       | projectId | name     | description | size | isEnabled | createdAt           | updatedAt           |
      | uuid-1   | 11        | test.zip | description | 100  | 0         | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | uuid-2   | 12        | test2.zip| description | 100  | 1         | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
      | uuid-3   | 11        | test3.zip| description | 100  | 1         | 2021-09-01 00:00:00 | 2021-09-01 00:00:00 |
    
  Scenario: Successful update asset with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "PATCH" request to "/api/v1/assets/uuid-1" with the following formdata:
      | key              | value             |
      | name             | test-updated.zip  |
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true
      }
      """
  
  Scenario: Successful update asset with "AdminAndContestant" user
    Given I am logged in as "user04"
    When I send a "PATCH" request to "/api/v1/assets/uuid-3" with the following formdata:
      | key              | value             |
      | name             | test3-updated.zip |
      | file             | {nginx project}   |
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true
      }
      """
    And I can download the asset "test3-updated.zip" with "GET" request to "/api/v1/assets/{assetId}/download"

  Scenario: Failed update asset with name is empty
    Given I am logged in as "user01"
    When I send a "PATCH" request to "/api/v1/assets/uuid-1" with the following formdata:
      | key              | value             |
      | name             |                   |
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "invalid file name",
        "details": null
      }
      """
  
  Scenario: Unauthorized update asset
    When I send a "PATCH" request to "/api/v1/assets/uuid-1" with the following formdata:
      | key              | value             |
      | name             | test.zip          |
      | file             | {nginx project}   |
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """
    
  Scenario: Forbidden update asset with "APIReader" user
    Given I am logged in as "user02"
    When I send a "PATCH" request to "/api/v1/assets/uuid-1" with the following formdata:
      | key              | value             |
      | name             | test.zip          |
      | file             | {nginx project}   |
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: Forbidden update asset with "Contestant" user
    Given I am logged in as "user03"
    When I send a "PATCH" request to "/api/v1/assets/uuid-1" with the following formdata:
      | key              | value             |
      | name             | test.zip          |
      | file             | {nginx project}   |
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: Failed update asset with asset not found
    Given I am logged in as "user01"
    When I send a "PATCH" request to "/api/v1/assets/uuid-4" with the following formdata:
      | key              | value             |
      | name             | test.zip          |
      | file             | {nginx project}   |
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "asset not found",
        "details": null
      }
      """
