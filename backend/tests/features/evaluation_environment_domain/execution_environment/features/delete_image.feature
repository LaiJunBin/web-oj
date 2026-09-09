Feature: Delete Image API
  As a user
  I want to delete an image by tag
  So that I can remove an image from the system if I have the permission

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

  Scenario: Successful delete image with "AdminUser" user for image "nginx:latest"
    Given I am logged in as "user01"
    When I send a "DELETE" request to "/api/v1/images/nginx:latest":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true
      }
      """
    And the image "nginx:latest" should not exist

  Scenario: Successful delete image with "AdminAndContestant" user for image "test-nginx:latest"
    Given I am logged in as "user04"
    When I send a "DELETE" request to "/api/v1/images/test-nginx:latest":
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true
      }
      """
    And the image "test-nginx:latest" should not exist

  Scenario: Unauthorized delete image with unauthenticated user for image "nginx:latest"
    When I send a "DELETE" request to "/api/v1/images/nginx:latest":
    Then the response code should be 401
    And the response should match json:
      """
      {
        "success": false,
        "message": "Unauthorized",
        "details": null
      }
      """

  Scenario: Forbidden delete image with "APIReader" user for image "nginx:latest"
    Given I am logged in as "user02"
    When I send a "DELETE" request to "/api/v1/images/nginx:latest":
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: Forbidden delete image with "Contestant" user for image "test-nginx:latest"
    Given I am logged in as "user03"
    When I send a "DELETE" request to "/api/v1/images/test-nginx:latest":
    Then the response code should be 403
    And the response should match json:
      """
      {
        "success": false,
        "message": "Forbidden",
        "details": null
      }
      """

  Scenario: Not found delete image with "AdminUser" user for image "not-found-image"
    And I am logged in as "user01"
    When I send a "GET" request to "/api/v1/images/not-found-image":
    Then the response code should be 404
    And the response should match json:
      """
      {
        "success": false,
        "message": "image not found",
        "details": null
      }
      """


  Scenario: Deleting an image frees its tag for re-creation
    Given I am logged in as "user01"
    When I send a "DELETE" request to "/api/v1/images/test-nginx:latest":
    Then the response code should be 200
    # Before the rename-on-delete fix this POST fails on MySQL: the soft-deleted
    # row still occupies the (non-partial) unique tag index, so the INSERT is
    # rejected — sqlite honours the partial index, which is why only BDD
    # (running on real MySQL) can prove this path.
    When I send a "POST" request to "/api/v1/images" with the following formdata:
      | key              | value             |
      | tag              | test-nginx:latest |
      | mountPath        | /var/www/html     |
      | description      | recreated         |
      | project          | {nginx project}   |
    Then the response code should be 200
    And the response should match json:
      """
      {
        "success": true,
        "data": "test-nginx:latest"
      }
      """
    # Delete again: contains the queued background build inside this scenario
    # (cancelBuild now runs while the row is still live) and proves the SECOND
    # soft-delete of the same base tag gets a distinct suffix on the real
    # MySQL unique index.
    When I send a "DELETE" request to "/api/v1/images/test-nginx:latest":
    Then the response code should be 200
