Feature: Create Image API
  As a user
  I want to create an image
  So that I can create an image and then build it if I have permission

  Background:
    Given I have users:
      | id  | name               | username | password | roles              |
      | 11  | AdminUser          | user01   | 123456   | admin              |
      | 12  | APIReader          | user02   | 123456   | api_reader         |
      | 13  | Contestant         | user03   | 123456   | contestant         |
      | 14  | AdminAndContestant | user04   | 123456   | admin,contestant   |
    And I have images:
      | id | tag          | mount_path    | description  | isReady | createdAt           | updatedAt           |
      | 1  | nginx:latest | /var/www/html | description1 | 1       | 2021-09-01T00:00:00 | 2021-09-01T00:00:00 |

  Scenario: Successful create image with "AdminUser" user
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/images" with the following formdata:
      | key              | value             |
      | tag              | test-nginx:test-2 |
      | mountPath        | /var/www/html     |
      | description      | description1      |
      | project          | {nginx project}   |
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": "test-nginx:test-2"
      }
      """
    And when start streaming logs via request to "/api/v1/images/test-nginx:test-2/builds/1/stream-logs":
    And wait for 30 seconds for the image "test-nginx:test-2" build to complete
    And streaming logs should have "Successfully built" message within 30 seconds
  
  Scenario: Successful create image with build error image
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/images" with the following formdata:
      | key              | value                  |
      | tag              | test-nginx:test-3      |
      | mountPath        | /var/www/html          |
      | description      | description1           |
      | project          | {nginx-fail project}   |
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": "test-nginx:test-3"
      }
      """
    And when start streaming logs via request to "/api/v1/images/test-nginx:test-3/builds/1/stream-logs":
    And wait for 30 seconds for the image "test-nginx:test-3" build to error
    And streaming logs should have "returned a non-zero code" message within 30 seconds

  Scenario: Failed create image with existing image
    Given I am logged in as "user01"
    When I send a "POST" request to "/api/v1/images" with the following formdata:
      | key              | value             |
      | tag              | nginx:latest      |
      | mountPath        | /var/www/html     |
      | description      | description1      |
      | project          | {nginx project}   |
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "image already exists",
        "details": null
      }
      """

  Scenario: Failed create image with "APIReader" user
    Given I am logged in as "user02"
    When I send a "POST" request to "/api/v1/images" with the following formdata:
      | key              | value             |
      | tag              | test-nginx:test-3 |
      | mountPath        | /var/www/html     |
      | description      | description1      |
      | project          | {nginx project}   |
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: Failed create image with non-login user
    When I send a "POST" request to "/api/v1/images" with the following formdata:
      | key              | value             |
      | tag              | test-nginx:test-4 |
      | mountPath        | /var/www/html     |
      | description      | description1      |
      | project          | {nginx project}   |
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """

  Scenario: Reject a tag using the reserved rename-on-delete namespace
    Given I am logged in as "user01"
    # "__deleted_" is how a soft-deleted image's tag is renamed to free the
    # unique index; a live image squatting on that namespace could collide
    # with a dead row's renamed tag (blocking its delete, or panicking the
    # rename migration at startup), so creation rejects it outright.
    When I send a "POST" request to "/api/v1/images" with the following formdata:
      | key              | value                    |
      | tag              | evil__deleted_1:latest   |
      | mountPath        | /var/www/html            |
      | description      | squatting                |
      | project          | {nginx project}          |
    Then the response code should be 400
    And the response should match json:
      """
      {
        "success": false,
        "message": "invalid tag",
        "details": null
      }
      """
