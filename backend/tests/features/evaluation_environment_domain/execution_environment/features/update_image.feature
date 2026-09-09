Feature: Update Image API
  As a user
  I want to update an image by tag
  So that I can update the image details if I have the permission

  Background:
    Given I have users:
      | id  | name               | username | password | roles              |
      | 11  | AdminUser          | user01   | 123456   | admin              |
      | 12  | APIReader          | user02   | 123456   | api_reader         |
      | 13  | Contestant         | user03   | 123456   | contestant         |
      | 14  | AdminAndContestant | user04   | 123456   | admin,contestant   |
    And I have images:
      | id | tag                | mount_path    | description  | is_ready | createdAt           | updatedAt           |
      | 1  | nginx:latest       | /var/www/html | description1 | 1        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 2  | mysql:8.0.26       | /var/www/html | description2 | 0        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 3  | test-nginx:latest  | /var/www/html | description3 | 0        | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
    And I have builds:
      | ID | Image             | Batch | Status    | CreatedAt           | UpdatedAt           |
      | 1  | nginx:latest      | 1     | CANCELLED | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 2  | nginx:latest      | 2     | SUCCESS   | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 3  | nginx:latest      | 3     | BUILDING  | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 4  | mysql:8.0.26      | 1     | ERROR     | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 5  | mysql:8.0.26      | 2     | PENDING   | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |
      | 6  | test-nginx:latest | 1     | ERROR     | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |

  Scenario: Successful update image description with "AdminUser" user for image "nginx:latest"
    Given I am logged in as "user01"
    When I send a "PATCH" request to "/api/v1/images/nginx:latest" with the following formdata:
      | key              | value                     |
      | description      | description1-updated      |
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true
      }
      """

  Scenario: Successful update image mount path with "AdminUser" user for image "nginx:latest"
    Given I am logged in as "user01"
    When I send a "PATCH" request to "/api/v1/images/nginx:latest" with the following formdata:
      | key              | value                     |
      | mount_path       | /usr/share/nginx/html     |
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true
      }
      """

  Scenario: Failed update image project with building image "nginx:latest"
    Given I am logged in as "user01"
    When I send a "PATCH" request to "/api/v1/images/nginx:latest" with the following formdata:
      | key              | value                     |
      | project          | {nginx project}           |
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "current build in progress, please cancel the build first",
        "details": null
      }
      """

  Scenario: Unauthorized update image with unauthenticated user for image "nginx:latest"
    When I send a "PATCH" request to "/api/v1/images/nginx:latest" with the following formdata:
      | key              | value                     |
      | description      | description1-updated      |
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """

  Scenario: Forbidden update image with "APIReader" user for image "nginx:latest"
    Given I am logged in as "user02"
    When I send a "PATCH" request to "/api/v1/images/nginx:latest" with the following formdata:
      | key              | value                     |
      | description      | description1-updated      |
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: Forbidden update image with "Contestant" user for image "nginx:latest"
    Given I am logged in as "user03"
    When I send a "PATCH" request to "/api/v1/images/nginx:latest" with the following formdata:
      | key              | value                     |
      | description      | description1-updated      |
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: Not found update image with "AdminUser" user for image "not-found-image:latest"
    And I am logged in as "user01"
    When I send a "PATCH" request to "/api/v1/images/not-found-image:latest" with the following formdata:
      | key              | value                     |
      | description      | description1-updated      |
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "image not found",
        "details": null
      }
      """

